package http

import (
	"strings"
	"testing"
)

func TestParseRawHttp(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantErr   bool
		checkFunc func(req *HttpReq, res *HttpRes) // only one of req or res will be non-nil
		isReq     bool
	}{
		// ==== Request Line ====
		{
			name:    "Empty request",
			raw:     "",
			wantErr: true,
			isReq:   true,
		},
		{
			name:    "Invalid method",
			raw:     "GOT / HTTP/1.1\r\nHost: example.com\r\n\r\n",
			wantErr: true,
			isReq:   true,
		},
		{
			name:  "Extra spaces in request line",
			raw:   "GET   /   HTTP/1.1\r\nHost: example.com\r\n\r\n",
			isReq: true,
			checkFunc: func(req *HttpReq, _ *HttpRes) {
				if req.Method != "GET" || req.Target != "/" || req.Version != "1.1" {
					t.Errorf("Failed to parse request line with extra spaces")
				}
			},
		},
		{
			name:  "Absolute-form target",
			raw:   "GET http://example.com/path HTTP/1.1\r\n\r\n",
			isReq: true,
			checkFunc: func(req *HttpReq, _ *HttpRes) {
				if req.Target != "http://example.com/path" {
					t.Errorf("Failed to parse absolute-form target")
				}
			},
		},
		{
			name:  "Asterisk-form target",
			raw:   "OPTIONS * HTTP/1.1\r\n\r\n",
			isReq: true,
			checkFunc: func(req *HttpReq, _ *HttpRes) {
				if req.Target != "*" {
					t.Errorf("Failed to parse asterisk-form target")
				}
			},
		},
		{
			name:  "Authority-form target",
			raw:   "CONNECT example.com:443 HTTP/1.1\r\n\r\n",
			isReq: true,
			checkFunc: func(req *HttpReq, _ *HttpRes) {
				if req.Target != "example.com:443" {
					t.Errorf("Failed to parse authority-form target")
				}
			},
		},
		{
			name:    "Invalid HTTP version",
			raw:     "GET / HTTP/1.x\r\n\r\n",
			wantErr: true,
			isReq:   true,
		},
		// ==== Headers ====
		{
			name:  "Header with multiple colons",
			raw:   "GET / HTTP/1.1\r\nAuth: user:pass\r\n\r\n",
			isReq: true,
			checkFunc: func(req *HttpReq, _ *HttpRes) {
				if req.Headers["auth"] != "user:pass" {
					t.Errorf("Failed to parse header with multiple colons")
				}
			},
		},
		{
			name:  "Header extra whitespace",
			raw:   "GET / HTTP/1.1\r\nHost:   example.com   \r\n\r\n",
			isReq: true,
			checkFunc: func(req *HttpReq, _ *HttpRes) {
				if req.Headers["host"] != "example.com" {
					t.Errorf("Failed to trim header whitespace")
				}
			},
		},
		// ==== Body ====
		{
			name:  "Request with body",
			raw:   "POST /submit HTTP/1.1\r\nContent-Length: 29\r\n\r\nfield1=value1&field2=value2",
			isReq: true,
			checkFunc: func(req *HttpReq, _ *HttpRes) {
				if string(req.Body) != "field1=value1&field2=value2" {
					t.Errorf("Body parsing failed")
				}
			},
		},
		// ==== Res ====
		{
			name:    "Res missing status code",
			raw:     "HTTP/1.1 \r\nContent-Type: text/html\r\n\r\n",
			wantErr: true,
			isReq:   false,
		},
		{
			name:  "Res with body",
			raw:   "HTTP/1.1 200 OK\r\nContent-Length: 14\r\n\r\n<html>OK</html>",
			isReq: false,
			checkFunc: func(_ *HttpReq, res *HttpRes) {
				if string(res.Body) != "<html>OK</html>" {
					t.Errorf("Res body not parsed correctly")
				}
			},
		},
		{
			name:  "Malformed response header",
			raw:   "HTTP/1.1 200 OK\r\nContent-Type text/html\r\n\r\n",
			isReq: false,
			checkFunc: func(_ *HttpReq, res *HttpRes) {
				if len(res.Headers) != 0 {
					t.Errorf("Malformed headers should be ignored")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isReq {
				req, err := ParseRawHttpReq(tt.raw)
				if (err != nil) != tt.wantErr {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.checkFunc != nil {
					tt.checkFunc(req, nil)
				}
			} else {
				res, err := ParseRawHttpRes(tt.raw)
				if (err != nil) != tt.wantErr {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.checkFunc != nil {
					tt.checkFunc(nil, res)
				}
			}
		})
	}
}

func TestHTTPParserEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		rawHTTP string
		wantErr bool
	}{
		{
			name: "Chunked Transfer-Encoding",
			rawHTTP: "POST /upload HTTP/1.1\r\n" +
				"Host: example.com\r\n" +
				"Transfer-Encoding: chunked\r\n\r\n" +
				"4\r\nWiki\r\n5\r\npedia\r\n0\r\n\r\n",
			wantErr: false, // parser can reject or handle chunked
		},
		{
			name: "Multiple Headers with Same Name",
			rawHTTP: "GET / HTTP/1.1\r\n" +
				"Host: example.com\r\n" +
				"Cookie: a=1\r\n" +
				"Cookie: b=2\r\n\r\n",
			wantErr: false, // parser should not panic
		},
		{
			name: "Empty Header Value",
			rawHTTP: "GET / HTTP/1.1\r\n" +
				"Host: example.com\r\n" +
				"X-Empty-Header:\r\n\r\n",
			wantErr: false,
		},
		{
			name: "Header Whitespace",
			rawHTTP: "GET / HTTP/1.1\r\n" +
				"Host:   example.com   \r\n\r\n",
			wantErr: false,
		},
		{
			name:    "Only LF Newlines",
			rawHTTP: "GET / HTTP/1.1\nHost: example.com\n\n",
			wantErr: true, // parser should normalize or reject consistently
		},
		{
			name:    "Missing HTTP Version",
			rawHTTP: "GET /\r\nHost: example.com\r\n\r\n",
			wantErr: true,
		},
		{
			name: "URL-encoded Target",
			rawHTTP: "GET /path/with%20space?query=1#frag HTTP/1.1\r\n" +
				"Host: example.com\r\n\r\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRawHttpReq(tt.rawHTTP)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRawHttp() error = %v, wantErr %v", err, tt.wantErr)
			}
			// Optionally inspect req object for headers/body if needed
		})
	}
}

