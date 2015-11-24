package main

import (
	"flag"
	"io"
	"net"
)

func main() {
	var addr string

	flag.StringVar(&addr, "addr", ":10010", "echo server address")
	flag.Parse()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	println("server start:", listener.Addr().String())
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		go io.Copy(conn, conn)
	}
}
