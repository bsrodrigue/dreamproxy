---

**DreamServer** is a lightweight HTTP reverse proxy and static file server written from scratch in Go.
It is designed as an educational exploration into networking, HTTP parsing, and proxying internals—while keeping a clear path toward production-ready use.

---

## ✨ Features

* **Custom HTTP parser** – handles raw TCP connections and full HTTP message parsing.
* **Reverse proxy mode** – forward requests to backend services (e.g. `djangoserver.com:8000`) with transparent header handling.
* **Static file server** – serves files from a configurable root (`staticfiles/`) with automatic MIME type detection.
* **Structured logging** – request and response logs include latency, status, bytes sent, and request IDs (UUID).
* **Error handling** – gracefully responds with `400 Bad Request` or `404 Not Found` using fallback HTML pages.
* **Connection management** – keep-alive support, versioned headers, and automatic close on errors.

---

## 🚧 Work in Progress

DreamServer is still evolving and **not yet production ready**.
The current focus areas are:

* [x] Complete HTTP message parsing
* [x] Complete support for all major HTTP methods (`POST`, `PUT`, `DELETE`, etc.)
* [x] Static File Serving
* [ ] More robust reverse proxy (streaming, TLS termination, retries)
* [ ] Configurable routing for multiple backends
* [ ] File caching and gzip compression for static assets
* [x] Structured logs written to a file (`/var/log/dreamserver/access.log`)
* [ ] Graceful shutdown and concurrent connection limits
* [ ] Security hardening (TLS, request size limits, input sanitization)

---

## 📦 Getting Started

### Prerequisites

* Go 1.21+
* Linux or macOS (Windows partially supported)

### Build

```bash
git clone https://github.com/yourname/dreamserver
cd dreamserver
go build -o dreamserver
```

### Run

```bash
./dreamserver
```

By default, DreamServer listens on `:8080` and serves files from `staticfiles/`.

---

## 🔧 Example Usage

### Serving Static Files

Put an `index.html` in `staticfiles/`, then open [http://localhost:8080/](http://localhost:8080/).

### Reverse Proxying

Requests to `http://djangoserver.com:8080/*` are proxied to a Django backend at `djangoserver.com:8000`.
Redirects are automatically followed.

---

## 📊 Logging

Each request generates a structured log entry with:

* Request ID
* Client IP
* Method, path, host
* Response status, bytes sent
* Latency (ms)

Example:

```
[INFO] DREAM_SERVER REQUEST | id=1c4f... | ip=127.0.0.1 | method=GET | path=/ | status=200 | bytes=5120 | latency=3ms
```

---

## 🧩 Project Goals

DreamServer is both a **learning playground** and a potential **foundation for a real production server**.
The long-term vision includes:

* Secure, configurable, high-performance reverse proxying
* Production-ready observability (structured logs, metrics, tracing)
* Modern HTTP/2 and HTTP/3 support
* Pluggable middleware for authentication, caching, rate limiting

---

## 📚 Why This Project?

Reinventing the wheel here isn’t wasteful—it’s deliberate.
DreamServer is meant to **demonstrate deep system knowledge**: from raw TCP handling, HTTP parsing, and state management, to building the same abstractions used in Nginx, Caddy, or HAProxy.

This is not just “toy code.” It’s a portfolio piece to show understanding of:

* Systems programming with Go
* Networking and TCP/IP
* HTTP internals
* Production-grade design tradeoffs

---

## 🤝 Contributing

Contributions, bug reports, and feedback are welcome!
This project is still under active development, and ideas for features or improvements are highly encouraged.

---

## 📜 License

MIT License. See [LICENSE](LICENSE) for details.

---

