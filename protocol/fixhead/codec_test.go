package fixhead

import (
	"bytes"
	"github.com/funny/link"
	"github.com/funny/unitest"
	"math/rand"
	"testing"
	"time"
)

func Test_Uint8(t *testing.T) {
	CodecTest(t, Uint8)
}

func Test_Uint16(t *testing.T) {
	CodecTest(t, Uint16BE)
	CodecTest(t, Uint16LE)
}

func Test_Uint24(t *testing.T) {
	CodecTest(t, Uint24BE)
	CodecTest(t, Uint24LE)
}

func Test_Uint32(t *testing.T) {
	CodecTest(t, Uint32BE)
	CodecTest(t, Uint32LE)
}

func Test_Uint40(t *testing.T) {
	CodecTest(t, Uint40BE)
	CodecTest(t, Uint40LE)
}

func Test_Uint48(t *testing.T) {
	CodecTest(t, Uint48BE)
	CodecTest(t, Uint48LE)
}

func Test_Uint56(t *testing.T) {
	CodecTest(t, Uint56BE)
	CodecTest(t, Uint56LE)
}

func Test_Uint64(t *testing.T) {
	CodecTest(t, Uint64BE)
	CodecTest(t, Uint64LE)
}

func CodecTest(t *testing.T, codec *protocol) {
	rand.Seed(time.Now().UnixNano())
	var (
		wbuf = link.MakeBuffer(0, 0)
		rbuf = link.MakeBuffer(0, 0)
		tbuf = link.MakeBuffer(0, 0)
	)
	for i := 0; i < 10000; i++ {
		msg := RandMessage(codec)

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

func RandMessage(codec *protocol) TestMessage {
	var n int
	if codec == Uint8 {
		n = rand.Intn(255)
	} else {
		n = rand.Intn(2048)
	}
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(rand.Intn(255))
	}
	return TestMessage(b)
}
