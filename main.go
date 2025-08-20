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

// Create a ConnectionContext struct to handle each connection in a cleaner way
func handleConn(c net.Conn) {
	// Parse HTTP header here to know whether to keep connection alive
	defer c.Close()

	for {
		req_raw, err := extractRequest(c)

		if err != nil {
			log.Println(err)
			return
		}

		req, err := http_parser.ParseRawHttpReq(req_raw.String())

		if err != nil {
			log.Println(err)
			return
		}

		req.Headers["x-forwarded-for"] = c.RemoteAddr().String()
		res, err := handleRequest(req)

		if err != nil {
			log.Println(err)
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

		res.Headers["content-length"] = fmt.Sprint(len(res.Body))

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
	}

	res.Headers["connection"] = req.Headers["connection"]

	host := req.Headers["host"]
	method := req.Method
	scheme := req.Scheme
	target_url, err := url.Parse(scheme + "://" + host + target)

	if err != nil {
		log.Println("Invalid URL:", err)
		res.Status = http_common.StatusBadRequest
		return res, err
	}

	target_url.Path = path.Clean(target_url.Path)

	// Check if Proxy
	if host == "djangoserver.com:8080" {

		server_conn, err := net.Dial("tcp4", "localhost:8000")

		if err != nil {
			log.Println(err)
		}

		// Forward Request
		n, err := server_conn.Write([]byte(req.ToStr()))

		if err != nil {
			log.Println("Bytes written: ", n)
			log.Println("Error while forwarding request: ", err)
		}

		res_buf, err := extractRequest(server_conn)

		if err != nil {
			log.Println(err)
		}

		res, err = http_parser.ParseRawHttpRes(res_buf.String())

	} else {
		switch method {
		case "HEAD":
			handleHead(target_url, res)
			break
		case "GET":
			handleGet(target_url, res)
			break
		default:
			log.Println("Invalid Method")
			break
		}
	}

	return res, nil
}

func extractRequest(c net.Conn) (*bytes.Buffer, error) {
	tmp_buf := make([]byte, 1024)
	var req_raw bytes.Buffer

	for {
		n, err := c.Read(tmp_buf)

		if err != nil {
			// End of File Reached
			if errors.Is(err, io.EOF) {
				return &req_raw, nil
			} else if errors.Is(err, net.ErrClosed) {
				log.Println("Client disconnected:", c.RemoteAddr())
				return nil, err
			} else {
				log.Println("Error while reading client socket: ", err)
				return nil, err
			}
		}

		req_raw.Write(tmp_buf[:n])

		if bytes.Index(req_raw.Bytes(), []byte("\r\n\r\n")) == -1 {
			continue
		} else {
			break
		}
	}

	return &req_raw, nil
}
