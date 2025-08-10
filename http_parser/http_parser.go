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
	Status  int
	Headers map[string]string
	Body    []byte
}

func ParseRawHttp(raw_http string) (HttpReq, error) {
	lines := strings.Split(raw_http, "\n")
	first_line := lines[0]

	if first_line == "" {
		return HttpReq{Status: 400}, fmt.Errorf("empty HTTP Request")
	}

	raw_parts := strings.Split(first_line, " ")

	if len(raw_parts) < 3 {
		return HttpReq{Status: 400}, fmt.Errorf("missing portions in first line")
	}

	raw_method := strings.TrimSpace(raw_parts[0])
	raw_target := strings.TrimSpace(raw_parts[1])
	raw_version := strings.TrimSpace(raw_parts[2])

	if !slices.Contains(http_methods, raw_method) {
		return HttpReq{Status: 400}, fmt.Errorf("invalid HTTP method")
	}

	if !strings.HasPrefix(raw_target, "/") {
		return HttpReq{Status: 400}, fmt.Errorf("invalid HTTP target")
	}

	if !strings.HasPrefix(raw_version, "HTTP/") {
		return HttpReq{Status: 400}, fmt.Errorf("invalid HTTP version")
	}

	version_split := strings.Split(raw_version, "/")

	if len(version_split) != 2 {
		return HttpReq{Status: 400}, fmt.Errorf("invalid HTTP version")
	}

	version_number := version_split[1]

	if !isValidHTTPVersion(version_number) {
		return HttpReq{Status: 400}, fmt.Errorf("invalid HTTP version:%s", version_number)
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
