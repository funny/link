package link

import (
	"encoding/binary"
	"io"
	"net"
	"time"
)

// The packet spliting protocol like Erlang's {packet, N}.
// Each packet has a fix length packet header to present packet length.
type FixProtocol struct {
	n  uint
	bo binary.ByteOrder
}

// Create a {packet, N} protocol.
// The n means how many bytes of the packet header.
// The 'bo' used to define packet header's byte order.
func NewFixProtocol(n uint, bo binary.ByteOrder) *FixProtocol {
	return &FixProtocol{
		n:  n,
		bo: bo,
	}
}

// Create a packet writer.
func (p FixProtocol) NewWriter() PacketWriter {
	return NewFixWriter(p.n, p.bo)
}

// Create a packet reader.
func (p FixProtocol) NewReader() PacketReader {
	return NewFixReader(p.n, p.bo)
}

// The {packet, N} writer.
type FixWriter struct {
	SimpleSetting
	n  uint
	bo binary.ByteOrder
}

// Create a new instance of {packet, N} writer.
// The n means how many bytes of the packet header.
// The 'bo' used to define packet header's byte order.
func NewFixWriter(n uint, bo binary.ByteOrder) *FixWriter {
	return &FixWriter{
		n:  n,
		bo: bo,
	}
}

// Begin a packet writing on the buff.
// If the size large than the buff capacity, the buff will be dropped and a new buffer will be created.
// This method give the session a way to reuse buffer and avoid invoke Conn.Writer() twice.
func (w *FixWriter) BeginPacket(size uint, buff []byte) []byte {
	packetLen := w.n + size
	if uint(cap(buff)) < packetLen {
		return make([]byte, w.n, packetLen)
	}
	return buff[0:w.n:packetLen]
}

// Finish a packet writing.
// Give the protocol writer a chance to set packet head data after packet body writed.
func (w *FixWriter) EndPacket(packet []byte) []byte {
	size := uint(len(packet)) - w.n

	if w.maxsize > 0 && size > w.maxsize {
		panic("too large packet")
	}

	switch w.n {
	case 1:
		packet[0] = byte(size)
	case 2:
		w.bo.PutUint16(packet, uint16(size))
	case 4:
		w.bo.PutUint32(packet, uint32(size))
	case 8:
		w.bo.PutUint64(packet, uint64(size))
	default:
		panic("unsupported packet head size")
	}

	return packet
}

// Write a packet to the conn.
func (w *FixWriter) WritePacket(conn net.Conn, packet []byte) error {
	if w.timeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(w.timeout))
	} else {
		conn.SetWriteDeadline(time.Time{})
	}

	if _, err := conn.Write(packet); err != nil {
		return err
	}

	return nil
}

// The {packet, N} reader.
type FixReader struct {
	SimpleSetting
	n    uint
	bo   binary.ByteOrder
	head []byte
}

// Create a new instance of {packet, N} reader.
// The n means how many bytes of the packet header.
// The 'bo' used to define packet header's byte order.
func NewFixReader(n uint, bo binary.ByteOrder) *FixReader {
	return &FixReader{
		n:    n,
		bo:   bo,
		head: make([]byte, n),
	}
}

// Read a packet from conn.
func (r *FixReader) ReadPacket(conn net.Conn, b []byte) ([]byte, error) {
	if r.timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(r.timeout))
	}

	if _, err := io.ReadFull(conn, r.head); err != nil {
		return nil, err
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
		return nil, PacketTooLargeError
	}

	var data []byte

	if uint(cap(b)) >= size {
		data = b[0:size]
	} else {
		data = make([]byte, size)
	}

	if len(data) == 0 {
		return data, nil
	}

	if r.timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(r.timeout))
	}

	_, err := io.ReadFull(conn, data)

	return data, err
}
