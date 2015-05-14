package main

import (
	"flag"
	"fmt"
	"github.com/funny/link"
	"github.com/funny/link/protocol/fixhead"
)

var (
	addr      = flag.String("addr", "127.0.0.1:10010", "server address")
	benchmark = flag.Bool("bench", false, "is for benchmark, will disable print")
	asyncChan = flag.Int("async", 10000, "async send chan size, 0 == sync send")
)

func log(v ...interface{}) {
	if !*benchmark {
		fmt.Println(v...)
	}
}

// This is an echo server demo work with the echo_client.
// usage:
//     go run main.go
//     go run main.go -bench
//     go run main.go -bench -sync
func main() {
	flag.Parse()

	link.DefaultConfig.InBufferSize = 1024
	link.DefaultConfig.OutBufferSize = 1024
	link.DefaultConfig.SendChanSize = *asyncChan
	pool := link.NewMemPool(10, 1, 10)

	server, err := link.Listen("tcp", *addr, fixhead.Uint16BE, pool)
	if err != nil {
		panic(err)
	}

	println("server start:", server.Listener().Addr().String())

	server.Serve(func(session *link.Session) {
		log("client", session.Conn().RemoteAddr().String(), "in")

		session.Process(func(buf *link.Buffer) error {
			log("client", session.Conn().RemoteAddr().String(), "say:", string(buf.Data))
			if *asyncChan == 0 {
				return session.Send(link.Bytes(buf.Data))
			} else {
				session.AsyncSend(link.Bytes(buf.Data))
			}
			return nil
		})

		log("client", session.Conn().RemoteAddr().String(), "close")
	})
}
