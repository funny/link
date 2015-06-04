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
