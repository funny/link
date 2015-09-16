package main

import (
	"flag"
	"io"

	"github.com/funny/link"
	"github.com/funny/link/example/codec"
)

func main() {
	var addr string

	flag.StringVar(&addr, "addr", ":10010", "echo server address")
	flag.Parse()

	server, err := link.Serve("tcp", addr, codec.Bytes(codec.Uint16BE))
	if err != nil {
		panic(err)
	}
	println("server start:", server.Listener().Addr().String())
	for {
		session, err := server.Accept()
		if err != nil {
			break
		}
		go io.Copy(session.Conn(), session.Conn())
	}
}
