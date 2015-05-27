package link

import (
	"github.com/funny/binary"
)

var (
	DefaultConfig = Config{
		ConnConfig{
			ReadBufferSize:  2048,
			WriteBufferSize: 2048,
		},
		SessionConfig{
			AutoFlush:         true,
			AsyncSendChanSize: 1000,
		},
	}
)

type OutMessage interface {
	Send(*binary.Writer) error
}

type InMessage interface {
	Receive(*binary.Reader) error
}
