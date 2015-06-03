package packet

import (
	"bytes"
	"math/rand"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/funny/binary"
	"github.com/funny/link"
	"github.com/funny/unitest"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandBytes(n int) []byte {
	n = rand.Intn(n) + 1
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(rand.Intn(255))
	}
	return b
}

func Test_Packet(t *testing.T) {
	protocol := New(binary.SplitByUvarint, 1024, 1024, 1024)

	server, err := link.Listen("tcp", "0.0.0.0:0", protocol)
	unitest.NotError(t, err)

	go server.Serve(func(session *link.Session) {
		for {
			var msg RAW
			if err := session.Receive(&msg); err != nil {
				break
			}
			if err := session.Send(msg); err != nil {
				break
			}
		}
	})

	session, err := link.Dial("tcp", server.Listener().Addr().String(), protocol)
	unitest.NotError(t, err)
	for i := 0; i < 100000; i++ {
		p := RandBytes(1024)
		err = session.Send(RAW(p))
		unitest.NotError(t, err)

		var msg2 RAW
		err = session.Receive(&msg2)
		unitest.NotError(t, err)
		unitest.Pass(t, bytes.Equal(p, msg2))
	}
	session.Close()
	server.Stop()

	MakeSureSessionGoroutineExit(t)
}

func MakeSureSessionGoroutineExit(t *testing.T) {
	buff := new(bytes.Buffer)
	goroutines := pprof.Lookup("goroutine")

	if err := goroutines.WriteTo(buff, 2); err != nil {
		t.Fatalf("Dump goroutine failed: %v", err)
	}

	if n := bytes.Index(buff.Bytes(), []byte("asyncSendLoop")); n >= 0 {
		t.Log(buff.String())
		t.Fatalf("Some session goroutine running")
	}
}
