package http_common

import "fmt"

var HTTP_METHODS = []string{
	"GET",
	"POST",
	"PUT",
	"PATCH",
	"DELETE",
	"OPTIONS",
	"HEAD",
}

type HttpReq struct {
	Scheme  string
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
			"Content-Type: %s\r\n"+
			"Connection: %s\r\n\r\n"+
			"%s",
		res.Version,
		res.Status,
		res.Status.ToStr(),
		res.ContentLength,
		res.ContentType,
		res.Connection,
		body_str,
	)

	return response_str
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
