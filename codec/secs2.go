package codec

import (
	"encoding/binary"
	"errors"
	"github.com/younglifestyle/link"
	"io"
)

type SecsIIProtocol struct {
}

func (b *SecsIIProtocol) NewCodec(rw io.ReadWriter) (link.Codec, error) {
	codec := &secsIICodec{
		rw:      rw,
		headBuf: make([]byte, 4), // 长度
		headDecoder: func(bytes []byte) uint32 {
			return binary.BigEndian.Uint32(bytes)
		},
	}

	codec.closer, _ = rw.(io.Closer)
	return codec, nil
}

func SECSII() *SecsIIProtocol {
	return &SecsIIProtocol{}
}

type secsIICodec struct {
	rw          io.ReadWriter
	closer      io.Closer
	headBuf     []byte
	headDecoder func([]byte) uint32
	headEncoder func([]byte, uint32)
}

func (c *secsIICodec) Receive() (interface{}, error) {
	if _, err := io.ReadFull(c.rw, c.headBuf); err != nil {
		return nil, err
	}
	size := c.headDecoder(c.headBuf)

	recvData := make([]byte, 4092)

	cnt, err := c.rw.Read(recvData)

	return recvData[:cnt], err
}

func (c *secsIICodec) Send(msg interface{}) error {

	b, ok := msg.([]byte)
	if !ok {
		return errors.New("Send Byte Format Error")
	}

	_, err := c.rw.Write(b)

	return err
}

func (c *secsIICodec) Close() error {
	if c.closer != nil {
		return c.closer.Close()
	}
	return nil
}
