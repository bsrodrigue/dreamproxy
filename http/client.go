package http

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"dreamproxy/multidict"
)

const dialTimeout = 10 * time.Second

type RequestConfig struct {
	Query   map[string]string
	Headers multidict.MultiDict
	Body    []byte
}

func makeTarget(path string, query map[string]string) string {
	var sb strings.Builder
	sb.Grow(len(path) + len(query)*50)

	sb.WriteString(path)

	if len(query) > 0 {
		sb.WriteByte('?')
		q := url.Values{}
		for key, val := range query {
			q.Set(key, val)
		}
		sb.WriteString(q.Encode())
	}

	return sb.String()
}

func PrepareRequest(cfg RequestConfig, host string, path string) (RequestConfig, string) {
	headers := multidict.NewMultiDict()
	for key, vals := range cfg.Headers.Map() {
		for _, val := range vals {
			headers.Set(key, val)
		}
	}

	if headers.GetOne("host") == "" {
		headers.Set("host", host)
	}

	cfg.Headers = headers
	target := makeTarget(path, cfg.Query)

	return cfg, target
}

func HandleRequest(req HTTPReq, host string, port int) (*HTTPRes, error) {
	connection, err := net.DialTimeout("tcp4", net.JoinHostPort(host, strconv.Itoa(port)), dialTimeout)
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	req_bytes := req.ToBytes()
	req_len := len(req_bytes)
	written_bytes := 0

	for written_bytes < req_len {
		n, err := connection.Write(req_bytes[written_bytes:])
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, fmt.Errorf("connection closed while writing request")
		}

		written_bytes += n
	}

	res_str, err := ReadFullHttpMessage(connection)
	if err != nil {
		return nil, err
	}

	return ParseRawHttpRes(res_str)
}

func MakeRequest(method string, host string, port int, path string, cfg RequestConfig) (*HTTPRes, error) {
	cfg, target := PrepareRequest(cfg, host, path)

	req := HTTPReq{
		Version: string(V1_1),
		Method:  strings.ToUpper(method),
		Scheme:  "http",
		Target:  target,
		Headers: cfg.Headers,
		Body:    cfg.Body,
	}

	return HandleRequest(req, host, port)
}

func Get(host string, port int, path string, cfg RequestConfig) (*HTTPRes, error) {
	return MakeRequest("GET", host, port, path, cfg)
}

func Post(host string, port int, path string, cfg RequestConfig) (*HTTPRes, error) {
	return MakeRequest("POST", host, port, path, cfg)
}

func Put(host string, port int, path string, cfg RequestConfig) (*HTTPRes, error) {
	return MakeRequest("PUT", host, port, path, cfg)
}

func Patch(host string, port int, path string, cfg RequestConfig) (*HTTPRes, error) {
	return MakeRequest("PATCH", host, port, path, cfg)
}

func Delete(host string, port int, path string, cfg RequestConfig) (*HTTPRes, error) {
	return MakeRequest("DELETE", host, port, path, cfg)
}

func Head(host string, port int, path string, cfg RequestConfig) (*HTTPRes, error) {
	return MakeRequest("HEAD", host, port, path, cfg)
}

func Options(host string, port int, path string, cfg RequestConfig) (*HTTPRes, error) {
	return MakeRequest("OPTIONS", host, port, path, cfg)
}
