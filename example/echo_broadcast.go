package main

import (
	"flag"
	"time"

	"github.com/funny/link"
)

// This is broadcast server demo work with the echo_client.
// usage:
//     go run echo_broadcast.go
func main() {
	var addr string

	flag.StringVar(&addr, "addr", ":10010", "echo server address")
	flag.Parse()

	server, err := link.Serve("tcp://"+addr, link.Packet(link.Uint16BE, link.String()))
	if err != nil {
		panic(err)
	}
	println("server start:", server.Listener().Addr().String())

	channel := link.NewChannel()
	go func() {
		for range time.Tick(time.Second * 2) {
			now := time.Now().Format("2006-01-02 15:04:05")
			channel.Broadcast("from channel: " + now)
		}
	}()

	server.Loop(func(session *link.Session) {
		addr := session.Conn().RemoteAddr().String()
		println("client", addr, "connected")

		session.EnableAsyncSend(1024)
		channel.Join(session)

		for {
			var msg string
			if err := session.Receive(&msg); err != nil {
				break
			}
			println(addr, "say:", msg)
			channel.Broadcast("from " + addr + ": " + string(msg))
		}

		println("client", addr, "closed")
		channel.Exit(session)
	})
}
