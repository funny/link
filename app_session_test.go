package link

import (
	"bytes"
	"github.com/funny/unitest"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"testing"
)

type SessionTestMessage []byte

func (msg SessionTestMessage) Send(conn *Conn) error {
	conn.WritePacket(msg, SplitByUint16BE)
	return nil
}

func (msg *SessionTestMessage) Receive(conn *Conn) error {
	*msg = conn.ReadPacket(SplitByUint16BE)
	return nil
}

func Test_Server(t *testing.T) {
	var (
		out = SessionTestMessage{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

		sessionStart   sync.WaitGroup
		sessionClose   sync.WaitGroup
		sessionRequest sync.WaitGroup

		sessionStartCount   int32
		sessionRequestCount int32
		sessionCloseCount   int32
		messageMatchFailed  bool
	)

	server, err0 := Serve("tcp", "0.0.0.0:0")
	unitest.NotError(t, err0)

	addr := server.Listener().Addr().String()

	go server.Serve(func(session *Session) {
		atomic.AddInt32(&sessionStartCount, 1)
		sessionStart.Done()

		for {
			var in SessionTestMessage
			if err := session.Receive(&in); err != nil {
				break
			}
			if !bytes.Equal(out, in) {
				messageMatchFailed = true
			}
			atomic.AddInt32(&sessionRequestCount, 1)
			sessionRequest.Done()
		}

		atomic.AddInt32(&sessionCloseCount, 1)
		sessionClose.Done()
	})

	// test session start
	sessionStart.Add(1)
	sessionClose.Add(1)
	client1, err1 := Connect("tcp", addr)
	unitest.NotError(t, err1)

	sessionStart.Add(1)
	sessionClose.Add(1)
	client2, err2 := Connect("tcp", addr)
	unitest.NotError(t, err2)

	t.Log("check session start")
	sessionStart.Wait()
	unitest.Pass(t, sessionStartCount == 2)

	// test session request
	sessionRequest.Add(1)
	unitest.NotError(t, client1.Send(out))

	sessionRequest.Add(1)
	unitest.NotError(t, client2.Send(out))

	sessionRequest.Add(1)
	unitest.NotError(t, client1.Send(out))

	sessionRequest.Add(1)
	unitest.NotError(t, client2.Send(out))

	t.Log("check session request")
	sessionRequest.Wait()

	unitest.Pass(t, sessionRequestCount == 4)
	unitest.Pass(t, messageMatchFailed == false)

	// test session close
	client1.Close()
	client2.Close()

	t.Log("check session close")
	sessionClose.Wait()
	unitest.Pass(t, sessionCloseCount == 2)

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
