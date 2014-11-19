package link

import (
	"encoding/binary"
	"io"
	"net"
)

// The packet spliting protocol like Erlang's {packet, N}.
// Each packet has a fix length packet header to present packet length.
type PNProtocol struct {
	n  uint
	bo binary.ByteOrder
}

// Create a {packet, N} protocol.
// The n means how many bytes of the packet header.
// The 'bo' used to define packet header's byte order.
func PacketN(n uint, bo binary.ByteOrder) *PNProtocol {
	return &PNProtocol{
		n:  n,
		bo: bo,
	}
}

// Create a packet writer.
func (p PNProtocol) NewWriter() PacketWriter {
	return NewPNWriter(p.n, p.bo)
}

// Create a packet reader.
func (p PNProtocol) NewReader() PacketReader {
	return NewPNReader(p.n, p.bo)
}

// The {packet, N} writer.
type PNWriter struct {
	SimpleSettings
	n  uint
	bo binary.ByteOrder
}

// Create a new instance of {packet, N} writer.
// The n means how many bytes of the packet header.
// The 'bo' used to define packet header's byte order.
func NewPNWriter(n uint, bo binary.ByteOrder) *PNWriter {
	return &PNWriter{
		n:  n,
		bo: bo,
	}
}

// Begin a packet writing on the buff.
// If the size large than the buff capacity, the buff will be dropped and a new buffer will be created.
// This method give the session a way to reuse buffer and avoid invoke Conn.Writer() twice.
func (w *PNWriter) BeginPacket(size uint, buffer *OutMessage) {
	packetLen := w.n + size
	if uint(cap(*buffer)) < packetLen {
		*buffer = make([]byte, w.n, packetLen)
	} else {
		*buffer = (*buffer)[0:w.n:cap(*buffer)]
	}
}

// Finish a packet writing.
// Give the protocol writer a chance to set packet head data after packet body writed.
func (w *PNWriter) EndPacket(buffer *OutMessage) {
	size := uint(len(*buffer)) - w.n

	if w.maxsize > 0 && size > w.maxsize {
		panic("too large packet")
	}

	switch w.n {
	case 1:
		(*buffer)[0] = byte(size)
	case 2:
		w.bo.PutUint16(*buffer, uint16(size))
	case 4:
		w.bo.PutUint32(*buffer, uint32(size))
	case 8:
		w.bo.PutUint64(*buffer, uint64(size))
	default:
		panic("unsupported packet head size")
	}
}

// Write a packet to the conn.
func (w *PNWriter) WritePacket(conn net.Conn, packet OutMessage) error {
	if _, err := conn.Write(packet); err != nil {
		return err
	}
	return nil
}

// The {packet, N} reader.
type PNReader struct {
	SimpleSettings
	n    uint
	bo   binary.ByteOrder
	head []byte
}

// Create a new instance of {packet, N} reader.
// The n means how many bytes of the packet header.
// The 'bo' used to define packet header's byte order.
func NewPNReader(n uint, bo binary.ByteOrder) *PNReader {
	return &PNReader{
		n:    n,
		bo:   bo,
		head: make([]byte, n),
	}
}

// Read a packet from conn.
func (r *PNReader) ReadPacket(conn net.Conn, buffer *InMessage) error {
	if _, err := io.ReadFull(conn, r.head); err != nil {
		return err
	}

	size := uint(0)

	switch r.n {
	case 1:
		size = uint(r.head[0])
	case 2:
		size = uint(r.bo.Uint16(r.head))
	case 4:
		size = uint(r.bo.Uint32(r.head))
	case 8:
		size = uint(r.bo.Uint64(r.head))
	default:
		panic("unsupported packet head size")
	}

	if r.maxsize > 0 && size > r.maxsize {
		return PacketTooLargeError
	}

	if uint(cap(*buffer)) >= size {
		*buffer = (*buffer)[0:size]
	} else {
		*buffer = make([]byte, size)
	}

	if size == 0 {
		return nil
	}

	_, err := io.ReadFull(conn, *buffer)
	if err != nil {
		return err
	}

	return nil
}
