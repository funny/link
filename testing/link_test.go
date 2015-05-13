package testing

import (
	"bytes"
	"github.com/funny/link"
	"github.com/funny/link/protocol/fixhead"
	"github.com/funny/unitest"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"testing"
)

func Test_Server(t *testing.T) {
	var (
		data    = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		message = link.Bytes(data)

		sessionStart   sync.WaitGroup
		sessionClose   sync.WaitGroup
		sessionRequest sync.WaitGroup

		sessionStartCount   int32
		sessionRequestCount int32
		sessionCloseCount   int32
		messageMatchFailed  bool
	)

	server, err0 := link.Listen("tcp", "0.0.0.0:0", fixhead.Uint32BE, nil)
	unitest.NotError(t, err0)

	addr := server.Listener().Addr().String()

	go server.Serve(func(session *link.Session) {
		atomic.AddInt32(&sessionStartCount, 1)
		sessionStart.Done()

		decoder := func(buffer *link.Buffer) (link.Request, error) {
			if !bytes.Equal(buffer.Data, data) {
				messageMatchFailed = true
			}
			atomic.AddInt32(&sessionRequestCount, 1)
			sessionRequest.Done()
			return nil, nil
		}

		session.Process(link.DecodeFunc(decoder))

		atomic.AddInt32(&sessionCloseCount, 1)
		sessionClose.Done()
	})

	// test session start
	sessionStart.Add(1)
	sessionClose.Add(1)
	client1, err1 := link.Dial("tcp", addr, fixhead.Uint32BE, nil)
	unitest.NotError(t, err1)

	sessionStart.Add(1)
	sessionClose.Add(1)
	client2, err2 := link.Dial("tcp", addr, fixhead.Uint32BE, nil)
	unitest.NotError(t, err2)

	t.Log("check session start")
	sessionStart.Wait()
	unitest.Pass(t, sessionStartCount == 2)

	// test session request
	sessionRequest.Add(1)
	unitest.NotError(t, client1.Send(message))

	sessionRequest.Add(1)
	unitest.NotError(t, client2.Send(message))

	sessionRequest.Add(1)
	unitest.NotError(t, client1.Send(message))

	sessionRequest.Add(1)
	unitest.NotError(t, client2.Send(message))

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
