package link

import (
	"encoding/binary"
	"errors"
	"io"
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
	BigEndian    = binary.BigEndian
	LittleEndian = binary.LittleEndian
	SimpleBuffer = SimpleBufferFactory{}
)

// Packet spliting protocol.
// You can implement custom packet protocol for special protocol.
type PacketProtocol interface {
	// Get buffer factory.
	BufferFactory() BufferFactory

	// Create a packet writer.
	NewWriter() PacketWriter

	// Create a packet reader.
	NewReader() PacketReader
}

// Packet writer.
// You can implement custom packet protocol for special protocol.
type PacketWriter interface {
	// Write a packet to the conn.
	WritePacket(conn net.Conn, buffer OutBuffer) error
}

// Packet reader.
// You can implement custom packet protocol for special protocol.
type PacketReader interface {
	// Read a packet from conn.
	// If the packet size large than the buffer capacity, a new buffer will be created otherwise the buffer will be reused.
	ReadPacket(conn net.Conn, buffer InBuffer) error
}

// Message buffer factory.
// You can implement custom buffer type for message encrypt or check sum calculate.
type BufferFactory interface {
	// Create a incoming message buffer.
	NewInBuffer() InBuffer

	// Create a outgoing message buffer.
	NewOutBuffer() OutBuffer
}

// Message buffer base interface.
// You can implement custom buffer type for message encrypt or check sum calculate.
type Buffer interface {
	// Get internal buffer.
	Get() []byte

	// Get buffer length.
	Len() int

	// Get buffer capacity.
	Cap() int

	// Copy buffer data.
	Copy() []byte

	// Prepare buffer for next read.
	// DO NOT use this method in application!
	Prepare(size int)
}

// Incoming message buffer.
// You can implement custom buffer type for message encrypt or check sum calculate.
type InBuffer interface {
	Buffer

	io.Reader

	// Slice some bytes from buffer.
	ReadSlice(n int) []byte

	// Copy some bytes from buffer.
	ReadBytes(n int) []byte

	// Read a string from buffer.
	ReadString(n int) string

	// Read a rune from buffer.
	ReadRune() rune

	// Read a float32 value from buffer.
	ReadFloat32() float32

	// Read a float64 value from buffer.
	ReadFloat64() float64

	// Read a uint8 value from buffer.
	ReadUint8() uint8

	// Read a uint16 value from buffer.
	ReadUint16LE() uint16

	// Read a uint16 value from buffer.
	ReadUint16BE() uint16

	// Read a uint32 value from buffer.
	ReadUint32LE() uint32

	// Read a uint32 value from buffer.
	ReadUint32BE() uint32

	// Read a uint64 value from buffer.
	ReadUint64LE() uint64

	// Read a uint64 value from buffer.
	ReadUint64BE() uint64
}

// Outgoing messsage buffer.
// You can implement custom buffer type for message encrypt or check sum calculate.
type OutBuffer interface {
	Buffer

	io.Writer

	// Write a byte slice into buffer.
	WriteBytes(d []byte)

	// Write a string into buffer.
	WriteString(s string)

	// Write a rune into buffer.
	WriteRune(r rune)

	// Write a float32 value into buffer.
	WriteFloat32(v float32)

	// Write a float64 value into buffer.
	WriteFloat64(v float64)

	// Write a uint8 value into buffer.
	WriteUint8(v uint8)

	// Write a uint16 value into buffer.
	WriteUint16LE(v uint16)

	// Write a uint16 value into buffer.
	WriteUint16BE(v uint16)

	// Write a uint32 value into buffer.
	WriteUint32LE(v uint32)

	// Write a uint32 value into buffer.
	WriteUint32BE(v uint32)

	// Write a uint64 value into buffer.
	WriteUint64LE(v uint64)

	// Write a uint64 value into buffer.
	WriteUint64BE(v uint64)
}
