package http_common

import (
	"strconv"
	"strings"
	"time"
)

type HttpVersion string

const (
	V0_9 HttpVersion = "0.9"
	V1_0 HttpVersion = "1.0"
	V1_1 HttpVersion = "1.1"
	V2_0 HttpVersion = "2.0"
	V3_0 HttpVersion = "3.0"
)

var HTTP_METHODS = []string{
	"GET",
	"HEAD",
	"OPTIONS",
	"TRACE",
	"DELETE",
	"PUT",
	"POST",
	"PATCH",
	"CONNECT",
}

type HttpReq struct {
	// Request Line Informations
	Scheme  string
	Method  string
	Target  string
	Version string

	// Request Headers
	Headers map[string]string

	// Request Body
	Body []byte
}

func (req *HttpReq) ToStr() string {
	var sb strings.Builder

	// Pre-allocate roughly enough space
	sb.Grow(1024 + len(req.Body))

	// Request line GET / HTTP/1.1
	sb.WriteString(req.Method)
	sb.WriteByte(' ')
	sb.WriteString(req.Target)
	sb.WriteByte(' ')
	sb.WriteString("HTTP/")
	sb.WriteString(string(req.Version))
	sb.WriteString("\r\n")

	// Headers
	for key, value := range req.Headers {
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString(value)
		sb.WriteString("\r\n")
	}

	sb.WriteString("\r\n\r\n")

	// Body
	sb.Write(req.Body)

	return sb.String()
}

type HttpRes struct {
	// Status Line Informations
	Version HttpVersion
	Status  StatusCode

	// Response Headers
	Headers map[string]string

	// Response Body
	Body []byte
}

func (res *HttpRes) SetServerHeaders() {
	now := time.Now().UTC() // Make this configurable
	res.Headers["server"] = "dreamserver/0.0.1 (Archlinux)"
	res.Headers["Via"] = "HTTP/1.1 dreamserver"
	res.Headers["date"] = now.Format(time.RFC1123)
}

func (res *HttpRes) SetReverseProxyHeaders() {

}

func (res *HttpRes) ToStr() string {
	var sb strings.Builder

	// Pre-allocate roughly enough space
	sb.Grow(1024 + len(res.Body))

	// Status line
	sb.WriteString("HTTP/")
	sb.WriteString(string(res.Version))
	sb.WriteByte(' ')
	sb.WriteString(strconv.Itoa(int(res.Status)))
	sb.WriteByte(' ')
	sb.WriteString(res.Status.ToStr())
	sb.WriteString("\r\n")

	// Headers
	for key, value := range res.Headers {
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString(value)
		sb.WriteString("\r\n")
	}

	sb.WriteString("\r\n\r\n")

	// Body
	sb.Write(res.Body)

	return sb.String()
}

func IsValidHTTPVersion(version string) bool {
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
var StatusText = map[StatusCode]string{
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
	if msg, ok := StatusText[c]; ok {
		return msg
	}
	return "Unknown Status"
}
