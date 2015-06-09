package main

import (
	"flag"
	"fmt"

	"github.com/funny/binary"
	"github.com/funny/link"
	"github.com/funny/link/packet"
	_ "github.com/funny/unitest"
)

var (
	addr  = flag.String("addr", ":10010", "echo server address")
	bench = flag.Bool("bench", false, "is benchmark server")
)

func main() {
	flag.Parse()

	server, err := link.Serve("tcp", *addr, packet.New(
		binary.SplitByUint16BE, 1024, 1024, 1024,
	))
	if err != nil {
		panic(err)
	}
	println("server start:", server.Listener().Addr().String())

	server.Serve(func(session *link.Session) {
		addr := session.Conn().RemoteAddr().String()
		log(addr, "connected")
		for {
			var msg packet.RAW
			if err := session.Receive(&msg); err != nil {
				break
			}
			log(addr, "say:", string(msg))
			if err = session.Send(msg); err != nil {
				break
			}
		}
		log(addr, "closed")
	})
}

func log(v ...interface{}) {
	if !*bench {
		fmt.Println(v...)
	}
}
