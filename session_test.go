package link

import (
	"bytes"
	"math/rand"
	"runtime/pprof"
	"sync"
	"testing"
	"time"

	"github.com/funny/binary"
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

type TestObject struct {
	X, Y, Z int
}

func (obj *TestObject) BinaryDecode(r *binary.Reader) error {
	obj.X = int(r.ReadVarint())
	obj.Y = int(r.ReadVarint())
	obj.Z = int(r.ReadVarint())
	return nil
}

func (obj *TestObject) BinaryEncode(w *binary.Writer) error {
	w.WriteVarint(int64(obj.X))
	w.WriteVarint(int64(obj.Y))
	w.WriteVarint(int64(obj.Z))
	return nil
}

func RandObject() TestObject {
	return TestObject{
		X: rand.Int(), Y: rand.Int(), Z: rand.Int(),
	}
}

func SessionTest(t *testing.T, protocol ClientProtocol, test func(*testing.T, *Session)) {
	server, err := Serve("tcp://0.0.0.0:0", Stream(Bytes()))
	unitest.NotError(t, err)
	addr := server.listener.Addr().String()

	serverWait := new(sync.WaitGroup)
	go server.Loop(func(session *Session) {
		serverWait.Add(1)
		Echo(session)
		serverWait.Done()
	})

	clientWait := new(sync.WaitGroup)
	for i := 0; i < 60; i++ {
		clientWait.Add(1)
		go func() {
			session, err := Connect("tcp://"+addr, protocol)
			unitest.NotError(t, err)
			test(t, session)
			session.Close()
			clientWait.Done()
		}()
	}
	clientWait.Wait()

	server.Stop()
	serverWait.Wait()

	MakeSureSessionGoroutineExit(t)
}

func Test_BytesStream(t *testing.T) {
	SessionTest(t, Stream(Bytes()), func(t *testing.T, session *Session) {
		for i := 0; i < 2000; i++ {
			msg1 := RandBytes(1024)
			err := session.Send(msg1)
			unitest.NotError(t, err)

			var msg2 = make([]byte, len(msg1))
			err = session.Receive(msg2)
			unitest.NotError(t, err)
			unitest.Pass(t, bytes.Equal(msg1, msg2))
		}
	})
}

func Test_BytesPacket(t *testing.T) {
	SessionTest(t, Packet(Uint16BE, Bytes()), func(t *testing.T, session *Session) {
		for i := 0; i < 2000; i++ {
			msg1 := RandBytes(1024)
			err := session.Send(msg1)
			unitest.NotError(t, err)

			var msg2 []byte
			err = session.Receive(&msg2)
			unitest.NotError(t, err)
			unitest.Pass(t, bytes.Equal(msg1, msg2))
		}
	})
}

func Test_StringPacket(t *testing.T) {
	SessionTest(t, Packet(Uint16BE, String()), func(t *testing.T, session *Session) {
		for i := 0; i < 2000; i++ {
			msg1 := string(RandBytes(1024))
			err := session.Send(msg1)
			unitest.NotError(t, err)

			var msg2 string
			err = session.Receive(&msg2)
			unitest.NotError(t, err)
			unitest.Pass(t, msg1 == msg2)
		}
	})
}

func ObjectTest(t *testing.T, session *Session) {
	for i := 0; i < 2000; i++ {
		msg1 := RandObject()
		err := session.Send(&msg1)
		unitest.NotError(t, err)

		var msg2 TestObject
		err = session.Receive(&msg2)
		unitest.NotError(t, err)
		unitest.Pass(t, msg1 == msg2)
	}
}

func Test_GobStream(t *testing.T) {
	SessionTest(t, Stream(Gob()), ObjectTest)
}

func Test_GobPacket(t *testing.T) {
	SessionTest(t, Packet(Uint16BE, Gob()), ObjectTest)
}

func Test_JsonStream(t *testing.T) {
	SessionTest(t, Stream(Json()), ObjectTest)
}

func Test_JsonPacket(t *testing.T) {
	SessionTest(t, Packet(Uint16BE, Json()), ObjectTest)
}

func Test_XmlStream(t *testing.T) {
	SessionTest(t, Stream(Xml()), ObjectTest)
}

func Test_XmlPacket(t *testing.T) {
	SessionTest(t, Packet(Uint16BE, Xml()), ObjectTest)
}

func Test_XmlSocket(t *testing.T) {
	SessionTest(t, Packet(Null, Xml()), ObjectTest)
}

func Test_SelfCodecStream(t *testing.T) {
	SessionTest(t, Stream(SelfCodec()), ObjectTest)
}

func Test_SelfCodecPacket(t *testing.T) {
	SessionTest(t, Packet(Uint16BE, SelfCodec()), ObjectTest)
}

func MakeSureSessionGoroutineExit(t *testing.T) {
	buff := new(bytes.Buffer)
	goroutines := pprof.Lookup("goroutine")

	if err := goroutines.WriteTo(buff, 2); err != nil {
		t.Fatalf("Dump goroutine failed: %v", err)
	}

	if n := bytes.Index(buff.Bytes(), []byte("link.HandlerFunc.Handle")); n >= 0 {
		t.Log(buff.String())
		t.Fatalf("Some handler goroutine running")
	}
}
