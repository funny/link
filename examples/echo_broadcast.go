package main

import (
	"fmt"
	"github.com/funny/link"
	"time"
)

// This is broadcast server demo work with the echo_client.
// usage:
//     go run echo_broadcast.go
func main() {
	server, err := link.Serve("tcp", "127.0.0.1:10010")
	if err != nil {
		panic(err)
	}

	channel := link.NewChannel()
	go func() {
		for {
			time.Sleep(time.Second * 2)
			// broadcast to server sessions
			server.Broadcast(Message("server say: " + time.Now().String()))
			// broadcast to channel sessions
			channel.Broadcast(Message("channel say: " + time.Now().String()))
		}
	}()

	println("server start")

	server.Serve(func(session *link.Session) {
		addr := session.Conn().RemoteAddr().String()
		println("client", addr, "connected")
		channel.Join(session, nil)

		for {
			var msg Message
			if err := session.Receive(&msg); err != nil {
				break
			}
			println(addr, "say:", string(msg))
			channel.Broadcast(msg)
		}

		println("client", addr, "closed")
		channel.Exit(session)
	})
}

type Message []byte

func (msg Message) Send(conn *link.Conn) error {
	conn.WritePacket([]byte(msg), link.SplitByUint16BE)
	return nil
}

func (msg *Message) Receive(conn *link.Conn) error {
	*msg = conn.ReadPacket(link.SplitByUint16BE)
	return nil
}
