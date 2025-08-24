package dream

import (
	"dreamproxy/config"
	"fmt"
	"log"
	"net"
)

const PROTOCOL string = "tcp4"

type DreamContext struct {
	Port    string
	Servers []config.Server
}

func (ctxt *DreamContext) RunDreamContext() {
	ln, err := net.Listen(PROTOCOL, fmt.Sprintf(":%s", ctxt.Port))

	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()

	log.Printf("%s", fmt.Sprintf("listening on :%s", ctxt.Port))

	for {

		connection, err := ln.Accept()

		if err != nil {
			log.Println(err)
			continue
		}

		client_session := NewClientSession(connection)

		go client_session.HandleConnection(ctxt.Servers)
	}
}

func NewDreamContext(port string, servers []config.Server) DreamContext {

	return DreamContext{
		Port:    port,
		Servers: servers,
	}
}
