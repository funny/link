package main

import (
	"flag"
	"fmt"
	"github.com/funny/link"
)

var (
	benchmark  = flag.Bool("bench", false, "is for benchmark, will disable print")
	buffersize = flag.Int("buffer", 1024, "session read buffer size")
)

func log(v ...interface{}) {
	if !*benchmark {
		fmt.Println(v...)
	}
}

// This is an echo server demo work with the echo_client.
// usage:
//     go run echo_server/main.go
func main() {
	flag.Parse()

	link.DefaultConnBufferSize = *buffersize

	server, err := link.Listen("tcp", "127.0.0.1:10010")
	if err != nil {
		panic(err)
	}

	println("server start:", server.Listener().Addr().String())

	server.Serve(func(session *link.Session) {
		log("client", session.Conn().RemoteAddr().String(), "in")

		session.Handle(func(msg *link.InBuffer) {
			log("client", session.Conn().RemoteAddr().String(), "say:", string(msg.Data))
			session.Send(link.Bytes(msg.Data))
		})

		log("client", session.Conn().RemoteAddr().String(), "close")
	})
}
