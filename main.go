package main

import (
	"dreamproxy/http_parser"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

const PROTOCOL string = "tcp4"
const PORT string = "8080"

func main() {
	ln, err := net.Listen(PROTOCOL, fmt.Sprintf(":%s", PORT))

	if err != nil {
		log.Fatal(err)
	}

	log.Println(fmt.Sprintf("Listening on :%s", PORT))

	defer ln.Close()

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(c net.Conn) {
	// Parse HTTP header here to know whether to keep connection alive
	defer c.Close()

	request_buffer := make([]byte, 1024)

	for {
		n, err := c.Read(request_buffer)

		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				log.Println("Client disconnected:", c.RemoteAddr())
				return

			}
			log.Println(err)
			return
		}

		if n <= 0 {
			log.Println("No data transmitted")
			return
		}

		request_str := string(request_buffer)

		line := strings.Split(request_str, "\n")[0]

		http_req, err := http_parser.ParseRawHttp(line)

		if err != nil {
			log.Println(err)
			return
		}

		log.Println(http_req.Method)
		log.Println(http_req.Target)
		log.Println(http_req.Version)

		response_buffer := []byte("Hello")
		c.Write(response_buffer)
	}

}
