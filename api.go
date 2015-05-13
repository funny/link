package link

type Protocol interface {
	NewCodec() Codec
}

type Codec interface {
	Handshake(conn *Conn, buf *Buffer) error
	Prepend(outBuf *Buffer, msg Message)
	Write(conn *Conn, outBuf *Buffer) error
	Read(conn *Conn, inBuf *Buffer) error
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
