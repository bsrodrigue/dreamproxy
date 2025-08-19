package http_parser

import (
	http_common "dreamproxy/http/common"
	"strings"
	"testing"
)

func TestErrorWhenEmptyRequest(t *testing.T) {
	raw_http := ""

	_, err := ParseRawHttpReq(raw_http)

	if err == nil {
		t.Errorf("Passing empty http request must return error")
	}
}

func TestErrorWhenFirstLineHasNotThreePortions(t *testing.T) {
	raw_http := "GET /foo"

	_, err := ParseRawHttpReq(raw_http)

	if err == nil {
		t.Errorf("Request first line must have three portions: <method> <path> <http-version>")
	}
}

func TestErrorWhenTargetInvalid(t *testing.T) {
	raw_http := "GET foo HTTP/1.1"

	_, err := ParseRawHttpReq(raw_http)

	if err == nil {
		t.Errorf("Invalid HTTP target")
	}
}

func TestErrorWhenVersionInvalid(t *testing.T) {
	raw_http := "GET /foo HTTP/blob"

	_, err := ParseRawHttpReq(raw_http)

	if err == nil {
		t.Errorf("Invalid HTTP version")
	}

}

func TestParseHttpMethod(t *testing.T) {
	for _, method := range http_common.HTTP_METHODS {
		raw_http := method + " / HTTP/1.1"

		parsed_http, _ := ParseRawHttpReq(raw_http)

		if parsed_http.Method != method {
			t.Errorf("parsed_http.Method = %s; want %s", parsed_http.Method, method)
		}
	}
}

func TestParseHttpTarget(t *testing.T) {
	for _, target := range http_targets {
		raw_http := "GET " + target + " HTTP/1.1"

		parsed_http, _ := ParseRawHttpReq(raw_http)

		if parsed_http.Target != target {
			t.Errorf("parsed_http.Target = %s; want %s", parsed_http.Target, target)
		}
	}
}

func TestParseHttpVersion(t *testing.T) {
	for _, version := range http_versions {
		raw_http := "GET / " + version

		parsed_http, _ := ParseRawHttpReq(raw_http)

		version_number := strings.Split(version, "/")[1]

		if parsed_http.Version != version_number {
			t.Errorf("parsed_http.Version = %s; want %s", parsed_http.Version, version_number)
		}
	}
}

var http_versions = []string{
	// HTTP/0.9 - The original HTTP (1991)
	"HTTP/0.9",

	// HTTP/1.0 - First standardized version (RFC 1945, 1996)
	"HTTP/1.0",

	// HTTP/1.1 - Most widely used version (RFC 2068/2616/7230-7235)
	"HTTP/1.1",

	// HTTP/2 - Binary protocol (RFC 7540, 2015)
	"HTTP/2.0", // Sometimes seen
	"HTTP/2",   // Standard format

	// HTTP/3 - Over QUIC (RFC 9114, 2022)
	"HTTP/3.0", // Sometimes seen
	"HTTP/3",   // Standard format
}

// Comprehensive HTTP target patterns for testing
var http_targets = []string{
	// Basic paths
	"/",
	"/index",
	"/home",
	"/about",
	"/contact",

	// Nested paths
	"/api/v1",
	"/api/v2",
	"/api/v1/users",
	"/api/v1/users/profile",
	"/admin/dashboard",
	"/user/settings/privacy",
	"/blog/2024/01/post-title",

	// Paths with file extensions
	"/index.html",
	"/style.css",
	"/script.js",
	"/image.png",
	"/document.pdf",
	"/data.json",
	"/feed.xml",
	"/sitemap.xml",

	// With single query parameters
	"/?q=search",
	"/search?query=golang",
	"/api/users?id=123",

	// With multiple query parameters
	"/?q=search&lang=en",
	"/search?query=golang&page=1&limit=10",
	"/api/users?id=123&include=profile&format=json",

	// With special characters in query
	"/search?q=hello%20world",
	"/api/data?filter=name%3D%22john%22",
	"/?utm_source=google&utm_medium=cpc&utm_campaign=spring_sale",

	// With fragments (hash)
	"/page#section1",
	"/docs#installation",
	"/article#comments",

	// Authentication and special headers scenarios
	"/login?redirect_uri=https%3A%2F%2Fexample.com%2Fdashboard",
	"/oauth/authorize?response_type=code&client_id=123&redirect_uri=callback",
	"/api/protected?token=abc123def456",

	// File uploads and downloads
	"/upload",
	"/download/file.zip",
	"/api/v1/files/upload",
	"/media/images/profile.jpg",
	"/static/css/main.min.css",

	// WebSocket endpoints (if applicable)
	"/ws",
	"/websocket",
	"/api/v1/ws/chat",

	// API versioning patterns
	"/v1/users",
	"/v2/users",
	"/api/2024-01-01/users",
	"/api/beta/features",

	// Internationalization
	"/en/home",
	"/fr/accueil",
	"/es/inicio",
	"/api/v1/i18n/messages?lang=en-US",

	// Mobile API endpoints
	"/mobile/api/v1/sync",
	"/m/dashboard",
	"/touch/interface",

	// Development and testing endpoints
	"/health",
	"/status",
	"/ping",
	"/metrics",
	"/debug/pprof",
	"/api/v1/health-check",
}
