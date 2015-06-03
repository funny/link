package main

import (
	"flag"
	"time"

	"github.com/funny/binary"
	"github.com/funny/link"
	"github.com/funny/link/packet"
)

var (
	addr = flag.String("addr", ":10010", "echo server address")
)

// This is broadcast server demo work with the echo_client.
// usage:
//     go run echo_broadcast.go
func main() {
	flag.Parse()

	protocol := packet.New(
		binary.SplitByUint16BE, 1024, 1024, 1024,
	)

	server, err := link.Listen("tcp", *addr, protocol)
	if err != nil {
		panic(err)
	}
	println("server start:", server.Listener().Addr().String())

	channel := link.NewChannel(protocol)
	go func() {
		for {
			time.Sleep(time.Second * 2)
			// broadcast to server sessions
			server.Broadcast(packet.RAW("server say: " + time.Now().String()))
			// broadcast to channel sessions
			channel.Broadcast(packet.RAW("channel say: " + time.Now().String()))
		}
	}()

	server.Serve(func(session *link.Session) {
		addr := session.Conn().RemoteAddr().String()
		println("client", addr, "connected")
		channel.Join(session, nil)

		for {
			var msg packet.RAW
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
