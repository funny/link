package varhead

import (
	"encoding/binary"
	"github.com/funny/rush/link"
	"github.com/funny/rush/link/linkutil"
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

func (codec protocol) Handshake(rw io.ReadWriter, inBuf, outBuf *link.Buffer) error {
	return nil
}

func (codec protocol) Prepend(buf *link.Buffer, msg link.Message) {
	size := 1024
	if sizeable, ok := msg.(link.Sizeable); ok {
		size = sizeable.BufferSize()
	}
	buf.Reset(binary.MaxVarintLen64, binary.MaxVarintLen64+size)
}

func (codec protocol) Write(w io.Writer, buf *link.Buffer) error {
	size := len(buf.Data) - binary.MaxVarintLen64
	headSize := linkutil.UvarintSize(uint64(size))
	p := buf.Data[binary.MaxVarintLen64-headSize:]
	binary.PutUvarint(p, uint64(size))
	_, err := w.Write(p)
	return err
}

func (codec protocol) Read(r io.Reader, buf *link.Buffer) error {
	byteReader, isByteReader := r.(io.ByteReader)
	if !isByteReader {
		byteReader = slowByteReader{r: r}
	}
	size, err := binary.ReadUvarint(byteReader)
	if err != nil {
		return err
	}
	buf.Reset(int(size), int(size))
	if _, err := io.ReadFull(r, buf.Data); err != nil {
		return err
	}
	return nil
}

type slowByteReader struct {
	r io.Reader
	b [1]byte
}

func (r slowByteReader) ReadByte() (c byte, err error) {
	if _, err := io.ReadFull(r.r, r.b[:]); err != nil {
		return 0, err
	}
	return r.b[0], nil
}
