package link

import (
	"encoding/binary"
	"io"
	"net"
)

// The packet spliting protocol like Erlang's {packet, N}.
// Each packet has a fix length packet header to present packet length.
type SimpleProtocol struct {
	MaxPacketSize int
	n             int
	bo            binary.ByteOrder
	bf            BufferFactory
}

// Create a {packet, N} protocol.
// The n means how many bytes of the packet header.
// The 'bo' used to define packet header's byte order.
func PacketN(n int, bo binary.ByteOrder, bf BufferFactory) *SimpleProtocol {
	return &SimpleProtocol{
		n:  n,
		bo: bo,
		bf: bf,
	}
}

// Get buffer factory.
func (p SimpleProtocol) BufferFactory() BufferFactory {
	return p.bf
}

// Create a packet writer.
func (p SimpleProtocol) NewWriter() PacketWriter {
	w := NewSimplePacketWriter(p.n, p.bo)
	w.MaxPacketSize = p.MaxPacketSize
	return w
}

// Create a packet reader.
func (p SimpleProtocol) NewReader() PacketReader {
	r := NewSimplePacketReader(p.n, p.bo)
	r.MaxPacketSize = p.MaxPacketSize
	return r
}

// The {packet, N} writer.
type SimplePacketWriter struct {
	MaxPacketSize int
	head          []byte
	encodeHead    func(int)
}

// Create a new instance of {packet, N} writer.
// The n means how many bytes of the packet header.
// The 'byteOrder' used to define packet header's byte order.
func NewSimplePacketWriter(n int, byteOrder binary.ByteOrder) *SimplePacketWriter {
	w := &SimplePacketWriter{
		head: make([]byte, n),
	}

	switch n {
	case 1:
		w.encodeHead = func(size int) {
			w.head[0] = byte(size)
		}
	case 2:
		w.encodeHead = func(size int) {
			byteOrder.PutUint16(w.head, uint16(size))
		}
	case 4:
		w.encodeHead = func(size int) {
			byteOrder.PutUint32(w.head, uint32(size))
		}
	case 8:
		w.encodeHead = func(size int) {
			byteOrder.PutUint64(w.head, uint64(size))
		}
	default:
		panic("unsupported packet head size")
	}

	return w
}

// Write a packet to the conn.
func (w *SimplePacketWriter) WritePacket(conn net.Conn, buffer OutBuffer) error {
	size := buffer.Len()

	if w.MaxPacketSize > 0 && size > w.MaxPacketSize {
		return PacketTooLargeError
	}

	w.encodeHead(size)

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
type SimplePacketReader struct {
	MaxPacketSize int
	head          []byte
	decodeHead    func() int
}

// Create a new instance of {packet, N} reader.
// The n means how many bytes of the packet header.
// The 'byteOrder' used to define packet header's byte order.
func NewSimplePacketReader(n int, byteOrder binary.ByteOrder) *SimplePacketReader {
	r := &SimplePacketReader{
		head: make([]byte, n),
	}

	switch n {
	case 1:
		r.decodeHead = func() int {
			return int(r.head[0])
		}
	case 2:
		r.decodeHead = func() int {
			return int(byteOrder.Uint16(r.head))
		}
	case 4:
		r.decodeHead = func() int {
			return int(byteOrder.Uint32(r.head))
		}
	case 8:
		r.decodeHead = func() int {
			return int(byteOrder.Uint64(r.head))
		}
	default:
		panic("unsupported packet head size")
	}

	return r
}

// Read a packet from conn.
func (r *SimplePacketReader) ReadPacket(conn net.Conn, buffer InBuffer) error {
	if _, err := io.ReadFull(conn, r.head); err != nil {
		return err
	}

	size := r.decodeHead()

	if size == 0 {
		return nil
	}

	if r.MaxPacketSize > 0 && size > r.MaxPacketSize {
		return PacketTooLargeError
	}

	buffer.Prepare(size)

	if _, err := io.ReadFull(conn, buffer.Get()); err != nil {
		return err
	}

	return nil
}
