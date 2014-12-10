package link

import (
	"encoding/binary"
	"io"
	"net"
)

// The packet spliting protocol like Erlang's {packet, N}.
// Each packet has a fix length packet header to present packet length.
type PNProtocol struct {
	MaxPacketSize
	n  int
	bo binary.ByteOrder
	bf BufferFactory
}

// Create a {packet, N} protocol.
// The n means how many bytes of the packet header.
// The 'bo' used to define packet header's byte order.
func PacketN(n int, bo binary.ByteOrder, bf BufferFactory) *PNProtocol {
	return &PNProtocol{
		n:  n,
		bo: bo,
		bf: bf,
	}
}

// Get buffer factory.
func (p PNProtocol) BufferFactory() BufferFactory {
	return p.bf
}

// Create a packet writer.
func (p PNProtocol) NewWriter() PacketWriter {
	w := NewPNWriter(p.n, p.bo)
	w.MaxPacketSize = p.MaxPacketSize
	return w
}

// Create a packet reader.
func (p PNProtocol) NewReader() PacketReader {
	r := NewPNReader(p.n, p.bo)
	r.MaxPacketSize = p.MaxPacketSize
	return r
}

// The {packet, N} writer.
type PNWriter struct {
	MaxPacketSize
	n    int
	bo   binary.ByteOrder
	head []byte
}

// Create a new instance of {packet, N} writer.
// The n means how many bytes of the packet header.
// The 'bo' used to define packet header's byte order.
func NewPNWriter(n int, bo binary.ByteOrder) *PNWriter {
	return &PNWriter{
		n:    n,
		bo:   bo,
		head: make([]byte, n),
	}
}

// Write a packet to the conn.
func (w *PNWriter) WritePacket(conn net.Conn, buffer OutBuffer) error {
	size := buffer.Len()

	if maxsize := w.MaxPacketSize.Get(); maxsize > 0 && size > maxsize {
		return PacketTooLargeError
	}

	switch w.n {
	case 1:
		w.head[0] = byte(size)
	case 2:
		w.bo.PutUint16(w.head, uint16(size))
	case 4:
		w.bo.PutUint32(w.head, uint32(size))
	case 8:
		w.bo.PutUint64(w.head, uint64(size))
	default:
		panic("unsupported packet head size")
	}

	if _, err := conn.Write(w.head); err != nil {
		return err
	}

	if size == 0 {
		return nil
	}

	if _, err := conn.Write(buffer.Get()); err != nil {
		return err
	}

	return nil
}

// The {packet, N} reader.
type PNReader struct {
	MaxPacketSize
	n    int
	bo   binary.ByteOrder
	head []byte
}

// Create a new instance of {packet, N} reader.
// The n means how many bytes of the packet header.
// The 'bo' used to define packet header's byte order.
func NewPNReader(n int, bo binary.ByteOrder) *PNReader {
	return &PNReader{
		n:    n,
		bo:   bo,
		head: make([]byte, n),
	}
}

// Read a packet from conn.
func (r *PNReader) ReadPacket(conn net.Conn, buffer InBuffer) error {
	if _, err := io.ReadFull(conn, r.head); err != nil {
		return err
	}

	size := 0

	switch r.n {
	case 1:
		size = int(r.head[0])
	case 2:
		size = int(r.bo.Uint16(r.head))
	case 4:
		size = int(r.bo.Uint32(r.head))
	case 8:
		size = int(r.bo.Uint64(r.head))
	default:
		panic("unsupported packet head size")
	}

	if size == 0 {
		return nil
	}

	if maxsize := r.MaxPacketSize.Get(); maxsize > 0 && size > maxsize {
		return PacketTooLargeError
	}

	buffer.Prepare(size)

	if _, err := io.ReadFull(conn, buffer.Get()); err != nil {
		return err
	}

	return nil
}
