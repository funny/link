package link

import (
	"io"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/funny/unitest"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type TestCodec struct{}

type TestEncoder struct {
	w io.Writer
}

type TestDecoder struct {
	r io.Reader
}

func (_ TestCodec) NewEncoder(w io.Writer) Encoder {
	return TestEncoder{w}
}

func (_ TestCodec) NewDecoder(r io.Reader) Decoder {
	return TestDecoder{r}
}

func (encoder TestEncoder) Encode(msg interface{}) error {
	_, err := encoder.w.Write(msg.([]byte))
	return err
}

func (decoder TestDecoder) Decode(msg interface{}) error {
	_, err := io.ReadFull(decoder.r, msg.([]byte))
	return err
}

func RandBytes(n int) []byte {
	n = rand.Intn(n) + 1
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(rand.Intn(255))
	}
	return b
}

func SessionTest(t *testing.T, codecType CodecType, test func(*testing.T, *Session)) {
	server, err := Serve("tcp", "0.0.0.0:0", TestCodec{})
	unitest.AssertNotError(t, err)
	addr := server.listener.Addr().String()

	serverWait := new(sync.WaitGroup)
	go func() {
		for {
			session, err := server.Accept()
			if err != nil {
				break
			}
			serverWait.Add(1)
			go func() {
				io.Copy(session.conn, session.conn)
				serverWait.Done()
			}()
		}
	}()

	clientWait := new(sync.WaitGroup)
	for i := 0; i < 60; i++ {
		clientWait.Add(1)
		go func() {
			session, err := Connect("tcp", addr, codecType)
			unitest.AssertNotError(t, err)
			test(t, session)
			session.Close()
			clientWait.Done()
		}()
	}
	clientWait.Wait()

	server.Stop()
	serverWait.Wait()
}

func BytesTest(t *testing.T, session *Session) {
	for i := 0; i < 2000; i++ {
		msg1 := RandBytes(512)
		err := session.Send(msg1)
		unitest.AssertNotError(t, err)

		var msg2 = make([]byte, len(msg1))
		err = session.Receive(msg2)
		unitest.AssertNotError(t, err)
		unitest.AssertBytes(t, msg1, msg2)
	}
}

func Test_Bytes(t *testing.T) {
	SessionTest(t, TestCodec{}, BytesTest)
}

func Test_Async_Bytes(t *testing.T) {
	SessionTest(t, Async(1024, TestCodec{}), BytesTest)
}

func Test_Bufio_Bytes(t *testing.T) {
	SessionTest(t, Bufio(TestCodec{}), BytesTest)
}

func Test_ThreadSafe_Bytes(t *testing.T) {
	SessionTest(t, ThreadSafe(TestCodec{}), BytesTest)
}

func Test_ThreadSafe_Bufio_Bytes(t *testing.T) {
	SessionTest(t, ThreadSafe(Bufio(TestCodec{})), BytesTest)
}

type TestObject struct {
	X, Y, Z int
}

func RandObject() TestObject {
	return TestObject{
		X: rand.Int(), Y: rand.Int(), Z: rand.Int(),
	}
}

func ObjectTest(t *testing.T, session *Session) {
	for i := 0; i < 2000; i++ {
		msg1 := RandObject()
		err := session.Send(&msg1)
		unitest.AssertNotError(t, err)

		var msg2 TestObject
		err = session.Receive(&msg2)
		unitest.AssertNotError(t, err)
		unitest.Assert(t, msg1 == msg2, msg1, msg2)
	}
}

func Test_Gob(t *testing.T) {
	SessionTest(t, Gob(), ObjectTest)
}

func Test_Bufio_Gob(t *testing.T) {
	SessionTest(t, Bufio(Gob()), ObjectTest)
}

func Test_Json(t *testing.T) {
	SessionTest(t, Json(), ObjectTest)
}

func Test_Bufio_Json(t *testing.T) {
	SessionTest(t, Bufio(Json()), ObjectTest)
}

func Test_Xml(t *testing.T) {
	SessionTest(t, Xml(), ObjectTest)
}

func Test_Bufio_Xml(t *testing.T) {
	SessionTest(t, Bufio(Xml()), ObjectTest)
}

func Test_Packet(t *testing.T) {
	SessionTest(t, Packet(2, 1024, 1024, LittleEndian, Json()), ObjectTest)
}
