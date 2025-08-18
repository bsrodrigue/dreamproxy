package main

import (
	"dreamproxy/http_parser"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const PROTOCOL string = "tcp4"
const PORT string = "8080"

func LoadFile(filepath string) ([]byte, error) {
	index_file, err := os.Open(filepath)

	if err != nil {
		log.Println(err)
		return []byte{}, err
	}

	index_content, err := io.ReadAll(index_file)

	if err != nil {
		return []byte{}, err
	}

	defer index_file.Close()

	return index_content, err
}

func main() {
	ln, err := net.Listen(PROTOCOL, fmt.Sprintf(":%s", PORT))

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s", fmt.Sprintf("listening on :%s", PORT))

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

		request_str := string(request_buffer[:n])

		http_req, err := http_parser.ParseRawHttp(request_str)

		if err != nil {
			log.Println(err)
			return
		}

		target := http_req.Target
		var res http_parser.HttpRes

		//TODO: Handle HTTP/0.9
		if http_req.Version == "0.9" {
		} else {
		}

		res = http_parser.HttpRes{
			Version:     http_parser.V0_9,
			ContentType: "text/html; charset=utf-8",
			Connection:  "close",
		}

		if target == "/" {
			target = "index"
		} else {
			target = strings.Split(target, "/")[1]
		}

		index_content, err := LoadFile(fmt.Sprintf("%s.html", target))

		if err != nil {
			log.Println(err)
			res.Status = http_parser.StatusNotFound
			res.Body = []byte("")
		} else {
			res.Status = http_parser.StatusOK
			res.Body = index_content
		}

		c.Write([]byte(res.ToStr()))
	}

}
