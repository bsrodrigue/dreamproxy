package main

import (
	"dreamproxy/file_system"
	http_client "dreamproxy/http/client"
	http_common "dreamproxy/http/common"
	"dreamproxy/http/parser"
	"dreamproxy/mime"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"net/url"
	"os"
	"path"
	"strings"
)

const PROTOCOL string = "tcp4"
const PORT string = "8080"
const ROOT_FS string = "staticfiles"

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

func CreateBadRequestRes() *http_common.HttpRes {
	res := http_common.CreateHttpRes()
	res.Version = http_common.V1_1
	res.Status = http_common.StatusBadRequest

	res.SetServerHeaders()

	return res
}

// Create a ClientSession struct to handle each connection in a cleaner way
func handleConn(c net.Conn) {
	defer c.Close()

	for {
		req_raw, err := http_parser.ReadFullHttpMessage(c)

		if err != nil {
			log.Println(err)
			res := CreateBadRequestRes()
			c.Write([]byte(res.ToStr()))
			return
		}

		req, err := http_parser.ParseRawHttpReq(req_raw)

		if err != nil {
			log.Println(err)
			res := CreateBadRequestRes()
			c.Write([]byte(res.ToStr()))
			return
		}

		req.Headers["x-forwarded-for"] = c.RemoteAddr().String()
		res, err := handleRequest(req)

		if err != nil {
			log.Println(err)
			res := CreateBadRequestRes()
			c.Write([]byte(res.ToStr()))
			return
		}

		res.SetServerHeaders() // Add final headers
		c.Write([]byte(res.ToStr()))

		// Check keep-alive
		if strings.ToLower(res.Headers["connection"]) != "keep-alive" {
			return
		}
	}
}

func handleHead(target_url *url.URL, res *http_common.HttpRes) error {
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
		log.Println(err)
		not_found_page, err := file_system.LoadFile(path.Join(ROOT_FS, "not_found.html"))

		if err != nil {
			log.Println(err)
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
		return CreateBadRequestRes(), err
	}

	target_url.Path = path.Clean(target_url.Path)

	// Check if Proxy
	if host == "djangoserver.com:8080" {
		res, err = http_client.Get("djangoserver.com", 8000, target_url.Path, http_client.RequestConfig{
			Headers: req.Headers,
			Body:    req.Body,
		})

		if err != nil {
			return CreateBadRequestRes(), err
		}

		if res.Status == http_common.StatusMovedPermanently {
			location := res.Headers["location"]

			res, err = http_client.Get("djangoserver.com", 8000, location, http_client.RequestConfig{
				Headers: req.Headers,
				Body:    req.Body,
			})

			if err != nil {
				return CreateBadRequestRes(), err
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
			return CreateBadRequestRes(), err
		}
	}

	return res, nil
}
