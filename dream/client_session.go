package dream

import (
	"dreamproxy/config"
	"dreamproxy/fs"
	"dreamproxy/http"
	"dreamproxy/logger"
	"dreamproxy/mime"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ClientSession struct {
	RemoteAddress string
	RemotePort    string
	Connection    net.Conn

	// Tunneling Fields
	IsTunneling bool
	TunnelHost  string
	TunnelPort  string
}

func NewClientSession(connection net.Conn) ClientSession {
	remote_addr := connection.RemoteAddr().String()
	remote_port := ""

	if strings.Contains(remote_addr, ":") {
		split := strings.Split(remote_addr, ":")

		remote_addr = split[0]
		remote_port = split[1]
	}

	return ClientSession{
		RemoteAddress: remote_addr,
		RemotePort:    remote_port,
		Connection:    connection,
	}
}

func (session *ClientSession) HandleConnection(server_configs []config.Server) {
	defer session.Connection.Close()

	connection := session.Connection

	for {
		req_start := time.Now()
		req_raw, err := http.ReadFullHttpMessage(connection)

		if err != nil {
			res := http.NewFailedToParseRes(connection.RemoteAddr().String(), err.Error())
			res.Version = http.V1_1
			res.SetServerHeaders()
			connection.Write([]byte(res.ToStr()))
			return
		}

		req, err := http.ParseRawHttpReq(req_raw)

		if err != nil {
			res := http.NewFailedToParseRes(connection.RemoteAddr().String(), err.Error())
			res.Version = http.V1_1
			res.SetServerHeaders()
			connection.Write([]byte(res.ToStr()))
			return
		}

		//------------- Request has been successfully parsed by now

		req.Headers["x-forwarded-for"] = connection.RemoteAddr().String()
		res, err := session.HandleRequest(req, server_configs)

		if err != nil {
			res := http.NewBadRequestRes(*req, connection.RemoteAddr().String(), err)
			connection.Write([]byte(res.ToStr()))
			return
		}

		// Add final headers
		res.Version = http.V1_1
		res.SetServerHeaders()

		res_bytes := res.ToBytes()

		latency := time.Since(req_start)
		connection.Write(res_bytes)

		log := logger.NewRequestLog(logger.DREAM_SERVER, logger.INFO, logger.REQUEST, "")
		log.Request.ID = uuid.New().String()
		log.Request.Method = req.Method
		log.Request.Path = req.Target
		log.Request.Host = req.Headers["host"]
		log.Request.ClientIP = connection.RemoteAddr().String()
		log.Response.StatusCode = int(res.Status)
		log.Response.BytesSent = int64(len(res.Body))
		log.Response.LatencyMS = latency.Milliseconds()
		log.Response.StatusCode = int(res.Status)

		// Create a log handler
		fmt.Println(log.ToText())

		// Check keep-alive
		if strings.ToLower(res.Headers["connection"]) == "close" {
			return
		}
	}
}

func (session *ClientSession) HandleRequest(req *http.HttpReq, server_configs []config.Server) (*http.HttpRes, error) {
	var res *http.HttpRes
	var req_url *url.URL
	var err error

	target := req.Target
	host := req.Headers["host"]
	scheme := req.Scheme
	method := req.Method

	// Prepare Response
	res = &http.HttpRes{
		Version: http.V1_1,
		Headers: make(map[string]string),
	}

	res.Headers["connection"] = req.Headers["connection"]

	// res.Headers["content-length"] = "0" // By default
	// res.Status = http.StatusNotFound    // By default

	// Handle OPTIONS * (Asterisk Form)
	if http.AsteriskForm.MatchString(target) {
		if req.Method != "OPTIONS" {
			// Bad Request
			return nil, errors.New("* can only be used with OPTIONS method")
		}
		res.Status = http.StatusOK
		res.Headers["allow"] = "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS" // We don't support CONNECT yet, maybe we will never...

		// Maybe it is pointless to set this header when there is no content
		// res.Headers["content-length"] = "0"

		return res, nil
	}

	// Handle CONNECT (Tunneling)
	if http.AuthorityForm.MatchString(target) {
		if req.Method != "CONNECT" {
			// Bad Request
			return nil, errors.New("Authority form currently only works with CONNECT method")
		}

		// Open a tunnel (Tunnels are blind and don't process nor peek in the data)
		// Implement it when you want to support forward proxying.

		return nil, errors.New("Not supported yet :'|")
	}

	// Handle Absolute Form
	if http.AbsoluteForm.MatchString(target) {
		req_url, err = url.Parse(target)
	}

	// Handle Origin Form
	if http.OriginForm.MatchString(target) {
		req_url, err = url.Parse(scheme + "://" + host + target)
	}

	if err != nil {
		log.Println("Invalid URL:", err)
		res.Status = http.StatusBadRequest
		return nil, err
	}

	req_url.Path = path.Clean(req_url.Path)

	// Handle Configs
	for _, server_cfg := range server_configs {

		if host != server_cfg.Name && !slices.Contains(server_cfg.Hosts, host) {
			// No configuration matches this request target.
			// Maybe you should learn about pattern matching for more precise matching.
			continue
		}

		// Check if port is part of host
		if strings.Contains(host, ":") {
			host = strings.SplitN(host, ":", 2)[0]
		}

		for _, location := range server_cfg.Locations {

			// Does not support globbing yet
			if !strings.HasPrefix(req_url.Path, path.Clean(location.Path)) {
				continue
			}

			// Check if Proxy
			if location.ProxyPass != "" {
				res, err = http.MakeRequest(req.Method, location.OriginHost, location.OriginPortInt, req_url.Path, http.RequestConfig{
					Headers: req.Headers,
					Body:    req.Body,
				})

				if err != nil {
					return nil, err
				}

				if res.Status == http.StatusMovedPermanently || res.Status == http.StatusFound {
					new_url_path := res.Headers["location"]

					res, err = http.MakeRequest(req.Method, location.OriginHost, location.OriginPortInt, new_url_path, http.RequestConfig{
						Headers: req.Headers,
						Body:    req.Body,
					})

					if err != nil {
						return nil, err
					}
				}
			} else {

				// Static File Server
				switch method {
				case "HEAD":
					handleHead(req_url.Path, res, location.Root)
					break
				case "GET":
					handleGet(req_url.Path, *req, res, location.Root)
					break
				default:
					// Method not allowed
					return nil, err
				}
			}

		}

		return res, nil
	}

	return res, nil
}

func handleHead(target_url string, res *http.HttpRes, root_fs string) error {
	file_path, stat, err := fs.ResolveFilePath(target_url, root_fs)

	ext := filepath.Ext(file_path)

	content_type := mime.MimeTypes[ext]

	if content_type == "" {
		content_type = "application/octet-stream"
	}

	if err != nil {
		log.Println(err)
		res.Status = http.StatusNotFound
		res.Headers["content-length"] = "0"
	} else {
		res.Status = http.StatusOK
		res.Headers["content-type"] = content_type
		res.Headers["content-length"] = fmt.Sprint(stat.Size())
	}

	return err
}

func handleGet(target_url string, req http.HttpReq, res *http.HttpRes, root_fs string) error {
	var res_body []byte
	var err error

	file_path, stat, err := fs.ResolveFilePath(target_url, root_fs)

	ext := filepath.Ext(file_path)

	content_type := mime.MimeTypes[ext]

	if content_type == "" {
		content_type = "application/octet-stream"
	}

	if err != nil {
		res.Status = http.StatusNotFound
		res.Body = make([]byte, 0)

	} else {
		if_none_match := req.Headers["if-none-match"]

		// Generate ETag
		file_size := stat.Size()
		last_modified := stat.ModTime().Unix()
		etag := strconv.Itoa(int(file_size) + int(last_modified))

		res.Headers["etag"] = etag
		res.Headers["last-modified"] = stat.ModTime().Format(time.RFC1123)

		// Handle Server Side caching
		// res.Headers["expires"] =

		if if_none_match == etag {
			res.Status = http.StatusNotModified
			res_body = make([]byte, 0)
		} else {
			// res_body, err = fs.LoadFile(file_path)
			res_body, err = fs.GlobalStaticFileCache.Get(file_path)

			if err != nil {
				return err
			}

			res.Status = http.StatusOK
			res.Headers["content-length"] = fmt.Sprint(len(res_body))
			res.Body = res_body
		}

	}

	return err
}
