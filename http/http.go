package http

import (
	"strconv"
	"strings"
	"time"

	"dreamproxy/format"
	"dreamproxy/multidict"
)

type HTTPVersion string

const (
	V0_9 HTTPVersion = "0.9"
	V1_0 HTTPVersion = "1.0"
	V1_1 HTTPVersion = "1.1"
	V2_0 HTTPVersion = "2.0"
	V3_0 HTTPVersion = "3.0"
)

var HTTPMethods = []string{
	// Regular verbs
	"GET",
	"HEAD",
	"OPTIONS",
	"DELETE",
	"PUT",
	"POST",
	"PATCH",

	"TRACE",   // Echoes back request
	"CONNECT", // Used for proxies
}

type HTTPReq struct {
	// Request Line Informations
	Scheme  string
	Method  string
	Target  string
	Version string

	// Request Headers
	Headers multidict.MultiDict

	// Request Body
	Body []byte
}

func (req *HTTPReq) ToStr() string {
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
	for key, values := range req.Headers.Map() {
		for _, value := range values {
			sb.WriteString(key)
			sb.WriteString(": ")
			sb.WriteString(value)
			sb.WriteString("\r\n")
		}
	}

	sb.WriteString("\r\n")

	// Body
	sb.Write(req.Body)

	return sb.String()
}

func (req *HTTPReq) ToBytes() []byte {
	return []byte(req.ToStr())
}

type HttpRes struct {
	// Status Line Informations
	Version HTTPVersion
	Status  StatusCode

	// Response Headers
	Headers multidict.MultiDict

	// Response Body
	Body []byte
}

func CreateHttpRes() *HttpRes {
	return &HttpRes{
		Version: V1_1,
		Headers: multidict.NewMultiDict(),
	}
}

func (res *HttpRes) SetServerHeaders() {
	now := time.Now().UTC() // Make this configurable

	res.Headers.Set("server", "dreamserver/0.0.1 (Archlinux)")
	res.Headers.Set("Via", "HTTP/1.1 dreamserver")
	res.Headers.Set("date", format.TimeToGMT(now))
}

func (res *HttpRes) SetReverseProxyHeaders() {
}

func (res *HttpRes) ToBytes() []byte {
	return []byte(res.ToStr())
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
	for key, values := range res.Headers.Map() {
		for _, value := range values {
			sb.WriteString(key)
			sb.WriteString(": ")
			sb.WriteString(value)
			sb.WriteString("\r\n")
		}
	}

	sb.WriteString("\r\n")

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
	StatusPartialContent      StatusCode = 206
	StatusMovedPermanently    StatusCode = 301
	StatusFound               StatusCode = 302
	StatusNotModified         StatusCode = 304
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
	StatusPartialContent:      "Partial Content",
	StatusMovedPermanently:    "Moved Permanently",
	StatusFound:               "Found",
	StatusNotModified:         "Not Modified",
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
