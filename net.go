package packnet

import (
	"errors"
	"net"
	"time"
)

// Errors
var (
	SendToClosedError          = errors.New("Send to closed session")
	BlockingError              = errors.New("Blocking happened")
	PacketTooLargeError        = errors.New("Packet too large")
	SessionDuplicateStartError = errors.New("Session duplicate start")
)

type PacketProtocol interface {
	NewWriter() PacketWriter

	NewReader() PacketReader
}

type PacketWriter interface {
	BeginPacket(size uint, buff []byte) []byte

	EndPacket([]byte) []byte

	WritePacket(net.Conn, []byte) error

	SetTimeout(time.Duration)

	SetMaxSize(uint)
}

type PacketReader interface {
	ReadPacket(net.Conn, []byte) ([]byte, error)

	SetTimeout(time.Duration)

	SetMaxSize(uint)
}

type Message interface {
	RecommendPacketSize() uint

	AppendToPacket([]byte) []byte
}

// Request handler.
type RequestHandler interface {
	// Handle a request from session.
	Handle(*Session, []byte)
}

type requestHandlerFunc struct {
	callback func(*Session, []byte)
}

func (handler requestHandlerFunc) Handle(session *Session, request []byte) {
	handler.callback(session, request)
}
