package main

import (
	"flag"
	"github.com/funny/link"
	"io"
)

func main() {
	var addr string

	flag.StringVar(&addr, "addr", ":10010", "echo server address")
	flag.Parse()

	server, err := link.Serve("tcp://"+addr, link.Stream(link.Bytes()))
	if err != nil {
		panic(err)
	}
	println("server start:", server.Listener().Addr().String())
	server.Loop(func(session *link.Session) {
		c := session.Conn().(*link.StreamConn)
		io.Copy(c.Conn, c.Conn)
	})
}
