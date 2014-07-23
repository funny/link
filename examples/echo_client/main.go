package main

import "fmt"
import "encoding/binary"
import "github.com/funny/packnet"

// This is an echo client demo work with the echo_server.
// usage:
//     go run github.com/funny/examples/echo_client/main.go
func main() {
	protocol := packnet.NewFixProtocol(4, binary.BigEndian)

	client, err := packnet.Dial("tcp", "127.0.0.1:10010", protocol)
	if err != nil {
		panic(err)
	}

	client.OnMessage(func(session *packnet.Session, message []byte) {
		println("message:", string(message))
	})

	client.Start(func(session *packnet.Session) {
		println("closed")
	})

	for {
		var input string
		if _, err := fmt.Scanf("%s\n", &input); err != nil {
			break
		}
		client.Send(EchoMessage{input})
	}

	client.Close()

	println("bye")
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
