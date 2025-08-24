package http

import (
	"fmt"
	"net"
	"strings"
)

type RequestConfig struct {
	Query   map[string]string
	Headers map[string]string
	Body    []byte
}

func PreprocessCfg(cfg RequestConfig, host string, path string) RequestConfig {
	if cfg.Headers == nil {
		cfg.Headers = make(map[string]string)
	}

	if cfg.Headers["host"] == "" {
		cfg.Headers["host"] = host
	}

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

	return cfg
}

func HandleRequest(req HttpReq, host string, port int) (*HttpRes, error) {
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

	res_str, err := ReadFullHttpMessage(connection)

	if err != nil {
		return nil, err
	}

	res, err := ParseRawHttpRes(res_str)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func MakeRequest(method string, host string, port int, path string, cfg RequestConfig) (*HttpRes, error) {
	cfg = PreprocessCfg(cfg, host, path)

	req := HttpReq{
		Version: string(V1_1), // Make configurable
		Method:  strings.ToUpper(method),
		Scheme:  "http",
		Target:  path,
		Headers: cfg.Headers,
		Body:    cfg.Body,
	}

	return HandleRequest(req, host, port)
}

func Get(host string, port int, path string, cfg RequestConfig) (*HttpRes, error) {
	return MakeRequest("GET", host, port, path, cfg)
}

func Post(host string, port int, path string, cfg RequestConfig) (*HttpRes, error) {
	return MakeRequest("POST", host, port, path, cfg)
}

func Put(host string, port int, path string, cfg RequestConfig) (*HttpRes, error) {
	return MakeRequest("PUT", host, port, path, cfg)
}

func Patch(host string, port int, path string, cfg RequestConfig) (*HttpRes, error) {
	return MakeRequest("PATCH", host, port, path, cfg)
}

func Delete(host string, port int, path string, cfg RequestConfig) (*HttpRes, error) {
	return MakeRequest("DELETE", host, port, path, cfg)
}

func Head(host string, port int, path string, cfg RequestConfig) (*HttpRes, error) {
	return MakeRequest("HEAD", host, port, path, cfg)
}

func Options(host string, port int, path string, cfg RequestConfig) (*HttpRes, error) {
	return MakeRequest("OPTIONS", host, port, path, cfg)
}
