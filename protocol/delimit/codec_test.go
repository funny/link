package delimit

import (
	"bytes"
	"encoding/base64"
	"github.com/funny/link"
	"github.com/funny/unitest"
	"io/ioutil"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

func Test_DelimitCodec(t *testing.T) {
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
		codec = protocol{'\n'}
		wbuf  = link.MakeBuffer(0, 0)
		rbuf  = link.MakeBuffer(0, 0)
	)
	for i := 0; i < 10000; i++ {
		msg1 := RandMessage()

		codec.SendMessage(writer, wbuf, msg1)

		codec.ProcessRequest(reader, rbuf, func(buf *link.Buffer) error {
			decoder := base64.NewDecoder(base64.StdEncoding, buf)
			msg2, err := ioutil.ReadAll(decoder)
			if err != nil {
				t.Log(err)
			}
			unitest.Pass(t, bytes.Equal(msg1, msg2))
			return nil
		})
	}
}

type TestMessage []byte

func (msg TestMessage) BufferSize() int {
	return base64.StdEncoding.EncodedLen(len(msg))
}

func (msg TestMessage) WriteBuffer(buf *link.Buffer) error {
	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	if _, err := encoder.Write(msg); err != nil {
		return err
	}
	return encoder.Close()
}

func RandMessage() TestMessage {
	n := rand.Intn(2048)
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(rand.Intn(255))
	}
	return TestMessage(b)
}
