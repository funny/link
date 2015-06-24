package main

import (
	"flag"
	"github.com/funny/link"
)

func main() {
	var addr string

	flag.StringVar(&addr, "addr", ":10010", "echo server address")
	flag.Parse()

	server, err := link.Serve("tcp://"+addr, link.Stream(), nil)
	if err != nil {
		panic(err)
	}
	println("server start:", server.Listener().Addr().String())
	server.Loop(link.Echo)
}
