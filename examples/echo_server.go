package main

import (
	"flag"
	"fmt"
	"github.com/funny/link"
	"github.com/funny/link/protocol/fixhead"
)

var (
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

	link.DefaultConfig.RequestBufferSize = 1024
	link.DefaultConfig.ResponseBufferSize = 1024
	link.DefaultConfig.SendChanSize = *asyncChan
	pool := link.NewMemPool(10, 1, 10)

	server, err := link.Listen("tcp", "127.0.0.1:10010", fixhead.Uint16BE, pool)
	if err != nil {
		panic(err)
	}

	println("server start:", server.Listener().Addr().String())

	server.Serve(func(session *link.Session) {
		log("client", session.Conn().RemoteAddr().String(), "in")

		session.Process(link.DecodeFunc(func(buf *link.Buffer) (link.Request, error) {
			log("client", session.Conn().RemoteAddr().String(), "say:", string(buf.Data))
			return EchoRequest(buf.ReadBytes(buf.Length())), nil
		}))

		log("client", session.Conn().RemoteAddr().String(), "close")
	})
}

type EchoRequest []byte

func (req EchoRequest) Process(session *link.Session) error {
	if *asyncChan == 0 {
		return session.Send(link.Bytes(req))
	} else {
		session.AsyncSend(link.Bytes(req))
	}
	return nil
}
