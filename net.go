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
	SyncSendTimeoutError       = errors.New("Sync send timeout")
)

// Packet spliting protocol.
type PacketProtocol interface {
	// Create a packet writer.
	NewWriter() PacketWriter

	// Create a packet reader.
	NewReader() PacketReader
}

// Packet writer.
type PacketWriter interface {
	// Begin a packet writing on the buff.
	// If the size large than the buff capacity, the buff will be dropped and a new buffer will be created.
	// This method give the session a way to reuse buffer and avoid invoke Conn.Writer() twice.
	BeginPacket(size uint, buff []byte) []byte

	// Finish a packet writing.
	// Give the protocol writer a chance to set packet head data after packet body writed.
	EndPacket([]byte) []byte

	// Write a packet to the conn.
	WritePacket(net.Conn, []byte) error

	// Set write timeout.
	SetTimeout(time.Duration)

	// Limit packet size.
	SetMaxSize(uint)
}

// Packet reader.
type PacketReader interface {
	// Create a new instance of {packet, N} reader.
	// The n means how many bytes of the packet header used to present packet length.
	// The 'bo' used to define packet header's byte order.
	ReadPacket(net.Conn, []byte) ([]byte, error)

	// Set read timeout.
	SetTimeout(time.Duration)

	// Limit packet size.
	SetMaxSize(uint)
}

// Message.
type Message interface {
	// Get a recommend packet size for packet buffer initialization.
	RecommendPacketSize() uint

	// Append the message to the packet buffer and returns the new buffer like append() function.
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
