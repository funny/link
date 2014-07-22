package main

import "sync"
import "encoding/binary"
import "github.com/funny/packnet"

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(2)

	protocol := packnet.NewFixProtocol(4, binary.BigEndian)

	server, err := packnet.ListenAndServe("tcp", "127.0.0.1:10010", protocol)
	if err != nil {
		panic(err)
	}

	server.OnSessionStart(func(session *packnet.Session) {
		println("client from: ", session.RawConn().RemoteAddr().String())

		session.OnMessage(func(session *packnet.Session, message []byte) {
			println("message:", string(message))

			session.Send(EchoMessage{message})
		})

		wg.Done()
	})

	server.OnSessionClose(func(session *packnet.Session) {
		wg.Done()
	})

	server.Start()

	println("server start")

	wg.Wait()

	println("bye")
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
