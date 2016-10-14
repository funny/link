package codec

import (
	"bufio"
	"io"

	"github.com/funny/link"
)

func Bufio(base link.Protocol, readBuf, writeBuf int) link.Protocol {
	return &bufioProtocol{
		base:     base,
		readBuf:  readBuf,
		writeBuf: writeBuf,
	}
}

type bufioProtocol struct {
	base     link.Protocol
	readBuf  int
	writeBuf int
}

func (b *bufioProtocol) NewCodec(rw io.ReadWriter) (cc link.Codec, err error) {
	codec := new(bufioCodec)

	if b.writeBuf > 0 {
		codec.stream.w = bufio.NewWriterSize(rw, b.writeBuf)
		codec.stream.Writer = codec.stream.w
	} else {
		codec.stream.Writer = rw
	}

	if b.readBuf > 0 {
		codec.stream.Reader = bufio.NewReaderSize(rw, b.readBuf)
	} else {
		codec.stream.Reader = rw
	}

	codec.stream.c, _ = rw.(io.Closer)

	codec.base, err = b.base.NewCodec(&codec.stream)
	if err != nil {
		return
	}
	cc = codec
	return
}

type bufioStream struct {
	io.Reader
	io.Writer
	c io.Closer
	w *bufio.Writer
}

func (s *bufioStream) Flush() error {
	if s.w != nil {
		return s.w.Flush()
	}
	return nil
}

func (s *bufioStream) close() error {
	if s.c != nil {
		return s.c.Close()
	}
	return nil
}

type bufioCodec struct {
	base   link.Codec
	stream bufioStream
}

func (c *bufioCodec) Send(msg interface{}) error {
	if err := c.base.Send(msg); err != nil {
		return err
	}
	return c.stream.Flush()
}

func (c *bufioCodec) Receive() (interface{}, error) {
	return c.base.Receive()
}

func (c *bufioCodec) Close() error {
	err1 := c.base.Close()
	err2 := c.stream.close()
	if err1 != nil {
		return err1
	}
	return err2
}
