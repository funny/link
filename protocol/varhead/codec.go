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

func (codec protocol) Prepend(buf *link.Buffer, msg link.Message) {
	size := 1024
	if sizeable, ok := msg.(link.Sizeable); ok {
		size = sizeable.BufferSize()
	}
	buf.Reset(binary.MaxVarintLen64, binary.MaxVarintLen64+size)
}

func (codec protocol) Write(conn *link.Conn, buf *link.Buffer) error {
	size := len(buf.Data) - binary.MaxVarintLen64
	headSize := linkutil.UvarintSize(uint64(size))
	p := buf.Data[binary.MaxVarintLen64-headSize:]
	binary.PutUvarint(p, uint64(size))
	_, err := conn.Write(p)
	return err
}

func (codec protocol) Read(conn *link.Conn, buf *link.Buffer) error {
	size, err := binary.ReadUvarint(conn.Reader)
	if err != nil {
		return err
	}
	buf.Reset(int(size), int(size))
	if _, err := io.ReadFull(conn, buf.Data); err != nil {
		return err
	}
	return nil
}
