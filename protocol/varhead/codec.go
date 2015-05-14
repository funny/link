package varhead

import (
	"encoding/binary"
	"github.com/funny/link"
	"github.com/funny/link/linkutil"
	"io"
)

var Protocol = protocol{}

var _ link.Protocol = protocol{}
var _ link.Codec = protocol{}

type protocol struct {
}

func (protocol protocol) NewCodec() link.Codec {
	return protocol
}

func (codec protocol) makeBuffer(buf *link.Buffer, msg link.Message) error {
	// prepend packet head
	size := 1024
	if sizeable, ok := msg.(link.Sizeable); ok {
		size = sizeable.BufferSize()
	}
	buf.Reset(binary.MaxVarintLen64, binary.MaxVarintLen64+size)

	// write packet body
	return msg.WriteBuffer(buf)
}

func (codec protocol) sendBuffer(conn *link.Conn, buf *link.Buffer) error {
	// write packet head
	size := len(buf.Data) - binary.MaxVarintLen64
	headSize := linkutil.UvarintSize(uint64(size))
	p := buf.Data[binary.MaxVarintLen64-headSize:]
	binary.PutUvarint(p, uint64(size))

	// send
	_, err := conn.Write(p)
	return err
}

func (codec protocol) MakeBroadcast(buf *link.Buffer, msg link.Message) error {
	return codec.makeBuffer(buf, msg)
}

func (codec protocol) SendBroadcast(conn *link.Conn, buf *link.Buffer) error {
	return codec.sendBuffer(conn, buf)
}

func (codec protocol) SendMessage(conn *link.Conn, buf *link.Buffer, msg link.Message) error {
	err := codec.makeBuffer(buf, msg)
	if err != nil {
		return err
	}
	return codec.sendBuffer(conn, buf)
}

func (codec protocol) ProcessRequest(conn *link.Conn, buf *link.Buffer, handler link.RequestHandler) error {
	size, err := binary.ReadUvarint(conn.Reader)
	if err != nil {
		return err
	}
	buf.Reset(int(size), int(size))
	if _, err := io.ReadFull(conn, buf.Data); err != nil {
		return err
	}
	return handler(buf)
}
