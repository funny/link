package main

import (
	"github.com/funny/link"
	"github.com/funny/link/protocol/fixhead"
	"time"
)

// This is broadcast server demo work with the echo_client.
// usage:
//     go run echo_broadcast.go
func main() {
	server, err := link.Listen("tcp", "127.0.0.1:10010", fixhead.Uint16BE, nil)
	if err != nil {
		panic(err)
	}

	channel := link.NewChannel(fixhead.Uint16BE, nil)
	go func() {
		for {
			time.Sleep(time.Second * 2)
			// broadcast to server sessions
			server.Broadcast(link.String("server say: " + time.Now().String()))
			// broadcast to channel sessions
			channel.Broadcast(link.String("channel say: " + time.Now().String()))
		}
	}()

	println("server start")

	server.Serve(func(session *link.Session) {
		println("client", session.Conn().RemoteAddr().String(), "in")
		channel.Join(session, nil)

		session.Process(link.DecodeFunc(func(buf *link.Buffer) (link.Request, error) {
			channel.Broadcast(link.String(
				"client " + session.Conn().RemoteAddr().String() + " say: " + string(buf.Data),
			))
			return nil, nil
		}))

		println("client", session.Conn().RemoteAddr().String(), "close")
		channel.Exit(session)
	})
}
