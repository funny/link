package main

import (
	"encoding/binary"
	"flag"
	"github.com/funny/link"
)

var (
	benchmark = flag.Bool("bench", false, "is for benchmark, will disable print")
)

// This is an echo server demo work with the echo_client.
// usage:
//     go run github.com/funny/examples/echo_server/main.go
func main() {
	flag.Parse()

	protocol := link.NewFixProtocol(2, binary.BigEndian)

	server, err := link.ListenAndServe("tcp", "127.0.0.1:10010", protocol)
	if err != nil {
		panic(err)
	}

	println("server start:", server.Listener().Addr().String())

	server.Handle(func(session *link.Session) {
		if !*benchmark {
			println("client", session.Conn().RemoteAddr().String(), "in")
		}

		session.OnMessage(func(session *link.Session, message []byte) {
			if !*benchmark {
				println("client", session.Conn().RemoteAddr().String(), "say:", string(message))
				session.Send(EchoMessage{message})
			} else {
				session.SyncSend(EchoMessage{message})
			}
		})

		session.OnClose(func(session *link.Session, reason error) {
			if !*benchmark {
				println("client", session.Conn().RemoteAddr().String(), "close, ", reason)
			}
		})

		session.Start()
	})
}

type EchoMessage struct {
	Content []byte
}

func (msg EchoMessage) RecommendPacketSize() uint {
	return uint(len(msg.Content))
}

func (msg EchoMessage) AppendToPacket(packet []byte) []byte {
	return append(packet, msg.Content...)
}
