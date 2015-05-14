package varhead

import (
	"bytes"
	"github.com/funny/link"
	"github.com/funny/unitest"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

func Test_Uvarint(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	unitest.NotError(t, err)
	defer l.Close()

	var (
		writer *link.Conn
		reader *link.Conn
		wg     sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		c, err := l.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		reader = link.NewConn(c, 1024)
	}()

	c, err := net.Dial("tcp", l.Addr().String())
	unitest.NotError(t, err)
	writer = link.NewConn(c, 1024)
	wg.Wait()
	unitest.Pass(t, reader != nil)

	rand.Seed(time.Now().UnixNano())
	var (
		codec protocol
		wbuf  = link.MakeBuffer(0, 0)
		rbuf  = link.MakeBuffer(0, 0)
	)
	for i := 0; i < 10000; i++ {
		msg := RandMessage()

		codec.SendMessage(writer, wbuf, msg)

		codec.ProcessRequest(reader, rbuf, func(buf *link.Buffer) error {
			unitest.Pass(t, bytes.Equal(msg, buf.Data))
			return nil
		})
	}
}

type TestMessage []byte

func (msg TestMessage) BufferSize() int {
	return len(msg)
}

func (msg TestMessage) WriteBuffer(buf *link.Buffer) error {
	buf.WriteBytes(msg)
	return nil
}

func RandMessage() TestMessage {
	n := rand.Intn(2048)
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(rand.Intn(255))
	}
	return TestMessage(b)
}
