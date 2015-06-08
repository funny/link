package stream

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

type TestMessage []byte

func (msg *TestMessage) Unmarshal(r *binary.Reader) error {
	*msg = r.ReadPacket(binary.SplitByUvarint)
	return nil
}

func (msg TestMessage) Marshal(w *binary.Writer) error {
	w.WritePacket(msg, binary.SplitByUvarint)
	return nil
}

func StreamTest(t *testing.T, callback func(*link.Session)) {
	protocol := New(1024, 1024, 1024)

	server, err := link.Serve("tcp", "0.0.0.0:0", protocol)
	unitest.NotError(t, err)
	addr := server.Listener().Addr().String()

	go server.Serve(func(session *link.Session) {
		for {
			var msg TestMessage
			if err := session.Receive(&msg); err != nil {
				break
			}
			if err := session.Send(msg); err != nil {
				break
			}
		}
	})

	session, err := link.Connect("tcp", addr, protocol)
	unitest.NotError(t, err)
	callback(session)
	session.Close()
	server.Stop()

	MakeSureSessionGoroutineExit(t)
}

func Test_Stream(t *testing.T) {
	StreamTest(t, func(session *link.Session) {
		for i := 0; i < 100000; i++ {
			msg1 := RandBytes(1024)
			err := session.Send(TestMessage(msg1))
			unitest.NotError(t, err)

			var msg2 TestMessage
			err = session.Receive(&msg2)
			unitest.NotError(t, err)
			unitest.Pass(t, bytes.Equal(msg1, msg2))
		}
	})
}

type TestObject struct {
	X, Y, Z int
}

func Test_GOB(t *testing.T) {
	StreamTest(t, func(session *link.Session) {
		for i := 0; i < 50000; i++ {
			msg1 := TestObject{
				X: rand.Int(), Y: rand.Int(), Z: rand.Int(),
			}
			err := session.Send(GOB{msg1})
			unitest.NotError(t, err)

			var msg2 TestObject
			err = session.Receive(GOB{&msg2})
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
