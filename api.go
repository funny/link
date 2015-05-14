package link

type Protocol interface {
	NewCodec() Codec
}

type Codec interface {
	MakeBroadcast(buf *Buffer, msg Message) error
	SendBroadcast(conn *Conn, buf *Buffer) error
	SendMessage(conn *Conn, buf *Buffer, msg Message) error
	ProcessRequest(conn *Conn, buf *Buffer, handler RequestHandler) error
}

type Message interface {
	WriteBuffer(*Buffer) error
}

type RequestHandler func(*Buffer) error

type Handshake interface {
	Handshake(conn *Conn, buf *Buffer) error
}

type Sizeable interface {
	BufferSize() int
}
