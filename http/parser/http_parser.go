package http_parser

import (
	"dreamproxy/http/common"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
)

func ParseRawHttpReq(raw_http string) (*http_common.HttpReq, error) {
	lines := strings.Split(raw_http, "\n")
	first_line := lines[0]

	if first_line == "" {
		return nil, fmt.Errorf("empty HTTP Request")
	}

	raw_parts := strings.Split(first_line, " ")

	if len(raw_parts) < 3 {
		return nil, fmt.Errorf("missing portions in first line")
	}

	raw_method := strings.TrimSpace(raw_parts[0])
	raw_target := strings.TrimSpace(raw_parts[1])
	raw_version := strings.TrimSpace(raw_parts[2])

	if !slices.Contains(http_common.HTTP_METHODS, raw_method) {
		return nil, fmt.Errorf("invalid HTTP method")
	}

	if !strings.HasPrefix(raw_target, "/") {
		return nil, fmt.Errorf("invalid HTTP target")
	}

	if !strings.HasPrefix(raw_version, "HTTP/") {
		return nil, fmt.Errorf("invalid HTTP version")
	}

	version_split := strings.Split(raw_version, "/")

	if len(version_split) != 2 {
		return nil, fmt.Errorf("invalid HTTP version")
	}

	version_number := version_split[1]

	if !http_common.IsValidHTTPVersion(version_number) {
		return nil, fmt.Errorf("invalid HTTP version:%s", version_number)
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

	return &http_common.HttpReq{
		Scheme:  "http",
		Method:  raw_method,
		Target:  raw_target,
		Version: version_number,
		Headers: headers,
	}, nil
}
func ParseRawHttpRes(raw_http string) (*http_common.HttpRes, error) {
	lines := strings.Split(strings.ReplaceAll(raw_http, "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty HTTP response")
	}

	first_line := strings.TrimSpace(lines[0])
	if first_line == "" {
		return nil, fmt.Errorf("empty HTTP response")
	}

	// HTTP/1.1 200 OK
	parts := strings.SplitN(first_line, " ", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid status line")
	}

	raw_version := strings.TrimSpace(parts[0])
	status_code_str := strings.TrimSpace(parts[1])

	if !strings.HasPrefix(raw_version, "HTTP/") {
		return nil, fmt.Errorf("invalid HTTP version")
	}

	version_split := strings.Split(raw_version, "/")
	if len(version_split) != 2 {
		return nil, fmt.Errorf("invalid HTTP version")
	}
	version_number := version_split[1]

	if !http_common.IsValidHTTPVersion(version_number) {
		return nil, fmt.Errorf("invalid HTTP version: %s", version_number)
	}

	// Parse status code
	status_code, err := strconv.Atoi(status_code_str)
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %s", status_code_str)
	}

	// Parse headers
	// TODO: Extract in separate function
	headers := map[string]string{}
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		key_val := strings.SplitN(line, ":", 2)
		if len(key_val) < 2 {
			continue
		}

		key := strings.TrimSpace(key_val[0])
		val := strings.TrimSpace(key_val[1])
		headers[key] = val
	}

	content_length, err := strconv.Atoi(headers["Content-Length"])

	if err != nil {
		log.Println("Error during content length conversion: ", err)
		return nil, err
	}

	if content_length > 0 {
		// Pinpoint body offset
		response_end := strings.Index(raw_http, "\r\n\r\n")

		if response_end == -1 {
			log.Println("Error while indexing")
			return nil, err
		}

	}

	return &http_common.HttpRes{
		Status:  http_common.StatusCode(status_code),
		Version: http_common.HttpVersion(version_number),
		Headers: headers,
	}, nil
}
