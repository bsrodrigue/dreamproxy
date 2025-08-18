package http_parser

import (
	"dreamproxy/http/common"
	"fmt"
	"slices"
	"strings"
)

func ParseRawHttp(raw_http string) (http_common.HttpReq, error) {
	lines := strings.Split(raw_http, "\n")
	first_line := lines[0]

	if first_line == "" {
		return http_common.HttpReq{}, fmt.Errorf("empty HTTP Request")
	}

	raw_parts := strings.Split(first_line, " ")

	if len(raw_parts) < 3 {
		return http_common.HttpReq{}, fmt.Errorf("missing portions in first line")
	}

	raw_method := strings.TrimSpace(raw_parts[0])
	raw_target := strings.TrimSpace(raw_parts[1])
	raw_version := strings.TrimSpace(raw_parts[2])

	if !slices.Contains(http_common.HTTP_METHODS, raw_method) {
		return http_common.HttpReq{}, fmt.Errorf("invalid HTTP method")
	}

	if !strings.HasPrefix(raw_target, "/") {
		return http_common.HttpReq{}, fmt.Errorf("invalid HTTP target")
	}

	if !strings.HasPrefix(raw_version, "HTTP/") {
		return http_common.HttpReq{}, fmt.Errorf("invalid HTTP version")
	}

	version_split := strings.Split(raw_version, "/")

	if len(version_split) != 2 {
		return http_common.HttpReq{}, fmt.Errorf("invalid HTTP version")
	}

	version_number := version_split[1]

	if !http_common.IsValidHTTPVersion(version_number) {
		return http_common.HttpReq{}, fmt.Errorf("invalid HTTP version:%s", version_number)
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

		key := strings.TrimSpace(key_val[0])

		if len(key_val) > 2 {
			val = strings.Join(key_val[1:], ":")
			val = strings.TrimSpace(val)
		} else {
			val = key_val[1]
		}

		headers[key] = val
	}

	return http_common.HttpReq{
		Scheme:  "http",
		Method:  raw_method,
		Target:  raw_target,
		Version: version_number,
		Headers: headers,
	}, nil
}
