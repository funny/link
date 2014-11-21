package link

import (
	"errors"
	"net"
)

// Errors
var (
	SendToClosedError   = errors.New("Send to closed session")
	BlockingError       = errors.New("Blocking happened")
	PacketTooLargeError = errors.New("Packet too large")
	NilBufferError      = errors.New("Buffer is nil")
)

var (
	BigEndian    = BufferFactoryBE{}
	LittleEndian = BufferFactoryLE{}
)

type Settings interface {
	// Set max packet size and returns old size limitation.
	// Set 0 means unlimit.
	MaxPacketSize(int) int
}

// Packet spliting protocol.
type PacketProtocol interface {
	// Create a packet writer.
	NewWriter() PacketWriter

	// Create a packet reader.
	NewReader() PacketReader
}

// Packet writer.
type PacketWriter interface {
	Settings

	// Begin a packet writing on the buffer.
	// If the packet size large than the buffer capacity, a new buffer will be created otherwise the buffer will be reused.
	// The size no need to equals really packet size, some time we could not knows a message's packet size before it encoded,
	// if the size less than really packet size, the buffer will auto grows when you append data into it.
	// This method give the session a way to reuse buffer and avoid invoke Write() twice.
	BeginPacket(size int, buffer OutBuffer)

	// Finish a packet writing.
	// Give the protocol writer a chance to set packet head data after packet body writed.
	EndPacket(buffer OutBuffer)

	// Write a packet to the conn.
	WritePacket(conn net.Conn, buffer OutBuffer) error
}

// Packet reader.
type PacketReader interface {
	Settings

	// Read a packet from conn.
	// If the packet size large than the buffer capacity, a new buffer will be created otherwise the buffer will be reused.
	ReadPacket(conn net.Conn, buffer InBuffer) error
}

// Message buffer factory.
type BufferFactory interface {
	// Create a incoming message buffer.
	NewInBuffer() InBuffer

	// Create a outgoing message buffer.
	NewOutBuffer() OutBuffer
}

// Big endian message buffer factory.
type BufferFactoryBE struct {
}

// Create a big endian incoming message buffer.
func (_ BufferFactoryBE) NewInBuffer() InBuffer {
	return new(InBufferBE)
}

// Create a big endian outgoing message buffer.
func (_ BufferFactoryBE) NewOutBuffer() OutBuffer {
	return new(OutBufferBE)
}

// Little endian message buffer factory.
type BufferFactoryLE struct {
}

// Create a little endian incoming message buffer.
func (_ BufferFactoryLE) NewInBuffer() InBuffer {
	return new(InBufferLE)
}

// Create a little endian outgoing message buffer.
func (_ BufferFactoryLE) NewOutBuffer() OutBuffer {
	return new(OutBufferLE)
}
