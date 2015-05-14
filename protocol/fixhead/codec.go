package fixhead

import (
	"github.com/funny/link"
	"io"
)

var _ link.Protocol = &protocol{}
var _ link.Codec = &protocol{}

type protocol struct {
	n       int
	decoder func(*link.Buffer) int
	encoder func([]byte, int)
}

func (protocol *protocol) NewCodec() link.Codec {
	return protocol
}

func (codec *protocol) makeBuffer(buf *link.Buffer, msg link.Message) error {
	// prepend packet head
	size := 1024
	if sizeable, ok := msg.(link.Sizeable); ok {
		size = sizeable.BufferSize()
	}
	buf.Reset(codec.n, codec.n+size)

	// write packet body
	if err := msg.WriteBuffer(buf); err != nil {
		return err
	}

	// write packet head
	size = len(buf.Data) - codec.n
	codec.encoder(buf.Data, size)
	return nil
}

func (codec *protocol) MakeBroadcast(buf *link.Buffer, msg link.Message) error {
	return codec.makeBuffer(buf, msg)
}

func (codec *protocol) SendBroadcast(conn *link.Conn, buf *link.Buffer) error {
	_, err := conn.Write(buf.Data)
	return err
}

func (codec *protocol) SendMessage(conn *link.Conn, buf *link.Buffer, msg link.Message) error {
	err := codec.makeBuffer(buf, msg)
	if err != nil {
		return err
	}
	_, err = conn.Write(buf.Data)
	return err
}

func (codec *protocol) ProcessRequest(conn *link.Conn, buf *link.Buffer, handler link.RequestHandler) error {
	buf.Reset(codec.n, codec.n)
	if _, err := io.ReadFull(conn, buf.Data); err != nil {
		return err
	}
	size := codec.decoder(buf)
	buf.Reset(size, size)
	if _, err := io.ReadFull(conn, buf.Data); err != nil {
		return err
	}
	return handler(buf)
}
