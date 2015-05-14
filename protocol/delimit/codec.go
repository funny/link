package delimit

import (
	"github.com/funny/link"
)

var LineBased = New('\n')

var _ link.Protocol = protocol{}
var _ link.Codec = protocol{}

func New(delim byte) link.Protocol {
	return protocol{delim}
}

type protocol struct {
	delim byte
}

func (protocol protocol) NewCodec() link.Codec {
	return protocol
}

func (codec protocol) makeBuffer(buf *link.Buffer, msg link.Message) error {
	// prepend packet buffer
	size := 1024
	if sizeable, ok := msg.(link.Sizeable); ok {
		size = sizeable.BufferSize() + 1
	}
	buf.Reset(0, size)

	// write pakcet content
	if err := msg.WriteBuffer(buf); err != nil {
		return err
	}

	// write packet delimiter
	buf.WriteByte(codec.delim)
	return nil
}

func (codec protocol) MakeBroadcast(buf *link.Buffer, msg link.Message) error {
	return codec.makeBuffer(buf, msg)
}

func (codec protocol) SendBroadcast(conn *link.Conn, buf *link.Buffer) error {
	_, err := conn.Write(buf.Data)
	return err
}

func (codec protocol) SendMessage(conn *link.Conn, buf *link.Buffer, msg link.Message) error {
	err := codec.makeBuffer(buf, msg)
	if err != nil {
		return err
	}
	_, err = conn.Write(buf.Data)
	return err
}

func (codec protocol) ProcessRequest(conn *link.Conn, buf *link.Buffer, handler link.RequestHandler) error {
	data, err := conn.Reader.ReadBytes(codec.delim)
	if err != nil {
		return err
	}
	data = data[:len(data)-1]
	buf.Reset(0, len(data))
	buf.WriteBytes(data)
	return handler(buf)
}
