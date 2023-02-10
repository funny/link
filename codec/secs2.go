package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/wolimst/lib-secs2-hsms-go/pkg/parser/hsms"
	"github.com/younglifestyle/link"
	"io"
)

var ErrMsgParsing = errors.New("message parsing error")

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

type secsReadWriter struct {
	recvBuf bytes.Reader
	sendBuf bytes.Buffer
}

func (rw *secsReadWriter) Read(p []byte) (int, error) {
	return rw.recvBuf.Read(p)
}

func (rw *secsReadWriter) Write(p []byte) (int, error) {
	return rw.sendBuf.Write(p)
}

type secsIICodec struct {
	rw          io.ReadWriter
	closer      io.Closer
	headBuf     []byte
	bodyBuf     []byte
	headDecoder func([]byte) uint32
	headEncoder func([]byte, uint32)
	secsReadWriter
}

func (c *secsIICodec) Receive() (interface{}, error) {
	var msgLength uint32

	if _, err := io.ReadFull(c.rw, c.headBuf); err != nil {
		return nil, err
	}
	msgLength = c.headDecoder(c.headBuf)

	// 协议允许的数据包很大，无需特殊考虑最大包大小
	if msgLength >= 10 {
		if cap(c.bodyBuf) < int(msgLength) {
			c.bodyBuf = make([]byte, msgLength+4)
		}

		if _, err := io.ReadFull(c.rw, c.bodyBuf[4:msgLength+4]); err != nil {
			return nil, err
		}
	}

	// 重组消息数据
	binary.BigEndian.PutUint32(c.bodyBuf, msgLength)

	msg, ok := hsms.Parse(c.bodyBuf)
	if !ok {
		return nil, ErrMsgParsing
	}

	return msg, nil
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
