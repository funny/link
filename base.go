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

type Settings interface {
	// Set max packet size and returns old size limitation.
	// Set 0 means unlimit.
	MaxPacketSize(uint) uint
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

	// Begin a packet writing on the buff.
	// If the size large than the buff capacity, the buff will be dropped and a new buffer will be created.
	// This method give the session a way to reuse buffer and avoid invoke Write() twice.
	BeginPacket(size uint, buffer *OutMessage)

	// Finish a packet writing.
	// Give the protocol writer a chance to set packet head data after packet body writed.
	EndPacket(packet *OutMessage)

	// Write a packet to the conn.
	WritePacket(conn net.Conn, buffer OutMessage) error
}

// Packet reader.
type PacketReader interface {
	Settings

	// Read a packet from conn.
	ReadPacket(conn net.Conn, buffer *InMessage) error
}
