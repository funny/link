package stream

import (
	"github.com/funny/binary"
)

type InMessage interface {
	Unmarshal(*binary.Reader) error
}

type OutMessage interface {
	Marshal(*binary.Writer) error
}

var Flush flushMsg

type flushMsg struct {
}

func (_ flushMsg) Marshal(w *binary.Writer) error {
	return w.Flush()
}
