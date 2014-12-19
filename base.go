package link

import (
	"encoding/binary"
	"errors"
	"io"
)

// Errors
var (
	SendToClosedError   = errors.New("Send to closed session")
	BlockingError       = errors.New("Blocking happened")
	PacketTooLargeError = errors.New("Packet too large")
	NilBufferError      = errors.New("Buffer is nil")
)

var (
	BigEndian    = binary.BigEndian
	LittleEndian = binary.LittleEndian
)

// Packet spliting protocol.
// You can implement custom packet protocol for special protocol.
type Protocol interface {
	// Pack a message into buffer.
	Packet(buffer []byte, message Message) ([]byte, error)

	// Write a packet to the conn.
	Write(writer io.Writer, buffer []byte) ([]byte, error)

	// Read a packet from conn.
	// If the packet size large than the buffer capacity, a new buffer will be created otherwise the buffer will be reused.
	Read(reader io.Reader, buffer []byte) ([]byte, error)
}
