package packnet

import "time"
import "bytes"
import "testing"
import "encoding/binary"

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
	proto := NewFixProtocol(4, binary.BigEndian)

	server, err0 := ListenAndServe("tcp", "0.0.0.0:0", proto)
	if err0 != nil {
		t.Fatalf("Setup server failed, Error = %v", err0)
	}
	addr := server.Listener().Addr().String()
	t.Logf("Server: %v", addr)

	var (
		sessionStartCount  int
		sessionCloseCount  int
		sessionMatchFailed bool
		messageCount       int
		messageMatchFailed bool
		message            = &TestMessage{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}
	)

	server.SetSessionStartHook(func(session1 *Session) {
		t.Log("Session start")
		sessionStartCount += 1
		session1.SetRequestHandlerFunc(func(session2 *Session, msg []byte) {
			messageCount += 1
			if session1 != session2 {
				sessionMatchFailed = true
			}
			if !bytes.Equal(msg, message.Message) {
				messageMatchFailed = true
				t.Logf("message: %v", msg)
			}
		})
	})

	server.SetSessionCloseHook(func(session *Session) {
		t.Log("Session close")
		sessionCloseCount += 1
	})

	server.Start()

	client1, err1 := DialTimeout("tcp", addr, time.Second, proto, 1, 1024)
	if err1 != nil {
		t.Fatal("Create client1 failed, Error = %v", err1)
	}
	client1.Start()

	client2, err2 := DialTimeout("tcp", addr, time.Second, proto, 2, 1024)
	if err2 != nil {
		t.Fatal("Create client2 failed, Error = %v", err2)
	}
	client2.Start()

	t.Log("Send 1")
	if err := client1.Send(message); err != nil {
		t.Fatal("Send message failed, Error = %v", err)
	}

	t.Log("Send 2")
	if err := client2.Send(message); err != nil {
		t.Fatal("Send message failed, Error = %v", err)
	}

	t.Log("Send 3")
	if err := client1.Send(message); err != nil {
		t.Fatal("Send message failed, Error = %v", err)
	}

	t.Log("Send 4")
	if err := client2.Send(message); err != nil {
		t.Fatal("Send message failed, Error = %v", err)
	}

	// close by manual
	t.Log("Close client1")
	client1.Close()
	time.Sleep(time.Second)

	t.Log("Stop server")
	server.Stop()

	if sessionStartCount != 2 {
		t.Fatal("Session start count not match")
	}

	if sessionCloseCount != 2 {
		t.Fatal("Session close count not match")
	}

	if sessionMatchFailed {
		t.Fatal("Session match failed")
	}

	if messageCount != 4 {
		t.Logf("Message count: %d", messageCount)
		t.Fatal("Message count not match")
	}

	if messageMatchFailed {
		t.Fatal("Message match failed")
	}
}
