package http

import (
	"net"
	"testing"
	"time"

	"dreamproxy/multidict"
)

func TestMakeTarget(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		query map[string]string
		want  string
	}{
		{
			name:  "no query",
			path:  "/foo",
			query: nil,
			want:  "/foo",
		},
		{
			name:  "empty query",
			path:  "/foo",
			query: map[string]string{},
			want:  "/foo",
		},
		{
			name:  "single query param",
			path:  "/search",
			query: map[string]string{"q": "hello"},
			want:  "/search?q=hello",
		},
		{
			name:  "multiple params",
			path:  "/",
			query: map[string]string{"a": "1", "b": "2"},
			want:  "/?a=1&b=2",
		},
		{
			name:  "URL-encodes values",
			path:  "/path",
			query: map[string]string{"key": "a&b=c"},
			want:  "/path?key=a%26b%3Dc",
		},
		{
			name:  "URL-encodes spaces",
			path:  "/path",
			query: map[string]string{"q": "hello world"},
			want:  "/path?q=hello+world",
		},
		{
			name:  "path with special chars is preserved",
			path:  "/a/b/c",
			query: map[string]string{"id": "42"},
			want:  "/a/b/c?id=42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := makeTarget(tt.path, tt.query)
			if got != tt.want {
				t.Errorf("makeTarget(%q, %v) = %q, want %q", tt.path, tt.query, got, tt.want)
			}
		})
	}
}

func TestPrepareRequest(t *testing.T) {
	t.Run("sets host header when missing", func(t *testing.T) {
		cfg := RequestConfig{
			Headers: multidict.NewMultiDict(),
		}
		cfg, _ = PrepareRequest(cfg, "example.com", "/")
		if cfg.Headers.GetOne("host") != "example.com" {
			t.Errorf("host header should be set, got %q", cfg.Headers.GetOne("host"))
		}
	})

	t.Run("preserves existing host header", func(t *testing.T) {
		h := multidict.NewMultiDict()
		h.Set("host", "custom.com")
		cfg := RequestConfig{Headers: h}

		cfg, _ = PrepareRequest(cfg, "example.com", "/")
		if cfg.Headers.GetOne("host") != "custom.com" {
			t.Errorf("existing host header should be preserved, got %q", cfg.Headers.GetOne("host"))
		}
	})

	t.Run("does not mutate caller's headers", func(t *testing.T) {
		original := multidict.NewMultiDict()
		original.Set("accept", "text/html")

		cfg := RequestConfig{Headers: original}
		PrepareRequest(cfg, "proxied.com", "/")

		if original.GetOne("host") != "" {
			t.Errorf("caller's original MultiDict should not be mutated, got host = %q", original.GetOne("host"))
		}
		if original.GetOne("accept") != "text/html" {
			t.Errorf("caller's original header should remain")
		}
	})

	t.Run("copies all headers from original", func(t *testing.T) {
		h := multidict.NewMultiDict()
		h.Set("cookie", "a=1")
		h.Set("cookie", "b=2")

		cfg := RequestConfig{Headers: h}
		cfg, _ = PrepareRequest(cfg, "example.com", "/")

		vals, ok := cfg.Headers.Get("cookie")
		if !ok || len(vals) != 2 {
			t.Errorf("should have both cookie values, got %v", vals)
		}
	})

	t.Run("builds target path with no query", func(t *testing.T) {
		cfg := RequestConfig{Headers: multidict.NewMultiDict()}
		_, target := PrepareRequest(cfg, "example.com", "/api")
		if target != "/api" {
			t.Errorf("target = %q, want %q", target, "/api")
		}
	})

	t.Run("builds target path with query", func(t *testing.T) {
		cfg := RequestConfig{
			Query:   map[string]string{"q": "test"},
			Headers: multidict.NewMultiDict(),
		}
		_, target := PrepareRequest(cfg, "example.com", "/search")
		if target != "/search?q=test" {
			t.Errorf("target = %q, want %q", target, "/search?q=test")
		}
	})
}

func TestHandleRequestDialError(t *testing.T) {
	// Connecting to a port that's not listening should fail fast (timeout or refused)
	_, err := HandleRequest(HTTPReq{}, "127.0.0.1", 1)
	if err == nil {
		t.Fatal("expected error when dialing unreachable port, got nil")
	}
}

func TestHandleRequestWriteError(t *testing.T) {
	// Listen on a random port, accept, then close immediately so the write fails
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		conn, _ := ln.Accept()
		conn.Close()
	}()

	time.Sleep(50 * time.Millisecond)

	_, err = HandleRequest(HTTPReq{
		Method:  "GET",
		Target:  "/",
		Version: "1.1",
		Body:    []byte("some data"),
	}, "127.0.0.1", ln.Addr().(*net.TCPAddr).Port)
	if err == nil {
		t.Fatal("expected error when writing to closed connection, got nil")
	}
}

func TestMakeRequestInvalidHost(t *testing.T) {
	_, err := MakeRequest("GET", "", -1, "/", RequestConfig{
		Headers: multidict.NewMultiDict(),
	})
	if err == nil {
		t.Fatal("expected error for invalid host/port, got nil")
	}
}

func TestShortcutFunctions(t *testing.T) {
	tests := []struct {
		name     string
		call     func(string, int, string, RequestConfig) (*HTTPRes, error)
		expected string // method that should appear in the error
	}{
		{"Get", func(h string, p int, pa string, c RequestConfig) (*HTTPRes, error) { return Get(h, p, pa, c) }, "GET"},
		{"Post", Post, "POST"},
		{"Put", Put, "PUT"},
		{"Patch", Patch, "PATCH"},
		{"Delete", Delete, "DELETE"},
		{"Head", Head, "HEAD"},
		{"Options", Options, "OPTIONS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// All shortcuts fail the same way: bad address
			_, err := tt.call("", -1, "/", RequestConfig{Headers: multidict.NewMultiDict()})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
