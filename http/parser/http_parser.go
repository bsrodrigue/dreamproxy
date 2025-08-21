package http_parser

import (
	"bytes"
	"dreamproxy/http/common"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// HTTP Request Format
// <method> <target> <version>\r\n
// <header>:<value>\r\n
// ....
// <header>:<value>\r\n\r\n
// <body>

func ParseRawHttpReq(raw_http string) (*http_common.HttpReq, error) {
	portions := strings.SplitN(raw_http, "\r\n", 2)

	if len(portions) != 2 {
		return nil, fmt.Errorf("Invalid HTTP Request")
	}

	request_line := portions[0]
	rest := strings.Split(portions[1], "\r\n\r\n")

	if len(rest) == 0 {
		return nil, fmt.Errorf("No HTTP Headers provided")
	}

	raw_headers := rest[0]

	//-------------------------- Parse Request Line

	request_line_parts := strings.Fields(request_line)

	if len(request_line_parts) != 3 {
		return nil, fmt.Errorf("missing parts on request line")
	}

	raw_method := strings.TrimSpace(request_line_parts[0])
	raw_target := strings.TrimSpace(request_line_parts[1])
	raw_version := strings.TrimSpace(request_line_parts[2])

	if !slices.Contains(http_common.HTTP_METHODS, raw_method) {
		return nil, fmt.Errorf("invalid HTTP method")
	}

	// Check Target Form
	if !isValidTarget(raw_target, strings.ToUpper(raw_method)) {
		return nil, fmt.Errorf("invalid HTTP target")
	}

	if !strings.HasPrefix(raw_version, "HTTP/") {
		return nil, fmt.Errorf("invalid HTTP version")
	}

	version_split := strings.Split(raw_version, "/")

	if len(version_split) != 2 {
		return nil, fmt.Errorf("invalid HTTP version")
	}

	version_number := version_split[1]

	if !http_common.IsValidHTTPVersion(version_number) {
		return nil, fmt.Errorf("invalid HTTP version:%s", version_number)
	}

	//-------------------------- Parse Headers

	headers := ParseHttpHeaders(raw_headers)

	//-------------------------- Parse Body
	var raw_body string

	if len(rest) == 2 {
		raw_body = rest[1]
	}

	return &http_common.HttpReq{
		Scheme:  "http",
		Method:  raw_method,
		Target:  raw_target,
		Version: version_number,
		Headers: headers,
		Body:    []byte(raw_body),
	}, nil
}

// HTTP Response Format
// <version> <status-code> <status-message>\r\n
// <header>:<value>\r\n
// ....
// <header>:<value>\r\n\r\n
// <body>

func ParseRawHttpRes(raw_http string) (*http_common.HttpRes, error) {

	portions := strings.SplitN(raw_http, "\r\n", 2)

	if len(portions) != 2 {
		return nil, fmt.Errorf("Invalid HTTP Response")
	}

	status_line := portions[0]
	rest := strings.Split(portions[1], "\r\n\r\n")

	if len(rest) == 0 {
		return nil, fmt.Errorf("No HTTP Headers provided")
	}

	raw_headers := rest[0]

	if status_line == "" {
		return nil, fmt.Errorf("empty HTTP response")
	}

	//-------------------------- Parse Status Line
	// Reason phrase is handled by our server
	parts := strings.SplitN(status_line, " ", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid status line")
	}

	raw_version := strings.TrimSpace(parts[0])
	status_code_str := strings.TrimSpace(parts[1])

	if !strings.HasPrefix(raw_version, "HTTP/") {
		return nil, fmt.Errorf("invalid HTTP version")
	}

	version_split := strings.Split(raw_version, "/")
	if len(version_split) != 2 {
		return nil, fmt.Errorf("invalid HTTP version")
	}
	version_number := version_split[1]

	if !http_common.IsValidHTTPVersion(version_number) {
		return nil, fmt.Errorf("invalid HTTP version: %s", version_number)
	}

	// Parse status code
	status_code, err := strconv.Atoi(status_code_str)
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %s", status_code_str)
	}

	//-------------------------- Parse Status Line
	headers := ParseHttpHeaders(raw_headers)

	var raw_body string

	if len(rest) == 2 {
		raw_body = rest[1]
	}

	return &http_common.HttpRes{
		Status:  http_common.StatusCode(status_code),
		Version: http_common.HttpVersion(version_number),
		Headers: headers,
		Body:    []byte(raw_body),
	}, nil
}

func ParseHttpHeaders(raw_headers string) map[string]string {
	lines := strings.Split(raw_headers, "\r\n")

	headers := map[string]string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		key_val := strings.SplitN(line, ":", 2)
		if len(key_val) < 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(key_val[0]))
		val := strings.TrimSpace(key_val[1])
		headers[key] = val
	}

	return headers
}

var (
	originForm    = regexp.MustCompile(`^/[^ ]*$`)
	absoluteForm  = regexp.MustCompile(`^https?://[^ ]+$`)
	authorityForm = regexp.MustCompile(`^[^/:]+(:[0-9]+)?$`) // host[:port]
	asteriskForm  = regexp.MustCompile(`^\*$`)
)

func isValidTarget(target string, method string) bool {
	switch {
	case asteriskForm.MatchString(target):
		return true
	case originForm.MatchString(target):
		return true
	case absoluteForm.MatchString(target):
		return true
	case method == "CONNECT" && authorityForm.MatchString(target):
		return true
	default:
		return false
	}
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

	headers := ParseHttpHeaders(header_str)

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
