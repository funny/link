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
	BigEndian     = binary.BigEndian
	LittleEndian  = binary.LittleEndian
	DefaultBuffer = SimpleBufferFactory{}
)

// Packet spliting protocol.
// You can implement custom packet protocol for special protocol.
type Protocol interface {
	// Get buffer factory.
	BufferFactory() BufferFactory

	// Prepare out buffer.
	Prepare(buffer Buffer, message Message)

	// Write a packet to the conn.
	Write(writer io.Writer, buffer Buffer) error

	// Read a packet from conn.
	// If the packet size large than the buffer capacity, a new buffer will be created otherwise the buffer will be reused.
	Read(reader io.Reader, buffer Buffer) error
}

// Message buffer factory.
// You can implement custom buffer type for message encrypt or check sum calculate.
type BufferFactory interface {
	// Create a message buffer.
	NewBuffer() Buffer
}

// Message buffer base interface.
// You can implement custom buffer type for message encrypt or check sum calculate.
type Buffer interface {
	// Get internal buffer.
	Data() []byte

	// Get buffer length.
	Len() int

	// Get buffer capacity.
	Cap() int

	// Prepare buffer for next read.
	// DO NOT use this method in application!
	PrepareRead(size int)

	// Prepare buffer for next write.
	// DO NOT use this method in application!
	PrepareWrite(size int)

	// Append p into buffer.
	Append(p ...byte)

	// Ignore some bytes.
	Ignore(n int)

	io.Writer
}
