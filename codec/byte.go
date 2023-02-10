package codec

import (
	"errors"
	"io"

	"github.com/younglifestyle/link"
)

type ByteProtocol struct {
	//data []byte
}

func (b *ByteProtocol) NewCodec(rw io.ReadWriter) (link.Codec, error) {
	codec := &byteCodec{
		p: b,
		r: rw,
		w: rw,
	}

	codec.closer, _ = rw.(io.Closer)
	return codec, nil
}

func Byte() *ByteProtocol {
	return &ByteProtocol{
		//data: make([]byte, 0),
	}
}

type byteCodec struct {
	r      io.Reader
	w      io.Writer
	p      *ByteProtocol
	closer io.Closer
}

//
func (c *byteCodec) Receive() (interface{}, error) {

	recvData := make([]byte, 4092)

	cnt, err := c.r.Read(recvData)

	return recvData[:cnt], err
}

func (c *byteCodec) Send(msg interface{}) error {

	b, ok := msg.([]byte)
	if !ok {
		return errors.New("Send Byte Format Error")
	}

	_, err := c.w.Write(b)

	return err
}

func (c *byteCodec) Close() error {
	if c.closer != nil {
		return c.closer.Close()
	}
	return nil
}
