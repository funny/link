package main

import (
	"github.com/funny/link"
	"time"
)

// This is broadcast server demo work with the echo_client.
// usage:
//     go run broadcast/main.go
func main() {
	protocol := link.PacketN(2, link.BigEndian)

	server, err := link.Listen("tcp", "127.0.0.1:10010", protocol)
	if err != nil {
		panic(err)
	}

	channel := link.NewChannel(server.Protocol())
	go func() {
		for {
			time.Sleep(time.Second)
			channel.Broadcast(link.Binary(time.Now().String()))
		}
	}()

	println("server start")

	server.AcceptLoop(func(session *link.Session) {
		println("client", session.Conn().RemoteAddr().String(), "in")
		channel.Join(session, nil)

		session.ReadLoop(func(msg link.InBuffer) {
			channel.Broadcast(link.Binary(
				session.Conn().RemoteAddr().String() + " say: " + string(msg.Get()),
			))
		})

		println("client", session.Conn().RemoteAddr().String(), "close")
		channel.Exit(session)
	})
}
