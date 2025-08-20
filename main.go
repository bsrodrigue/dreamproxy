package main

import (
	"bytes"
	"dreamproxy/file_system"
	http_common "dreamproxy/http/common"
	"dreamproxy/http/parser"
	"dreamproxy/mime"
	"errors"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"net/url"
	"os"
	"path"
	"strconv"
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
		req_raw, err := ReadFullHttpMessage(c)

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

		server_conn, err := net.Dial("tcp4", "localhost:8000")

		if err != nil {
			return CreateBadRequestRes(), err
		}

		// Forward Request
		n, err := server_conn.Write([]byte(req.ToStr()))

		if err != nil {
			log.Println("Bytes written: ", n)
			return CreateBadRequestRes(), err
		}

		res_buf, err := ReadFullHttpMessage(server_conn)

		if err != nil {
			return CreateBadRequestRes(), err
		}

		res, err = http_parser.ParseRawHttpRes(res_buf)

		if err != nil {
			return CreateBadRequestRes(), err
		}

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

func ExtractHeadersAndBodyStart(c net.Conn) ([]byte, []byte, error) {
	var req_buf bytes.Buffer
	var body_buf bytes.Buffer
	var end_of_headers int = -1
	var err error
	tmp_buf := make([]byte, 1024)

	for {
		n, err := c.Read(tmp_buf)

		// EOF
		if n == 0 {
			log.Println("Client disconnected before full message")
			return nil, nil, err
		}

		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("Client disconnected:", c.RemoteAddr())
				return nil, nil, err
			} else {
				log.Println("Error while reading client socket: ", err)
				return nil, nil, err
			}
		}

		_, err = req_buf.Write(tmp_buf[:n])

		if err != nil {
			log.Println(err)
		}

		// End of Headers?
		end_of_headers = bytes.Index(req_buf.Bytes(), []byte("\r\n\r\n"))

		if end_of_headers == -1 {
			continue
		}

		break
	}

	leftover_bytes := req_buf.Bytes()[end_of_headers:]

	if len(leftover_bytes)-4 != 0 { // There are body leftovers
		leftover_bytes = leftover_bytes[4:]
		body_buf.Grow(len(leftover_bytes) + 1024)
		body_buf.Write(leftover_bytes)
	}

	req_bytes := req_buf.Bytes()

	req_bytes = req_bytes[:end_of_headers+4]

	return req_bytes, body_buf.Bytes(), err
}

func ReadFullHttpMessage(c net.Conn) (string, error) {
	tmp_buf := make([]byte, 2048)

	var req_buf bytes.Buffer
	var body_buf bytes.Buffer
	var max_body_len int
	var keepAlive bool = false

	req_bytes, body_bytes, err := ExtractHeadersAndBodyStart(c)

	if err != nil {
		return "", err
	}

	body_buf.Grow(1024 + len(body_bytes))
	req_buf.Grow(1024 + len(req_bytes))

	body_buf.Write(body_bytes)
	req_buf.Write(req_bytes)

	// Extract Headers
	eo_reqline := bytes.Index(req_bytes, []byte("\r\n"))

	header_bytes := req_bytes[eo_reqline+2:]

	header_str := string(header_bytes)

	headers := http_parser.ParseHttpHeaders(header_str)

	content_length := headers["content-length"]
	keepAlive = strings.ToLower(headers["connection"]) == "keep-alive"

	if content_length == "" || content_length == "0" {
		max_body_len = 0
	} else {
		max_body_len, err = strconv.Atoi(content_length)

		if err != nil {
			log.Println("Error while parsing content-length: ", err)
			return "", err
		}
	}

	// Check leftover body content from previous reads
	body_read := len(body_bytes)

	for max_body_len != 0 && body_read < max_body_len {
		n, err := c.Read(tmp_buf)

		// EOF
		if n == 0 && !keepAlive {
			log.Println("Client disconnected before full body")
			return "", err
		}

		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("Client disconnected:", c.RemoteAddr())
				return "", err
			} else {
				log.Println("Error while reading client socket: ", err)
				return "", err
			}
		}

		body_read += n
		body_buf.Write(tmp_buf[:n])
	}

	n, err := req_buf.Write(body_buf.Bytes())

	if n < max_body_len {
		log.Println("Truncated")
	}

	if err != nil {
		log.Println("Error while assembling request: ", err)
		return "", err
	}

	return req_buf.String(), nil
}
