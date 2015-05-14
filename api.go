package link

type Protocol interface {
	NewCodec() Codec
}

type Handshake interface {
	Handshake(conn *Conn, buf *Buffer) error
}

type Codec interface {
	Prepend(outBuf *Buffer, msg Message)
	Write(conn *Conn, outBuf *Buffer) error
	Read(conn *Conn, inBuf *Buffer) error
}

type FrameCodec interface {
	PrependFrame(outBuf *Buffer, frame FrameMessage)
	WriteFrame(conn *Conn, outBuf *Buffer) error
	ReadFrame(conn *Conn, inBuf *Buffer) (isFinal bool, err error)
}

type FrameMessage interface {
	Message
	IsFinalFrame() bool
	NextFrame() FrameMessage
}

type Message interface {
	WriteBuffer(*Buffer) error
}

type Sizeable interface {
	BufferSize() int
}

type RequestHandler func(*Buffer) error
