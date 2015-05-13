package varhead

import (
	"bytes"
	"github.com/funny/rush/link"
	"github.com/funny/unitest"
	"math/rand"
	"testing"
	"time"
)

func Test_Uvarint(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var (
		codec protocol
		wbuf  = link.MakeBuffer(0, 0)
		rbuf  = link.MakeBuffer(0, 0)
		tbuf  = link.MakeBuffer(0, 0)
	)
	for i := 0; i < 10000; i++ {
		msg := RandMessage()

		codec.Prepend(wbuf, msg)
		msg.WriteBuffer(wbuf)
		codec.Write(rbuf, wbuf)

		codec.Read(rbuf, tbuf)
		unitest.Pass(t, bytes.Equal(msg, tbuf.Data))
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
