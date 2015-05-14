package delimit

import (
	"github.com/funny/link"
)

var LineBased = New('\n')

var _ link.Protocol = &protocol{}
var _ link.Codec = &protocol{}

func New(delim byte) link.Protocol {
	return protocol{delim}
}

type protocol struct {
	delim byte
}

func (protocol protocol) NewCodec() link.Codec {
	return protocol
}

func (codec protocol) Handshake(conn *link.Conn, buf *link.Buffer) error {
	return nil
}

func (codec protocol) Prepend(buf *link.Buffer, msg link.Message) {
	size := 1024
	if sizeable, ok := msg.(link.Sizeable); ok {
		size = sizeable.BufferSize()
	}
	buf.Reset(0, size+1)
}

func (codec protocol) Write(conn *link.Conn, buf *link.Buffer) error {
	buf.WriteByte(codec.delim)
	_, err := conn.Write(buf.Data)
	return err
}

func (codec protocol) Read(conn *link.Conn, buf *link.Buffer) error {
	data, err := conn.Reader.ReadBytes(codec.delim)
	if err != nil {
		return err
	}
	data = data[:len(data)-1]
	buf.Reset(0, len(data))
	buf.WriteBytes(data)
	return nil
}
