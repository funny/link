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

type FrameCodec interface {
	ReadFrame(conn *Conn, inBuf *Buffer) (isFinal bool, err error)
}

type Frame interface {
	IsFinalFrame() bool
}

type FrameMessage interface {
	Message
	Frame
	NextFrame() FrameMessage
}

type RequestHandler func(*Buffer) error
