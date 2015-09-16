package codec

import (
	"bytes"
	"io"
	"math/rand"
	"runtime/pprof"
	"sync"
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

type TestObject struct {
	X, Y, Z int
}

func (obj *TestObject) SelfDecode(r *binary.Reader) error {
	obj.X = int(r.ReadVarint())
	obj.Y = int(r.ReadVarint())
	obj.Z = int(r.ReadVarint())
	return nil
}

func (obj *TestObject) SelfEncode(w *binary.Writer) error {
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

func SessionTest(t *testing.T, codecType link.CodecType, test func(*testing.T, *link.Session)) {
	server, err := link.Serve("tcp", "0.0.0.0:0", Bytes(Uint16BE))
	unitest.NotError(t, err)
	addr := server.Listener().Addr().String()

	serverWait := new(sync.WaitGroup)
	go func() {
		for {
			session, err := server.Accept()
			if err != nil {
				break
			}
			serverWait.Add(1)
			go func() {
				io.Copy(session.Conn(), session.Conn())
				serverWait.Done()
			}()
		}
	}()

	clientWait := new(sync.WaitGroup)
	for i := 0; i < 60; i++ {
		clientWait.Add(1)
		go func() {
			session, err := link.Connect("tcp", addr, codecType)
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

func BytesTest(t *testing.T, session *link.Session) {
	for i := 0; i < 2000; i++ {
		msg1 := RandBytes(512)
		err := session.Send(msg1)
		unitest.NotError(t, err)

		var msg2 []byte
		err = session.Receive(&msg2)
		unitest.NotError(t, err)
		unitest.Pass(t, bytes.Equal(msg1, msg2))
	}
}

func Test_Bytes(t *testing.T) {
	SessionTest(t, Bytes(Uint16BE), BytesTest)
}

func Test_Bufio_Bytes(t *testing.T) {
	SessionTest(t, Bufio(Bytes(Uint16BE)), BytesTest)
}

func Test_Packet_Bytes(t *testing.T) {
	SessionTest(t, Packet(Uint16BE, Bytes(Uint16BE)), BytesTest)
}

func Test_Bufio_Packet_Bytes(t *testing.T) {
	SessionTest(t, Bufio(Packet(Uint16BE, Bytes(Uint16BE))), BytesTest)
}

func StringTest(t *testing.T, session *link.Session) {
	for i := 0; i < 2000; i++ {
		msg1 := string(RandBytes(512))
		err := session.Send(msg1)
		unitest.NotError(t, err)

		var msg2 string
		err = session.Receive(&msg2)
		unitest.NotError(t, err)
		unitest.Pass(t, msg1 == msg2)
	}
}

func Test_String(t *testing.T) {
	SessionTest(t, String(Uint16BE), StringTest)
}

func Test_Bufio_String(t *testing.T) {
	SessionTest(t, Bufio(String(Uint16BE)), StringTest)
}

func Test_Packet_String(t *testing.T) {
	SessionTest(t, Packet(Uint16BE, String(Uint16BE)), StringTest)
}

func Test_Bufio_Packet_String(t *testing.T) {
	SessionTest(t, Bufio(Packet(Uint16BE, String(Uint16BE))), StringTest)
}

func ObjectTest(t *testing.T, session *link.Session) {
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

func Test_Gob(t *testing.T) {
	SessionTest(t, Gob(), ObjectTest)
}

func Test_Bufio_Gob(t *testing.T) {
	SessionTest(t, Bufio(Gob()), ObjectTest)
}

func Test_Packet_Gob(t *testing.T) {
	SessionTest(t, Packet(Uint16BE, Gob()), ObjectTest)
}

func Test_Bufio_Packet_Gob(t *testing.T) {
	SessionTest(t, Bufio(Packet(Uint16BE, Gob())), ObjectTest)
}

func Test_Json(t *testing.T) {
	SessionTest(t, Json(), ObjectTest)
}

func Test_Bufio_Json(t *testing.T) {
	SessionTest(t, Bufio(Json()), ObjectTest)
}

func Test_Packet_Json(t *testing.T) {
	SessionTest(t, Packet(Uint16BE, Json()), ObjectTest)
}

func Test_Bufio_Packet_Json(t *testing.T) {
	SessionTest(t, Bufio(Packet(Uint16BE, Json())), ObjectTest)
}

func Test_Xml(t *testing.T) {
	SessionTest(t, Xml(), ObjectTest)
}

func Test_Bufio_Xml(t *testing.T) {
	SessionTest(t, Bufio(Xml()), ObjectTest)
}

func Test_Packet_Xml(t *testing.T) {
	SessionTest(t, Packet(Uint16BE, Xml()), ObjectTest)
}

func Test_Bufio_Packet_Xml(t *testing.T) {
	SessionTest(t, Bufio(Packet(Uint16BE, Xml())), ObjectTest)
}

func Test_SelfCodec(t *testing.T) {
	SessionTest(t, SelfCodec(), ObjectTest)
}

func Test_Bufio_SelfCodec(t *testing.T) {
	SessionTest(t, Bufio(SelfCodec()), ObjectTest)
}

func Test_Packet_SelfCodec(t *testing.T) {
	SessionTest(t, Packet(Uint16BE, SelfCodec()), ObjectTest)
}

func Test_Bufio_Packet_SelfCodec(t *testing.T) {
	SessionTest(t, Bufio(Packet(Uint16BE, SelfCodec())), ObjectTest)
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
