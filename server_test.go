package link

import (
	"bytes"
	"encoding/binary"
	"runtime/pprof"
	"sync"
	"testing"
)

type TestMessage struct {
	Message []byte
}

func (msg TestMessage) RecommendPacketSize() uint {
	return uint(len(msg.Message))
}

func (msg TestMessage) AppendToPacket(packet []byte) []byte {
	return append(packet, msg.Message...)
}

func Test_Server(t *testing.T) {
	proto := PacketN(4, binary.BigEndian)

	server, err0 := Listen("tcp", "0.0.0.0:0", proto)
	if err0 != nil {
		t.Fatalf("Setup server failed, Error = %v", err0)
	}

	var (
		addr    = server.Listener().Addr().String()
		message = TestMessage{[]byte{0, 1, 2, 3, 4, 45, 6, 7, 8, 9}}

		sessionStart       sync.WaitGroup
		sessionClose       sync.WaitGroup
		sessionRequest     sync.WaitGroup
		messageMatchFailed bool
	)

	go server.AcceptLoop(func(session *Session) {
		sessionStart.Done()

		session.ReadLoop(func(msg []byte) {
			sessionRequest.Done()

			if !bytes.Equal(msg, message.Message) {
				messageMatchFailed = true
			}
		})

		sessionClose.Done()
	})

	// test session start
	sessionStart.Add(1)
	client1, err1 := Dial("tcp", addr, proto)
	if err1 != nil {
		t.Fatal("Create client1 failed, Error = %v", err1)
	}

	sessionStart.Add(1)
	client2, err2 := Dial("tcp", addr, proto)
	if err2 != nil {
		t.Fatal("Create client2 failed, Error = %v", err2)
	}

	t.Log("check session start")
	sessionStart.Wait()

	// test session request
	sessionRequest.Add(1)
	if err := client1.Send(message); err != nil {
		t.Fatal("Send message failed, Error = %v", err)
	}

	sessionRequest.Add(1)
	if err := client2.Send(message); err != nil {
		t.Fatal("Send message failed, Error = %v", err)
	}

	sessionRequest.Add(1)
	if err := client1.Send(message); err != nil {
		t.Fatal("Send message failed, Error = %v", err)
	}

	sessionRequest.Add(1)
	if err := client2.Send(message); err != nil {
		t.Fatal("Send message failed, Error = %v", err)
	}

	t.Log("check session request")
	sessionRequest.Wait()

	if messageMatchFailed {
		t.Fatal("Message match failed")
	}

	// test session close
	sessionClose.Add(1)
	client1.Close(nil)

	sessionClose.Add(1)
	client2.Close(nil)

	t.Log("check session close")
	sessionClose.Wait()

	MakeSureSessionGoroutineExit(t)
}

func MakeSureSessionGoroutineExit(t *testing.T) {
	buff := new(bytes.Buffer)
	goroutines := pprof.Lookup("goroutine")

	if err := goroutines.WriteTo(buff, 2); err != nil {
		t.Fatalf("Dump goroutine failed: %v", err)
	}

	if n := bytes.Index(buff.Bytes(), []byte("sendLoop")); n >= 0 {
		t.Log(buff.String())
		t.Fatalf("Some session goroutine running")
	}
}
