package link

import (
	"bytes"
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

func NewTestCodec(rw io.ReadWriter) (Codec, error) {
	return &TestCodec{
		rw: rw.(io.ReadWriteCloser),
	}, nil
}

type TestCodec struct {
	rw io.ReadWriteCloser
}

func (c *TestCodec) Send(msg interface{}) error {
	var head [2]byte
	binary.LittleEndian.PutUint16(head[:], uint16(len(msg.([]byte))))
	_, err := c.rw.Write(head[:])
	if err != nil {
		return err
	}
	_, err = c.rw.Write(msg.([]byte))
	if err != nil {
		return err
	}
	return nil
}

func (c *TestCodec) Receive() (interface{}, error) {
	var head [2]byte
	_, err := io.ReadFull(c.rw, head[:])
	if err != nil {
		return nil, err
	}
	n := binary.LittleEndian.Uint16(head[:])
	buf := make([]byte, n)
	_, err = io.ReadFull(c.rw, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (c *TestCodec) Close() error {
	return c.rw.Close()
}

func (c *TestCodec) ClearSendChan(ch <-chan interface{}) {
	for _ = range ch {
	}
}

func RandBytes(n int) []byte {
	n = rand.Intn(n) + 1
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(rand.Intn(255))
	}
	return b
}

func SessionTest(t *testing.T, sendChanSize int, test func(*testing.T, *Session)) {
	server, err := Listen("tcp", "0.0.0.0:0", ProtocolFunc(NewTestCodec), sendChanSize, HandlerFunc(func(session *Session) {
		defer session.Close()
		for {
			msg, err := session.Receive()
			if err != nil {
				return
			}
			err = session.Send(msg)
			if err != nil {
				return
			}
		}
	}))
	utest.IsNilNow(t, err)
	go server.Serve()

	addr := server.Listener().Addr().String()

	clientWait := new(sync.WaitGroup)
	for i := 0; i < 60; i++ {
		clientWait.Add(1)
		go func() {
			session, err := Dial("tcp", addr, ProtocolFunc(NewTestCodec), sendChanSize)
			utest.IsNilNow(t, err)
			test(t, session)
			session.Close()
			clientWait.Done()
		}()
	}
	clientWait.Wait()

	server.Stop()
}

func BytesTest(t *testing.T, session *Session) {
	for i := 0; i < 2000; i++ {
		msg1 := RandBytes(512)
		err := session.Send(msg1)
		utest.IsNilNow(t, err)

		msg2, err := session.Receive()
		utest.IsNilNow(t, err)
		utest.Assert(t, bytes.Equal(msg1, msg2.([]byte)))
	}
}

func Test_CloseCallback(t *testing.T) {
	session := newSession(nil, nil, 0)

	c := make(chan int, 10)
	for i := 0; i < 10; i++ {
		func(n int) {
			callback := func() {
				c <- n
			}
			session.AddCloseCallback(nil, n, callback)
			session.RemoveCloseCallback(nil, n)
			session.AddCloseCallback(nil, n, callback)
		}(i)
	}

	session.invokeCloseCallbacks()

	for i := 0; i < 10; i++ {
		n := <-c
		utest.EqualNow(t, i, n)
	}
}

func Test_Sync(t *testing.T) {
	SessionTest(t, 0, BytesTest)
}

func Test_Async(t *testing.T) {
	SessionTest(t, 1024, BytesTest)
}

func Test_Channel(t *testing.T) {
	waitTestDone := make(chan struct{})

	channel := NewChannel()
	testMessages := make([][]byte, 2000)

	waitClientReady := new(sync.WaitGroup)
	waitClientReady.Add(60)
	go func() {
		waitClientReady.Wait()

		for i := 0; i < 2000; i++ {
			msg := RandBytes(128)
			testMessages[i] = msg
			channel.Fetch(func(s *Session) {
				s.Send(testMessages[i])
			})
		}

		<-waitTestDone

		channel.Close()
	}()

	server, err := Listen("tcp", "0.0.0.0:0", ProtocolFunc(NewTestCodec), 2000, HandlerFunc(func(session *Session) {
		defer session.Close()
		channel.Put(session.ID(), session)

		utest.EqualNow(t, channel.Get(session.ID()), session)
		utest.Assert(t, channel.Remove(session.ID()))
		utest.EqualNow(t, channel.Get(session.ID()), nil)

		channel.Put(session.ID(), session)

		waitClientReady.Done()

		<-waitTestDone
	}))
	utest.IsNilNow(t, err)
	go server.Serve()

	addr := server.Listener().Addr().String()

	waitTestFinish := new(sync.WaitGroup)
	for i := 0; i < 60; i++ {
		waitTestFinish.Add(1)

		go func() {
			session, err := DialTimeout("tcp", addr, time.Second, ProtocolFunc(NewTestCodec), 0)
			utest.IsNilNow(t, err)

			for j := 0; j < 2000; j++ {
				msg, err := session.Receive()
				utest.IsNilNow(t, err)
				utest.EqualNow(t, msg.([]byte), testMessages[j])
			}

			session.Close()

			waitTestFinish.Done()
		}()
	}
	waitTestFinish.Wait()

	close(waitTestDone)

	server.Stop()
}

func Benchmark_BytesToInterface(b *testing.B) {
	var a = []byte{}
	var x interface{}
	for i := 0; i < b.N; i++ {
		x = a
	}
	_ = x
}

func Benchmark_InterfaceToBytes(b *testing.B) {
	var x interface{} = []byte{}
	var a []byte
	for i := 0; i < b.N; i++ {
		a = x.([]byte)
	}
	_ = a
}

func Benchmark_PointerToInterface(b *testing.B) {
	var a = struct{}{}
	var x interface{}
	for i := 0; i < b.N; i++ {
		x = &a
	}
	_ = x
}

func Benchmark_InterfaceToPointer(b *testing.B) {
	var x interface{} = new(struct{})
	var a *struct{}
	for i := 0; i < b.N; i++ {
		a = x.(*struct{})
	}
	_ = a
}
