package main

import (
	"encoding/binary"
	"github.com/funny/link"
)

// This is an echo server demo work with the echo_client.
// usage:
//     go run github.com/funny/examples/echo_server/main.go
func main() {
	protocol := link.NewFixProtocol(4, binary.BigEndian)

	server, err := link.ListenAndServe("tcp", "127.0.0.1:10010", protocol)
	if err != nil {
		panic(err)
	}

	println("server start")

	server.Handle(func(session *link.Session) {
		println("client", session.RawConn().RemoteAddr().String(), "in")

		session.OnMessage(func(session *link.Session, message []byte) {
			println("client", session.RawConn().RemoteAddr().String(), "say:", string(message))

			session.Send(EchoMessage{message})
		})

		session.OnClose(func(session *link.Session) {
			println("client", session.RawConn().RemoteAddr().String(), "close")
		})
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
