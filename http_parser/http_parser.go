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

func ParseRawHttp(raw_http string) (HttpReq, error) {
	if raw_http == "" {
		return HttpReq{}, fmt.Errorf("Empty HTTP Request")
	}

	raw_parts := strings.Split(raw_http, " ")

	if len(raw_parts) < 3 {
		return HttpReq{}, fmt.Errorf("Missing portions in first line")
	}

	raw_method := strings.TrimSpace(raw_parts[0])
	raw_target := strings.TrimSpace(raw_parts[1])
	raw_version := strings.TrimSpace(raw_parts[2])

	if !slices.Contains(http_methods, raw_method) {
		return HttpReq{}, fmt.Errorf("Invalid HTTP method")
	}

	if !strings.HasPrefix(raw_target, "/") {
		return HttpReq{}, fmt.Errorf("Invalid HTTP target")
	}

	if !strings.HasPrefix(raw_version, "HTTP/") {
		return HttpReq{}, fmt.Errorf("Invalid HTTP version")
	}

	version_split := strings.Split(raw_version, "/")

	if len(version_split) != 2 {
		return HttpReq{}, fmt.Errorf("Invalid HTTP version")
	}

	version_number := version_split[1]

	if !isValidHTTPVersion(version_number) {
		return HttpReq{}, fmt.Errorf("Invalid HTTP version:%s", version_number)
	}

	return HttpReq{
		Method:  raw_method,
		Target:  raw_target,
		Version: version_number,
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
