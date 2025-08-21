package http_client

import (
	http_common "dreamproxy/http/common"
	http_parser "dreamproxy/http/parser"
	"fmt"
	"net"
	"strings"
)

type RequestConfig struct {
	Query   map[string]string
	Headers map[string]string
	Body    []byte
}

// <scheme>://<hostname>:<port>/path
// domain.com:8080/page1/sbh=80
func Get(host string, port int, path string, cfg RequestConfig) (*http_common.HttpRes, error) {

	if cfg.Headers == nil {
		cfg.Headers = make(map[string]string)
	}

	cfg.Headers["host"] = host

	if strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/") // In case path has a trailing slash
	}

	sb_path := strings.Builder{}

	sb_path.Grow(len(path) + len(cfg.Query))

	sb_path.WriteString(path)

	if len(cfg.Query) != 0 {
		sb_path.WriteString("?")
	}

	for key, val := range cfg.Query {
		sb_path.WriteString(fmt.Sprintf("%s=%s&", key, val))
	}

	path = sb_path.String()

	path = strings.TrimSuffix(path, "&")

	path += "/"

	req := http_common.HttpReq{
		Version: string(http_common.V1_1), // Make configurable
		Method:  "GET",
		Scheme:  "http",
		Target:  path,
		Headers: cfg.Headers,
		Body:    cfg.Body,
	}

	connection, err := net.Dial("tcp4", net.JoinHostPort(host, fmt.Sprint(port)))

	if err != nil {
		return nil, err
	}

	req_bytes := req.ToBytes()
	req_len := len(req_bytes)
	written_bytes := 0

	for written_bytes < req_len {
		n, err := connection.Write(req_bytes[written_bytes:])

		if n == 0 { // EOF
			return nil, err // Learn to create custom errors
		}

		if err != nil {
			return nil, err
		}

		written_bytes += n
	}

	res_str, err := http_parser.ReadFullHttpMessage(connection)

	if err != nil {
		return nil, err
	}

	res, err := http_parser.ParseRawHttpRes(res_str)

	if err != nil {
		return nil, err
	}

	return res, nil
}
