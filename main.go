package main

import (
	"bytes"
	"dreamproxy/file_system"
	http_common "dreamproxy/http/common"
	"dreamproxy/http/parser"
	"dreamproxy/mime"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
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
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	ln, err := net.Listen(PROTOCOL, fmt.Sprintf(":%s", PORT))

	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()

	log.Printf("%s", fmt.Sprintf("listening on :%s", PORT))

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(c net.Conn) {
	// Parse HTTP header here to know whether to keep connection alive
	defer c.Close()

	for {
		req_raw, err := extractRequest(c)

		if err != nil {
			log.Println(err)
			return
		}

		req, err := http_parser.ParseRawHttp(req_raw.String())

		if err != nil {
			log.Println(err)
			return
		}

		res, err := handleRequest(req)

		if err != nil {
			log.Println(err)
			return
		}

		c.Write([]byte(res.ToStr()))

		// Check keep-alive
		if strings.ToLower(res.Connection) != "keep-alive" {
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
		res.ContentType = mime.MimeTypes[".html"]
		file_path = path.Join(ROOT_FS, target_url.Path)

		// Is Root
		if target_url.Path == "/" {
			file_path = path.Join(ROOT_FS, "index.html")
		}

	} else { // Resource URLs
		res.ContentType = mime.MimeTypes[ext]
		if res.ContentType == "" {
			res.ContentType = "application/octet-stream"
		}
		file_path = path.Join(ROOT_FS, target_url.Path)
	}

	stat, err := os.Stat(file_path)

	if err != nil {
		log.Println(err)
		res.Status = http_common.StatusNotFound
		res.ContentLength = 0
	} else {
		res.Status = http_common.StatusOK
		res.ContentLength = int(stat.Size())
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
		res.ContentType = mime.MimeTypes[".html"]
		file_path := path.Join(ROOT_FS, target_url.Path)

		// Is Root
		if target_url.Path == "/" {
			file_path = path.Join(ROOT_FS, "index.html")
		}

		res_body, err = file_system.LoadFile(file_path)

	} else { // Resource URLs
		res.ContentType = mime.MimeTypes[ext]
		if res.ContentType == "" {
			res.ContentType = "application/octet-stream"
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
		res.Body = []byte(not_found_page)
		res.ContentLength = len(not_found_page)
	} else {
		res.Status = http_common.StatusOK
		res.Body = res_body
		res.ContentLength = len(res.Body)
	}

	return err
}

func handleRequest(req http_common.HttpReq) (http_common.HttpRes, error) {
	var res http_common.HttpRes
	target := req.Target

	// Prepare Response
	req_connection := req.Headers["Connection"]
	res = http_common.HttpRes{
		Version:    http_common.V1_1,
		Connection: req_connection,
	}

	host := req.Headers["Host"]
	method := req.Method
	scheme := req.Scheme
	target_url, err := url.Parse(scheme + "://" + host + target)
	target_url.Path = path.Clean(target_url.Path)

	if err != nil {
		log.Println(err)
	}

	switch method {
	case "HEAD":
		handleHead(target_url, &res)
		break
	case "GET":
		handleGet(target_url, &res)
		break
	default:
		log.Println("Invalid Method")
		break
	}

	return res, nil
}

func extractRequest(c net.Conn) (bytes.Buffer, error) {
	// Implement proper HTTP reading (read till /r/n/r/n)
	tmp_buf := make([]byte, 1024)
	var req_raw bytes.Buffer

	for {
		var n, err = c.Read(tmp_buf)

		if err != nil {
			if errors.Is(err, io.EOF) {
				return req_raw, err
			} else if errors.Is(err, net.ErrClosed) {
				log.Println("Client disconnected:", c.RemoteAddr())
				return req_raw, err
			}
		}

		if n <= 0 {
			log.Println("No data transmitted")
			return req_raw, err
		}

		req_raw.Write(tmp_buf[:n])

		is_header_end := strings.Contains(req_raw.String(), "\r\n\r\n")

		if !is_header_end {
			continue
		} else {
			break
		}
	}

	return req_raw, nil
}
