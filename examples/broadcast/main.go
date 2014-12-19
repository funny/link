package main

import (
	"github.com/funny/link"
	"time"
)

// This is broadcast server demo work with the echo_client.
// usage:
//     go run broadcast/main.go
func main() {
	protocol := link.PacketN(2, link.BigEndian, link.DefaultBuffer)

	server, err := link.Listen("tcp", "127.0.0.1:10010", protocol)
	if err != nil {
		panic(err)
	}

	channel := link.NewChannel(server.Protocol())
	go func() {
		for {
			time.Sleep(time.Second)
			link.Broadcast(channel, link.Binary(time.Now().String()))
		}
	}()

	println("server start")

	server.Handle(func(session *link.Session) {
		println("client", session.Conn().RemoteAddr().String(), "in")
		channel.Join(session, nil)

		session.Handle(func(msg link.Buffer) {
			link.Broadcast(channel, link.Binary(
				session.Conn().RemoteAddr().String()+" say: "+string(msg.Data()),
			))
		})

		println("client", session.Conn().RemoteAddr().String(), "close")
		channel.Exit(session)
	})
}
