package link

import (
	"encoding/binary"
	"io"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/funny/utest"
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
	utest.IsNilNow(t, err)
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

	testFunc := func() {
		session, err := Connect("tcp", addr, codecType)
		utest.IsNilNow(t, err)
		test(t, session)
		session.Close()
		clientWait.Done()
	}

	for i := 0; i < 30; i++ {
		clientWait.Add(1)
		go testFunc()
	}
	clientWait.Wait()

	for i := 0; i < 30; i++ {
		clientWait.Add(1)
		go testFunc()
	}
	clientWait.Wait()

	server.Stop()
	serverWait.Wait()
}

func BytesTest(t *testing.T, session *Session) {
	for i := 0; i < 2000; i++ {
		msg1 := RandBytes(512)
		err := session.Send(msg1)
		utest.IsNilNow(t, err)

		var msg2 = make([]byte, len(msg1))
		err = session.Receive(msg2)
		utest.IsNilNow(t, err)
		utest.EqualNow(t, msg1, msg2)
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

func (tb TestObject) Equals(a interface{}) bool {
	return tb == a.(TestObject)
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
		utest.IsNilNow(t, err)

		var msg2 TestObject
		err = session.Receive(&msg2)
		utest.IsNilNow(t, err)
		utest.EqualNow(t, msg1, msg2)
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

func Test_BufioSize(t *testing.T) {
	SessionTest(t, BufioSize(0, 0, Json()), ObjectTest)
}

type TestFbService struct{}

func (s TestFbService) ServiceID() byte {
	return 1
}

func (s TestFbService) NewRequest(id byte) (FbMessage, FbHandler) {
	switch id {
	case 1:
		return new(TestFbMessage), nil
	case 2:
		return new(TestFbMessage), nil
	}
	return nil, nil
}

type TestFbMessage struct {
	messageID byte
	TestObject
}

func (m *TestFbMessage) ServiceID() byte {
	return 1
}

func (m *TestFbMessage) MessageID() byte {
	return 2
}

func (tb *TestObject) BinarySize() int {
	return 8 + 8 + 8
}

func (tb *TestObject) MarshalPacket(b []byte) {
	binary.LittleEndian.PutUint64(b[0:8], uint64(tb.X))
	binary.LittleEndian.PutUint64(b[8:16], uint64(tb.Y))
	binary.LittleEndian.PutUint64(b[16:24], uint64(tb.Z))
}

func (tb *TestObject) UnmarshalPacket(b []byte) {
	tb.X = int(binary.LittleEndian.Uint64(b[0:8]))
	tb.Y = int(binary.LittleEndian.Uint64(b[8:16]))
	tb.Z = int(binary.LittleEndian.Uint64(b[16:24]))
}

func Test_Fastbin(t *testing.T) {
	fbCodecType := Fastbin(4096, nil)
	fbCodecType.Register(TestFbService{})

	SessionTest(t, fbCodecType, func(t *testing.T, session *Session) {
		for i := 0; i < 2000; i++ {
			var msgID byte = 1
			if rand.Intn(100) < 50 {
				msgID = 2
			}
			msg1 := RandObject()
			err := session.Send(&TestFbMessage{
				msgID,
				msg1,
			})
			utest.IsNilNow(t, err)

			var msg2 FbRequest
			err = session.Receive(&msg2)
			utest.IsNilNow(t, err)
			utest.EqualNow(t, msg1, msg2.message.(*TestFbMessage).TestObject)
		}
	})
}
