package link

import (
	"bytes"
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

func (_ TestCodec) NewEncoder(w io.Writer) Encoder {
	return TestEncoder{w}
}

func (_ TestCodec) NewDecoder(r io.Reader) Decoder {
	return TestDecoder{r}
}

type TestEncoder struct {
	w io.Writer
}

func (encoder TestEncoder) Encode(msg interface{}) error {
	_, err := encoder.w.Write(msg.([]byte))
	return err
}

type TestDecoder struct {
	r io.Reader
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
	unitest.NotError(t, err)
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
			unitest.NotError(t, err)
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
		unitest.NotError(t, err)

		var msg2 = make([]byte, len(msg1))
		err = session.Receive(msg2)
		unitest.NotError(t, err)
		unitest.Pass(t, bytes.Equal(msg1, msg2))
	}
}

func Test_Bytes(t *testing.T) {
	SessionTest(t, TestCodec{}, BytesTest)
}
