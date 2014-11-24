package link

import (
	"encoding/binary"
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
	// Big endian byte order.
	BigEndianBO = binary.BigEndian
	// Big endian buffer factory.
	BigEndianBF = BufferFactoryBE{}

	// Little endian byte order.
	LittleEndianBO = binary.LittleEndian
	// Little endian buffer factory.
	LittleEndianBF = BufferFactoryLE{}
)

type Settings interface {
	// Set max packet size and returns old size limitation.
	// Set 0 means unlimit.
	MaxPacketSize(int) int
}

// Packet spliting protocol.
type PacketProtocol interface {
	// Get buffer factory.
	BufferFactory() BufferFactory

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

// Incoming message buffer.
type InBuffer interface {
	// Get internal buffer data.
	Get() []byte

	// Set internal buffer data.
	Set([]byte)

	// Prepare buffer for next read.
	Prepare(size int)

	// Copy data.
	Copy() []byte

	// Slice some bytes from buffer.
	ReadSlice(n int) []byte

	// Copy some bytes from buffer.
	ReadBytes(n int) []byte

	// Read a string from buffer.
	ReadString(n int) string

	// Read a rune from buffer.
	ReadRune() rune

	// Read a byte value from buffer.
	ReadByte() byte

	// Read a int8 value from buffer.
	ReadInt8() int8

	// Read a uint8 value from buffer.
	ReadUint8() uint8

	// Read a int16 value from buffer.
	ReadInt16() int16

	// Read a uint16 value from buffer.
	ReadUint16() uint16

	// Read a int32 value from buffer.
	ReadInt32() int32

	// Read a uint32 value from buffer.
	ReadUint32() uint32

	// Read a int64 value from buffer.
	ReadInt64() int64

	// Read a uint64 value from buffer.
	ReadUint64() uint64
}

// Outgoing messsage buffer.
type OutBuffer interface {
	// Get internal buffer.
	Get() []byte

	// Set internal buffer.
	Set([]byte)

	// Get message length.
	Len() int

	// Prepare buffer for next write.
	Prepare(head, size int)

	// Copy data.
	Copy() []byte

	// Append a byte slice into buffer.
	AppendBytes(d []byte)

	// Append a string into buffer.
	AppendString(s string)

	// Append a rune into buffer.
	AppendRune(r rune)

	// Append a byte value into buffer.
	AppendByte(v byte)

	// Append a int8 value into buffer.
	AppendInt8(v int8)

	// Append a uint8 value into buffer.
	AppendUint8(v uint8)

	// Append a int16 value into buffer.
	AppendInt16(v int16)

	// Append a uint16 value into buffer.
	AppendUint16(v uint16)

	// Append a int32 value into buffer.
	AppendInt32(v int32)

	// Append a uint32 value into buffer.
	AppendUint32(v uint32)

	// Append a int64 value into buffer.
	AppendInt64(v int64)

	// Append a uint64 value into buffer.
	AppendUint64(v uint64)
}
