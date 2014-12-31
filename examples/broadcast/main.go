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
			time.Sleep(time.Second * 2)
			// broadcast to server sessions
			link.Broadcast(server, link.String("server say: "+time.Now().String()))
			// broadcast to channel sessions
			link.Broadcast(channel, link.String("channel say: "+time.Now().String()))
		}
	}()

	println("server start")

	server.Handle(func(session *link.Session) {
		println("client", session.Conn().RemoteAddr().String(), "in")
		channel.Join(session, nil)

		session.Handle(func(msg *link.InBuffer) {
			link.Broadcast(channel, link.String(
				"client "+session.Conn().RemoteAddr().String()+" say: "+string(msg.Data),
			))
		})

		println("client", session.Conn().RemoteAddr().String(), "close")
		channel.Exit(session)
	})
}
