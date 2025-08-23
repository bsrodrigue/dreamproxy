package main

import (
	"dreamproxy/file_system"
	http_client "dreamproxy/http/client"
	http_common "dreamproxy/http/common"
	"dreamproxy/http/parser"
	"dreamproxy/logger"
	"dreamproxy/mime"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

const PROTOCOL string = "tcp4"
const PORT string = "8080"
const ROOT_FS string = "staticfiles"
const LOG_FILE string = "/var/log/dreamserver/access.log"

func main() {
	ln, err := net.Listen(PROTOCOL, fmt.Sprintf(":%s", PORT))

	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()

	log.Printf("%s", fmt.Sprintf("listening on :%s", PORT))

	// Server Loop
	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)
	}
}

func NewFailedToParseRes(remoteAddr string, msg string) *http_common.HttpRes {
	res := http_common.CreateHttpRes()
	res.Status = http_common.StatusBadRequest

	log := logger.NewRequestLog(logger.DREAM_SERVER, logger.ERROR, logger.REQ_PARSE_ERROR, msg)
	log.Request.ClientIP = remoteAddr

	// Create a log handler
	fmt.Println(log.ToText())
	return res
}

func NewBadRequestRes(req http_common.HttpReq, remoteAddr string, err error) *http_common.HttpRes {
	res := http_common.CreateHttpRes()
	res.Status = http_common.StatusBadRequest

	log := logger.NewRequestLog(logger.DREAM_SERVER, logger.ERROR, logger.BAD_REQUEST, err.Error())
	log.Request.ID = uuid.New().String()
	log.Request.Method = req.Method
	log.Request.Path = req.Target
	log.Request.ClientIP = remoteAddr
	log.Response.StatusCode = int(res.Status)
	log.Response.BytesSent = 0

	// Create a log handler
	fmt.Println(log.ToText())
	return res
}

// Create a ClientSession struct to handle each connection in a cleaner way
func handleConn(connection net.Conn) {
	defer connection.Close()

	for {
		req_start := time.Now()
		req_raw, err := http_parser.ReadFullHttpMessage(connection)

		if err != nil {
			res := NewFailedToParseRes(connection.RemoteAddr().String(), err.Error())
			res.Version = http_common.V1_1
			res.SetServerHeaders()
			connection.Write([]byte(res.ToStr()))
			return
		}

		req, err := http_parser.ParseRawHttpReq(req_raw)

		if err != nil {
			res := NewFailedToParseRes(connection.RemoteAddr().String(), err.Error())
			res.Version = http_common.V1_1
			res.SetServerHeaders()
			connection.Write([]byte(res.ToStr()))
			return
		}

		//------------- Request has been successfully parsed by now

		req.Headers["x-forwarded-for"] = connection.RemoteAddr().String()
		res, err := handleRequest(req)

		if err != nil {
			res := NewBadRequestRes(*req, connection.RemoteAddr().String(), err)
			connection.Write([]byte(res.ToStr()))
			return
		}

		// Add final headers
		res.Version = http_common.V1_1
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
		log.Response.BytesSent = int64(len(res_bytes))
		log.Response.LatencyMS = latency.Milliseconds()
		log.Response.StatusCode = int(res.Status)

		// Create a log handler
		fmt.Println(log.ToText())

		// Check keep-alive
		if strings.ToLower(res.Headers["connection"]) != "keep-alive" {
			return
		}
	}
}

func handleHead(target_url *url.URL, res *http_common.HttpRes) error {
	// Write a static file finder (resolver)
	var err error
	var file_path string
	ext := path.Ext(target_url.Path)
	ext = strings.ToLower(ext)

	// Page URLs
	if ext == "" {
		res.Headers["content-type"] = mime.MimeTypes[".html"]
		file_path = path.Join(ROOT_FS, target_url.Path)

		// Is Root
		if target_url.Path == "/" {
			file_path = path.Join(ROOT_FS, "index.html")
		}

	} else { // Resource URLs
		res.Headers["content-type"] = mime.MimeTypes[ext]
		if res.Headers["content-type"] == "" {
			res.Headers["content-type"] = "application/octet-stream"
		}
		file_path = path.Join(ROOT_FS, target_url.Path)
	}

	stat, err := os.Stat(file_path)

	if err != nil {
		log.Println(err)
		res.Status = http_common.StatusNotFound
		res.Headers["content-length"] = "0"
	} else {
		res.Status = http_common.StatusOK
		res.Headers["content-length"] = fmt.Sprint(stat.Size())
	}

	return err
}

func handleGet(target_url *url.URL, res *http_common.HttpRes) error {
	var res_body []byte
	var err error
	ext := path.Ext(target_url.Path)
	ext = strings.ToLower(ext)

	// Page URLs
	if ext == "" {
		res.Headers["Content-Type"] = mime.MimeTypes[".html"]
		file_path := path.Join(ROOT_FS, target_url.Path)

		// Is Root
		if target_url.Path == "/" {
			file_path = path.Join(ROOT_FS, "index.html")
		}

		res_body, err = file_system.LoadFile(file_path)

	} else { // Resource URLs
		res.Headers["Content-Type"] = mime.MimeTypes[ext]

		if res.Headers["content-type"] == "" {
			res.Headers["content-type"] = "application/octet-stream"
		}
		file_path := path.Join(ROOT_FS, target_url.Path)
		res_body, err = file_system.LoadFile(file_path)
	}

	if err != nil {
		not_found_page, err := file_system.LoadFile(path.Join(ROOT_FS, "not_found.html"))

		if err != nil {
			not_found_page = []byte("<h1>404 Not Found</h1>")
		}

		res.Status = http_common.StatusNotFound

		res.Headers["content-length"] = fmt.Sprint(len(not_found_page))
		res.Headers["connection"] = "close"

		res.Body = []byte(not_found_page)

	} else {
		res.Status = http_common.StatusOK

		res.Headers["content-length"] = fmt.Sprint(len(res_body))

		res.Body = res_body
	}

	return err
}

func handleRequest(req *http_common.HttpReq) (*http_common.HttpRes, error) {
	var res *http_common.HttpRes
	target := req.Target

	// Prepare Response
	res = &http_common.HttpRes{
		Version: http_common.V1_1,
		Headers: make(map[string]string),
	}

	res.Headers["connection"] = req.Headers["connection"]

	host := req.Headers["host"]
	method := req.Method
	scheme := req.Scheme
	target_url, err := url.Parse(scheme + "://" + host + target)

	if err != nil {
		log.Println("Invalid URL:", err)
		res.Status = http_common.StatusBadRequest
		return nil, err
	}

	target_url.Path = path.Clean(target_url.Path)

	// Check if Proxy
	if host == "djangoserver.com:8080" {

		res, err = http_client.MakeRequest(req.Method, "djangoserver.com", 8000, target_url.Path, http_client.RequestConfig{
			Headers: req.Headers,
			Body:    req.Body,
		})

		if err != nil {
			return nil, err
		}

		if res.Status == http_common.StatusMovedPermanently || res.Status == http_common.StatusFound {
			location := res.Headers["location"]

			res, err = http_client.MakeRequest(req.Method, "djangoserver.com", 8000, location, http_client.RequestConfig{
				Headers: req.Headers,
				Body:    req.Body,
			})

			if err != nil {
				return nil, err
			}

		}

		return res, nil
	} else {
		switch method {
		case "HEAD":
			handleHead(target_url, res)
			break
		case "GET":
			handleGet(target_url, res)
			break
		default:
			return nil, err
		}
	}

	return res, nil
}
