package main

import (
	"io"
	"log"
	"net"

	"github.com/husio/apps/qux/qux"
)

func main() {
	ln, err := net.Listen("tcp", "localhost:12345")
	if err != nil {
		log.Fatalf("cannot start server: %s", err)
	}
	defer ln.Close()

	s := qux.NewServer()

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("cannot accept client: %s", err)
			continue
		}
		go handleClient(c, s)
	}
}

func handleClient(c net.Conn, s *qux.Server) {
	defer c.Close()
	log.Printf(">> client connected: %v", c)
	if err := s.Serve(c); err != nil || err != io.EOF {
		log.Printf("server error: %s", err)
	}
	log.Printf("<< client disconnected: %v", c)
}
