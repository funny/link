package main

import "time"
import "encoding/binary"
import "github.com/funny/packnet"

// This is broadcast server demo work with the echo_client.
// usage:
//     go run github.com/funny/examples/broadcast/main.go
func main() {
	protocol := packnet.NewFixProtocol(4, binary.BigEndian)

	server, err := packnet.ListenAndServe("tcp", "127.0.0.1:10010", protocol)
	if err != nil {
		panic(err)
	}

	channel := server.NewChannel()

	server.OnSessionStart(func(session *packnet.Session) {
		println("client", session.RawConn().RemoteAddr().String(), "in")

		channel.Join(session, nil)
	})

	server.OnSessionClose(func(session *packnet.Session) {
		println("client", session.RawConn().RemoteAddr().String(), "close")

		channel.Exit(session)
	})

	server.Start()

	go func() {
		for {
			time.Sleep(time.Second)

			channel.Broadcast(EchoMessage{time.Now().String()})
		}
	}()

	println("server start")

	<-make(chan int)
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