func TestHttpReq_ToStr(t *testing.T) {
	tests := []struct {
		name   string
		req    HttpReq
		expect string
	}{
		{
			name: "Simple GET request",
			req: HttpReq{
				Method:  "GET",
				Target:  "/",
				Version: "1.1",
				Headers: map[string]string{"Host": "example.com"},
				Body:    nil,
			},
			expect: "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n\r\n",
		},
		{
			name: "POST with body",
			req: HttpReq{
				Method:  "POST",
				Target:  "/submit",
				Version: "1.1",
				Headers: map[string]string{
					"Content-Type":   "application/x-www-form-urlencoded",
					"Content-Length": "11",
				},
				Body: []byte("hello=world"),
			},
			expect: "POST /submit HTTP/1.1\r\nContent-Type: application/x-www-form-urlencoded\r\nContent-Length: 11\r\n\r\n\r\nhello=world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.req.ToStr()
			if !strings.Contains(got, tt.expect[:len(tt.expect)-len(tt.req.Body)]) {
				t.Errorf("Request string missing headers or status line.\nGot:\n%s\nWant:\n%s", got, tt.expect)
			}
			if tt.req.Body != nil && !strings.HasSuffix(got, string(tt.req.Body)) {
				t.Errorf("Body mismatch.\nGot:\n%s\nWant body:\n%s", got, tt.req.Body)
			}
		})
	}
}

func TestHttpRes_ToStr(t *testing.T) {
	tests := []struct {
		name string
		res  HttpRes
	}{
		{
			name: "200 OK with body",
			res: HttpRes{
				Version: V1_1,
				Status:  StatusOK,
				Headers: map[string]string{
					"Content-Type":   "text/plain",
					"Content-Length": "5",
				},
				Body: []byte("hello"),
			},
		},
		{
			name: "404 Not Found no body",
			res: HttpRes{
				Version: V1_1,
				Status:  StatusNotFound,
				Headers: map[string]string{},
				Body:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.res.ToStr()

			// Must start with HTTP version and status line
			if !strings.HasPrefix(got, "HTTP/"+string(tt.res.Version)+" ") {
				t.Errorf("Response must start with HTTP version, got:\n%s", got)
			}

			// Status code and text must be present
			if !strings.Contains(got, tt.res.Status.ToStr()) {
				t.Errorf("Response missing status text, got:\n%s", got)
			}

			// Body check
			if len(tt.res.Body) > 0 && !strings.HasSuffix(got, string(tt.res.Body)) {
				t.Errorf("Response body mismatch.\nGot:\n%s\nWant body:\n%s", got, tt.res.Body)
			}
		})
	}
}
