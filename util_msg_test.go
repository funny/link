package link

import (
	"bytes"
	"github.com/funny/unitest"
	"math/rand"
	"testing"
)

type TestMessage struct {
	Command int
	Value   []byte
}

func Test_JSON(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := TestMessage{rand.Intn(0xFFFFFFFF), RandBytes(1024)}
		err := JSON{b1, SplitByLine}.Send(w)
		w.Flush()
		unitest.NotError(t, err)
		unitest.NotError(t, w.werr)

		var b2 TestMessage
		err = JSON{&b2, SplitByLine}.Receive(r)
		unitest.NotError(t, err)
		unitest.NotError(t, r.rerr)

		unitest.Pass(t, b1.Command == b2.Command && bytes.Equal(b1.Value, b2.Value))
	})
}

func Test_GOB(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := TestMessage{rand.Intn(0xFFFFFFFF), RandBytes(1024)}
		err := GOB{b1, SplitByUint16BE}.Send(w)
		w.Flush()
		unitest.NotError(t, err)
		unitest.NotError(t, w.werr)

		var b2 TestMessage
		err = GOB{&b2, SplitByUint16BE}.Receive(r)
		unitest.NotError(t, err)
		unitest.NotError(t, r.rerr)

		unitest.Pass(t, b1.Command == b2.Command && bytes.Equal(b1.Value, b2.Value))
	})
}
