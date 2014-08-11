package main

import (
	"encoding/binary"
	"github.com/funny/link"
	"time"
)

// This is broadcast server demo work with the echo_client.
// usage:
//     go run github.com/funny/examples/broadcast/main.go
func main() {
	protocol := link.NewFixProtocol(4, binary.BigEndian)

	server, err := link.ListenAndServe("tcp", "127.0.0.1:10010", protocol)
	if err != nil {
		panic(err)
	}

	channel := server.NewChannel()
	go func() {
		for {
			time.Sleep(time.Second)
			channel.Broadcast(EchoMessage{time.Now().String()})
		}
	}()

	println("server start")

	server.Handle(func(session *link.Session) {
		println("client", session.Conn().RemoteAddr().String(), "in")
		channel.Join(session, nil)

		session.OnClose(func(session *link.Session, reason error) {
			println("client", session.Conn().RemoteAddr().String(), "close, ", reason)
			channel.Exit(session)
		})

		session.Start()
	})
}

type EchoMessage struct {
	Content string
}

func (msg EchoMessage) RecommendPacketSize() uint {
	return uint(len(msg.Content))
}

func (msg EchoMessage) AppendToPacket(packet []byte) []byte {
	return append(packet, msg.Content...)
}
