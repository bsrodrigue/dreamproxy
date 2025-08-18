package http_parser

import (
	"fmt"
	"slices"
	"strings"
)

var http_methods = []string{
	"GET",
	"POST",
	"PUT",
	"PATCH",
	"DELETE",
	"OPTIONS",
	"HEAD",
}

type HttpReq struct {
	Method  string
	Target  string
	Version string
	Headers map[string]string
	Body    []byte
}

type HttpRes struct {
	Version       HttpVersion
	Status        StatusCode
	Server        string
	ContentLength int
	ContentType   string
	Connection    string
	Body          []byte
}

func (res HttpRes) ToStr() string {
	body_str := string(res.Body)

	response_str := fmt.Sprintf(
		"HTTP/%s %d %s\r\n"+
			"Server: dreamserver/0.0.1 (Archlinux)\r\n"+
			"Content-Length: %d\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n"+
			"Connection: close\r\n\r\n"+
			"%s",
		res.Version,
		res.Status,
		res.Status.ToStr(),
		len(body_str), body_str,
	)

	return response_str
}

func ParseRawHttp(raw_http string) (HttpReq, error) {
	lines := strings.Split(raw_http, "\n")
	first_line := lines[0]

	if first_line == "" {
		return HttpReq{}, fmt.Errorf("empty HTTP Request")
	}

	raw_parts := strings.Split(first_line, " ")

	if len(raw_parts) < 3 {
		return HttpReq{}, fmt.Errorf("missing portions in first line")
	}

	raw_method := strings.TrimSpace(raw_parts[0])
	raw_target := strings.TrimSpace(raw_parts[1])
	raw_version := strings.TrimSpace(raw_parts[2])

	if !slices.Contains(http_methods, raw_method) {
		return HttpReq{}, fmt.Errorf("invalid HTTP method")
	}

	if !strings.HasPrefix(raw_target, "/") {
		return HttpReq{}, fmt.Errorf("invalid HTTP target")
	}

	if !strings.HasPrefix(raw_version, "HTTP/") {
		return HttpReq{}, fmt.Errorf("invalid HTTP version")
	}

	version_split := strings.Split(raw_version, "/")

	if len(version_split) != 2 {
		return HttpReq{}, fmt.Errorf("invalid HTTP version")
	}

	version_number := version_split[1]

	if !isValidHTTPVersion(version_number) {
		return HttpReq{}, fmt.Errorf("invalid HTTP version:%s", version_number)
	}

	// Handle Headers
	header_lines := lines[1:]
	headers := map[string]string{}

	for _, line := range header_lines {
		key_val := strings.Split(line, ":")
		val := string("")

		if len(key_val) < 2 {
			continue
		}

		key := key_val[0]

		if len(key_val) > 2 {
			val = strings.Join(key_val[1:], ":")
		} else {
			val = key_val[1]
		}

		headers[key] = val
	}

	return HttpReq{
		Method:  raw_method,
		Target:  raw_target,
		Version: version_number,
		Headers: headers,
	}, nil
}

func isValidHTTPVersion(version string) bool {
	validVersions := map[string]bool{
		"0.9": true,
		"1.0": true,
		"1.1": true,
		"2":   true,
		"2.0": true,
		"3":   true,
		"3.0": true,
	}
	return validVersions[version]
}

type HttpVersion string

const (
	V0_9 HttpVersion = "0.9"
	V1_0 HttpVersion = "1.0"
	V1_1 HttpVersion = "1.1"
	V2_0 HttpVersion = "2.0"
	V3_0 HttpVersion = "3.0"
)

// StatusCode represents an HTTP status code.
type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusCreated             StatusCode = 201
	StatusAccepted            StatusCode = 202
	StatusNoContent           StatusCode = 204
	StatusMovedPermanently    StatusCode = 301
	StatusFound               StatusCode = 302
	StatusBadRequest          StatusCode = 400
	StatusUnauthorized        StatusCode = 401
	StatusForbidden           StatusCode = 403
	StatusNotFound            StatusCode = 404
	StatusMethodNotAllowed    StatusCode = 405
	StatusConflict            StatusCode = 409
	StatusInternalServerError StatusCode = 500
	StatusNotImplemented      StatusCode = 501
	StatusBadGateway          StatusCode = 502
	StatusServiceUnavailable  StatusCode = 503
)

// statusText maps HTTP status codes to their messages.
var statusText = map[StatusCode]string{
	StatusOK:                  "OK",
	StatusCreated:             "Created",
	StatusAccepted:            "Accepted",
	StatusNoContent:           "No Content",
	StatusMovedPermanently:    "Moved Permanently",
	StatusFound:               "Found",
	StatusBadRequest:          "Bad Request",
	StatusUnauthorized:        "Unauthorized",
	StatusForbidden:           "Forbidden",
	StatusNotFound:            "Not Found",
	StatusMethodNotAllowed:    "Method Not Allowed",
	StatusConflict:            "Conflict",
	StatusInternalServerError: "Internal Server Error",
	StatusNotImplemented:      "Not Implemented",
	StatusBadGateway:          "Bad Gateway",
	StatusServiceUnavailable:  "Service Unavailable",
}

// Text returns the standard text for the HTTP status code.
func (c StatusCode) ToStr() string {
	if msg, ok := statusText[c]; ok {
		return msg
	}
	return "Unknown Status"
}
