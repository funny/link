package link

import (
	"io"
)

type Protocol interface {
	NewCodec() Codec
}

type Codec interface {
	Handshake(rw io.ReadWriter, inBuf, outBuf *Buffer) error
	Prepend(outBuf *Buffer, msg Message)
	Write(w io.Writer, outBuf *Buffer) error
	Read(r io.Reader, inBuf *Buffer) error
}

type Request interface {
	Process(*Session) error
}

type Decoder interface {
	Decode(*Buffer) (Request, error)
}

func DecodeFunc(callback func(*Buffer) (Request, error)) Decoder {
	return decodeFunc{callback}
}

type decodeFunc struct {
	Callback func(*Buffer) (Request, error)
}

func (decoder decodeFunc) Decode(buffer *Buffer) (Request, error) {
	return decoder.Callback(buffer)
}
