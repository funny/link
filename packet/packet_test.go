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

func PacketTest(t *testing.T, callback func(*link.Session)) {
	protocol := New(binary.SplitByUvarint, 1024, 1024, 1024)

	server, err := link.Serve("tcp", "0.0.0.0:0", protocol)
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

	session, err := link.Connect("tcp", server.Listener().Addr().String(), protocol)
	unitest.NotError(t, err)
	callback(session)
	session.Close()
	server.Stop()

	MakeSureSessionGoroutineExit(t)
}
func Test_Packet(t *testing.T) {
	PacketTest(t, func(session *link.Session) {
		for i := 0; i < 100000; i++ {
			msg1 := RandBytes(1024)
			err := session.Send(RAW(msg1))
			unitest.NotError(t, err)

			var msg2 RAW
			err = session.Receive(&msg2)
			unitest.NotError(t, err)
			unitest.Pass(t, bytes.Equal(msg1, msg2))
		}
	})
}

type TestObject struct {
	X, Y, Z int
}

func Test_JSON(t *testing.T) {
	PacketTest(t, func(session *link.Session) {
		for i := 0; i < 50000; i++ {
			msg1 := TestObject{
				X: rand.Int(), Y: rand.Int(), Z: rand.Int(),
			}
			err := session.Send(JSON{msg1})
			unitest.NotError(t, err)

			var msg2 TestObject
			err = session.Receive(JSON{&msg2})
			unitest.NotError(t, err)
			unitest.Pass(t, msg1 == msg2)
		}
	})
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
