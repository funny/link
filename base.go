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
	SimpleBuffer = SimpleBufferFactory{}
)

// Packet spliting protocol.
// You can implement custom packet protocol for special protocol.
type Protocol interface {
	// Get buffer factory.
	BufferFactory() BufferFactory

	// Prepare out buffer.
	Prepare(buffer OutBuffer, message Message)

	// Write a packet to the conn.
	Write(writer io.Writer, buffer OutBuffer) error

	// Read a packet from conn.
	// If the packet size large than the buffer capacity, a new buffer will be created otherwise the buffer will be reused.
	Read(reader io.Reader, buffer InBuffer) error
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
	Slice(n int) []byte

	// Copy some bytes from buffer.
	ReadBytes(n int) []byte

	// Read a string from buffer.
	ReadString(n int) string

	// Read a rune from buffer.
	ReadRune() rune

	// Read a uint8 value from buffer.
	ReadUint8() uint8

	// Read a uint16 value from buffer using little endian byte order.
	ReadUint16LE() uint16

	// Read a uint16 value from buffer using big endian byte order.
	ReadUint16BE() uint16

	// Read a uint32 value from buffer using little endian byte order.
	ReadUint32LE() uint32

	// Read a uint32 value from buffer using big endian byte order.
	ReadUint32BE() uint32

	// Read a uint64 value from buffer using little endian byte order.
	ReadUint64LE() uint64

	// Read a uint64 value from buffer using big endian byte order.
	ReadUint64BE() uint64

	// Read a float32 value from buffer using little endian byte order.
	ReadFloat32LE() float32

	// Read a float32 value from buffer using big endian byte order.
	ReadFloat32BE() float32

	// Read a float64 value from buffer using little endian byte order.
	ReadFloat64LE() float64

	// Read a float64 value from buffer using big endian byte order.
	ReadFloat64BE() float64
}

// Outgoing messsage buffer.
// You can implement custom buffer type for message encrypt or check sum calculate.
type OutBuffer interface {
	Buffer

	io.Writer

	// Ignore some bytes.
	Ignore(n int)

	// Write a byte slice into buffer.
	WriteBytes(d []byte)

	// Write a string into buffer.
	WriteString(s string)

	// Write a rune into buffer.
	WriteRune(r rune)

	// Write a uint8 value into buffer.
	WriteUint8(v uint8)

	// Write a uint16 value into buffer using little endian byte order.
	WriteUint16LE(v uint16)

	// Write a uint16 value into buffer using big endian byte order.
	WriteUint16BE(v uint16)

	// Write a uint32 value into buffer using little endian byte order.
	WriteUint32LE(v uint32)

	// Write a uint32 value into buffer using big endian byte order.
	WriteUint32BE(v uint32)

	// Write a uint64 value into buffer using little endian byte order.
	WriteUint64LE(v uint64)

	// Write a uint64 value into buffer using big endian byte order.
	WriteUint64BE(v uint64)

	// Write a float32 value into buffer using little endian byte order.
	WriteFloat32LE(v float32)

	// Write a float32 value into buffer using big endian byte order.
	WriteFloat32BE(v float32)

	// Write a float64 value into buffer using little endian byte order.
	WriteFloat64LE(v float64)

	// Write a float64 value into buffer using big endian byte order.
	WriteFloat64BE(v float64)
}
