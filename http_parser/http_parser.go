package http_parser

import (
	"fmt"
	"slices"
	"strconv"
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

func ParseRawHttp(raw string) (HttpReq, error) {
	if raw == "" {
		return HttpReq{}, fmt.Errorf("Empty HTTP Request")
	}

	raw_portions := strings.Split(raw, " ")

	if len(raw_portions) < 3 {
		return HttpReq{}, fmt.Errorf("Missing portions in first line")
	}

	raw_method := raw_portions[0]
	raw_target := raw_portions[1]
	raw_version := raw_portions[2]

	if !slices.Contains(http_methods, raw_method) {
		return HttpReq{}, fmt.Errorf("Invalid HTTP method")
	}

	if !strings.HasPrefix(raw_target, "/") {
		return HttpReq{}, fmt.Errorf("Invalid HTTP target")
	}

	version_split := strings.Split(raw_version, "/")

	if len(version_split) != 2 {
		return HttpReq{}, fmt.Errorf("Invalid HTTP version")
	}

	version_number := version_split[1]

	_, err := strconv.ParseFloat(version_number, 64)

	if err != nil {
		return HttpReq{}, fmt.Errorf("Invalid HTTP version: Expected float")
	}

	return HttpReq{
		Method:  raw_method,
		Target:  raw_target,
		Version: version_number,
	}, nil
}
