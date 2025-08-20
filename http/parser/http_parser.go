package http_parser

import (
	"bytes"
	"dreamproxy/http/common"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
)

func ParseRawHttpReq(raw_http string) (*http_common.HttpReq, error) {
	portions := strings.SplitN(raw_http, "\r\n", 2)

	if len(portions) == 0 {
		return nil, fmt.Errorf("Empty HTTP Request")
	}

	request_line := portions[0]
	rest := strings.Split(portions[1], "\r\n\r\n")

	if len(rest) == 0 {
		return nil, fmt.Errorf("No HTTP Headers provided")
	}

	raw_header := rest[0]

	// Parse Body (Optional)

	raw_parts := strings.Split(request_line, " ")

	if len(raw_parts) < 3 {
		return nil, fmt.Errorf("missing parts on request line")
	}

	raw_method := strings.TrimSpace(raw_parts[0])
	raw_target := strings.TrimSpace(raw_parts[1])
	raw_version := strings.TrimSpace(raw_parts[2])

	if !slices.Contains(http_common.HTTP_METHODS, raw_method) {
		return nil, fmt.Errorf("invalid HTTP method")
	}

	// Not permissive enough?
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
	headers := ParseHttpHeaders(raw_header)

	return &http_common.HttpReq{
		Scheme:  "http",
		Method:  raw_method,
		Target:  raw_target,
		Version: version_number,
		Headers: headers,
	}, nil
}

func ParseRawHttpRes(raw_http string) (*http_common.HttpRes, error) {
	var body bytes.Buffer
	lines := strings.Split(raw_http, "\r\n")

	if len(lines) == 0 {
		return nil, fmt.Errorf("empty HTTP response")
	}

	status_line := strings.TrimSpace(lines[0])

	if status_line == "" {
		return nil, fmt.Errorf("empty HTTP response")
	}

	// HTTP/1.1 200 OK
	parts := strings.SplitN(status_line, " ", 3)
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
		return nil, fmt.Errorf("Error during content length conversion: %s", err)
	}

	if content_length > 0 {
		body.Grow(content_length)
		// Pinpoint body offset
		response_end := strings.Index(raw_http, "\r\n\r\n")

		if response_end == -1 {
			log.Println("Error while indexing")
			return nil, err
		}

		response_end += 4 // Move offset to body

		body.WriteString(raw_http[response_end : response_end+content_length])
	}

	return &http_common.HttpRes{
		Status:  http_common.StatusCode(status_code),
		Version: http_common.HttpVersion(version_number),
		Headers: headers,
		Body:    body.Bytes(),
	}, nil
}

func ParseHttpHeaders(raw_headers string) map[string]string {
	lines := strings.Split(raw_headers, "\r\n")

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

	return headers
}
