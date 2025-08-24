package dream

import (
	"dreamproxy/config"
	"dreamproxy/fs"
	"dreamproxy/http"
	"dreamproxy/logger"
	"dreamproxy/mime"
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
		res, err := HandleRequest(req, server_configs)

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

func HandleRequest(req *http.HttpReq, server_configs []config.Server) (*http.HttpRes, error) {
	var res *http.HttpRes
	target := req.Target

	// Prepare Response
	res = &http.HttpRes{
		Version: http.V1_1,
		Headers: make(map[string]string),
	}

	res.Headers["connection"] = req.Headers["connection"]

	host := req.Headers["host"]
	method := req.Method
	scheme := req.Scheme
	target_url, err := url.Parse(scheme + "://" + host + target)

	if err != nil {
		log.Println("Invalid URL:", err)
		res.Status = http.StatusBadRequest
		return nil, err
	}

	target_url.Path = path.Clean(target_url.Path)

	// Handle Configs
	for _, server_cfg := range server_configs {

		if host == server_cfg.Name || slices.Contains(server_cfg.Hosts, host) {

			// Check if port is part of host
			if strings.Contains(host, ":") {
				host = strings.SplitN(host, ":", 2)[0]
			}

			// Check if Proxy
			for _, location := range server_cfg.Locations {

				// Does not support globbing yet
				if !strings.HasPrefix(target_url.Path, path.Clean(location.Path)) {
					continue
				}

				if location.ProxyPass != "" {
					origin_server := location.ProxyPass
					origin_host := ""
					origin_port_str := ""

					if strings.Contains(origin_server, "://") {
						scheme_host := strings.SplitN(origin_server, "://", 2)

						origin_host = scheme_host[1]

						if strings.Contains(origin_host, ":") {
							origin_host_port := strings.SplitN(origin_host, ":", 2)
							origin_host = origin_host_port[0]
							origin_port_str = origin_host_port[1]
						}
					}

					origin_port, err := strconv.Atoi(origin_port_str)

					if err != nil {
						continue
					}

					res, err = http.MakeRequest(req.Method, origin_host, origin_port, target_url.Path, http.RequestConfig{
						Headers: req.Headers,
						Body:    req.Body,
					})

					if err != nil {
						return nil, err
					}

					if res.Status == http.StatusMovedPermanently || res.Status == http.StatusFound {
						location := res.Headers["location"]

						res, err = http.MakeRequest(req.Method, origin_host, origin_port, location, http.RequestConfig{
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
						handleHead(target_url.Path, res, location.Root)
						break
					case "GET":
						handleGet(target_url.Path, res, location.Root)
						break
					default:
						return nil, err
					}
				}

			}

			return res, nil
		}

		continue
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

func handleGet(target_url string, res *http.HttpRes, root_fs string) error {
	var res_body []byte
	var err error

	file_path, _, err := fs.ResolveFilePath(target_url, root_fs)

	ext := filepath.Ext(file_path)

	content_type := mime.MimeTypes[ext]

	if content_type == "" {
		content_type = "application/octet-stream"
	}

	res_body, err = fs.LoadFile(file_path)
	if err != nil {
		not_found_page, err := fs.LoadFile(path.Join(root_fs, "not_found.html"))

		if err != nil {
			not_found_page = []byte("<h1>404 Not Found</h1>")
		}

		res.Status = http.StatusNotFound

		res.Headers["content-length"] = fmt.Sprint(len(not_found_page))
		res.Headers["connection"] = "close"

		res.Body = []byte(not_found_page)

	} else {
		res.Status = http.StatusOK

		res.Headers["content-length"] = fmt.Sprint(len(res_body))

		res.Body = res_body
	}

	return err
}
