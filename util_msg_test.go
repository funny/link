package link

import (
	"bytes"
	"github.com/funny/binary"
	"github.com/funny/unitest"
	"math/rand"
	"os"
	"testing"
)

type TestMessage struct {
	Command int
	Value   []byte
}

func RandBytes(n int) []byte {
	n = rand.Intn(n)
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(rand.Intn(255))
	}
	return b
}

func ReadWriteTest(t *testing.T, n int, callback func(r *binary.Reader, w *binary.Writer)) {
	pipeR, pipeW, err := os.Pipe()
	unitest.NotError(t, err)
	r := binary.NewBufioReader(pipeR, 1024)
	w := binary.NewBufioWriter(pipeW, 1024)
	for i := 0; i < n; i++ {
		callback(r, w)
	}
}

func Test_JSON(t *testing.T) {
	ReadWriteTest(t, 10000, func(r *binary.Reader, w *binary.Writer) {
		b1 := TestMessage{rand.Intn(0xFFFFFFFF), RandBytes(1024)}
		err := JSON{b1, binary.SplitByLine}.Send(w)
		w.Flush()
		unitest.NotError(t, err)
		unitest.NotError(t, w.Error())

		var b2 TestMessage
		err = JSON{&b2, binary.SplitByLine}.Receive(r)
		unitest.NotError(t, err)
		unitest.NotError(t, r.Error())

		unitest.Pass(t, b1.Command == b2.Command && bytes.Equal(b1.Value, b2.Value))
	})
}

func Test_GOB(t *testing.T) {
	ReadWriteTest(t, 10000, func(r *binary.Reader, w *binary.Writer) {
		b1 := TestMessage{rand.Intn(0xFFFFFFFF), RandBytes(1024)}
		err := GOB{b1, binary.SplitByUint16BE}.Send(w)
		w.Flush()
		unitest.NotError(t, err)
		unitest.NotError(t, w.Error())

		var b2 TestMessage
		err = GOB{&b2, binary.SplitByUint16BE}.Receive(r)
		unitest.NotError(t, err)
		unitest.NotError(t, r.Error())

		unitest.Pass(t, b1.Command == b2.Command && bytes.Equal(b1.Value, b2.Value))
	})
}
