package main

import (
	"dreamproxy/file_system"
	http_common "dreamproxy/http/common"
	"dreamproxy/http/parser"
	"dreamproxy/mime"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"path"
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
		req_raw, err := exctractRequest(c)

		if err != nil {
			log.Println(err)
			return
		}

		req, err := http_parser.ParseRawHttp(req_raw)

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
	}
}

func handleRequest(req http_common.HttpReq) (http_common.HttpRes, error) {
	var res http_common.HttpRes
	var res_body = []byte("")
	var file_path = string("")
	target := req.Target

	// Prepare Response
	req_connection := req.Headers["Connection"]
	res = http_common.HttpRes{
		Version:    http_common.V0_9,
		Connection: req_connection,
	}

	host := req.Headers["Host"]
	scheme := req.Scheme
	target_path, err := url.Parse(scheme + "://" + host + target)

	if err != nil {
		log.Println(err)
	}

	target_path.Path = path.Clean(target_path.Path)
	ext := path.Ext(target_path.Path)

	// Page URLs
	if ext == "" {
		res.ContentType = mime.MimeTypes[".html"]
		file_path = path.Join(ROOT_FS, target_path.Path)

		// Is Root
		if target_path.Path == "/" {
			file_path = path.Join(ROOT_FS, "index.html")
		}

		res_body, err = file_system.LoadFile(file_path)

	} else { // Resource URLs
		res.ContentType = mime.MimeTypes[ext]
		if res.ContentType == "" {
			res.ContentType = "application/octet-stream"
		}
		file_path = path.Join(ROOT_FS, target_path.Path)
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
	} else {
		res.Status = http_common.StatusOK
		res.Body = res_body
	}

	return res, err
}

func exctractRequest(c net.Conn) (string, error) {
	// Implement proper HTTP reading (read till /r/n/r/n)
	request_buffer := make([]byte, 1024)
	n, err := c.Read(request_buffer)

	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", err
		} else if errors.Is(err, net.ErrClosed) {
			log.Println("Client disconnected:", c.RemoteAddr())
			return "", err
		}
	}

	if n <= 0 {
		log.Println("No data transmitted")
		return "", err
	}

	req_raw := string(request_buffer[:n])

	return req_raw, err
}
