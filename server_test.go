package link

import (
	"bytes"
	"encoding/binary"
	"runtime/pprof"
	"sync/atomic"
	"testing"
	"time"
)

type TestMessage struct {
	Message []byte
}

func (msg *TestMessage) RecommendPacketSize() uint {
	return uint(len(msg.Message))
}

func (msg *TestMessage) AppendToPacket(packet []byte) []byte {
	return append(packet, msg.Message...)
}

func Test_Server(t *testing.T) {
	proto := PacketN(4, binary.BigEndian)

	server, err0 := Listen("tcp", "0.0.0.0:0", proto)
	if err0 != nil {
		t.Fatalf("Setup server failed, Error = %v", err0)
	}
	addr := server.Listener().Addr().String()

	var (
		sessionStartCount  int32
		sessionCloseCount  int32
		messageCount       int32
		messageMatchFailed bool
		message            = &TestMessage{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}
		serverStopChan     = make(chan int)
	)

	go func() {
		server.AcceptLoop(func(session *Session) {
			atomic.AddInt32(&sessionStartCount, 1)

			session.ReadLoop(func(msg []byte) {
				atomic.AddInt32(&messageCount, 1)

				if !bytes.Equal(msg, message.Message) {
					messageMatchFailed = true
				}
			})

			atomic.AddInt32(&sessionCloseCount, 1)
		})
		server.Stop()
		close(serverStopChan)
	}()

	client1, err1 := Dial("tcp", addr, proto)
	if err1 != nil {
		t.Fatal("Create client1 failed, Error = %v", err1)
	}

	client2, err2 := Dial("tcp", addr, proto)
	if err2 != nil {
		t.Fatal("Create client2 failed, Error = %v", err2)
	}

	if err := client1.SyncSend(message); err != nil {
		t.Fatal("Send message failed, Error = %v", err)
	}

	if err := client2.SyncSend(message); err != nil {
		t.Fatal("Send message failed, Error = %v", err)
	}

	if err := client1.SyncSend(message); err != nil {
		t.Fatal("Send message failed, Error = %v", err)
	}

	if err := client2.SyncSend(message); err != nil {
		t.Fatal("Send message failed, Error = %v", err)
	}

	// close by manual
	client1.Close(nil)
	time.Sleep(time.Second)

	server.Stop()
	<-serverStopChan

	if sessionStartCount != 2 {
		t.Fatal("Session start count not match")
	}

	if sessionCloseCount != 2 {
		t.Fatal("Session close count not match")
	}

	if messageCount != 4 {
		t.Logf("Message count: %d", messageCount)
		t.Fatal("Message count not match")
	}

	if messageMatchFailed {
		t.Fatal("Message match failed")
	}

	MakeSureSessionGoroutineExit(t)
}

func MakeSureSessionGoroutineExit(t *testing.T) {
	buff := new(bytes.Buffer)
	goroutines := pprof.Lookup("goroutine")

	if err := goroutines.WriteTo(buff, 2); err != nil {
		t.Fatalf("Dump goroutine failed: %v", err)
	}

	if n := bytes.Index(buff.Bytes(), []byte("writeLoop")); n >= 0 {
		t.Log(buff.String())
		t.Fatalf("Some session goroutine running")
	}

	if n := bytes.Index(buff.Bytes(), []byte("Process")); n >= 0 {
		t.Log(buff.String())
		t.Fatalf("Some session goroutine running")
	}
}
