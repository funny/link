package link

import (
	"encoding/binary"
	"io"
)

var (
	BigEndian    = binary.BigEndian
	LittleEndian = binary.LittleEndian
)

// Packet spliting protocol.
// You can implement custom packet protocol for special protocol.
type Protocol interface {
	// Pack a message into buffer.
	Packet(buffer []byte, message Message) ([]byte, error)

	// Write a packet to the conn.
	Write(writer io.Writer, buffer []byte) ([]byte, error)

	// Read a packet from conn.
	// If the packet size large than the buffer capacity, a new buffer will be created otherwise the buffer will be reused.
	Read(reader io.Reader, buffer []byte) ([]byte, error)
}

// The packet spliting protocol like Erlang's {packet, N}.
// Each packet has a fix length packet header to present packet length.
type SimpleProtocol struct {
	n             int
	bo            binary.ByteOrder
	head          []byte
	encodeHead    func(int, []byte)
	decodeHead    func() int
	MaxPacketSize int
}

// Create a {packet, N} protocol.
// The n means how many bytes of the packet header.
// The 'bo' used to define packet header's byte order.
func PacketN(n int, byteOrder binary.ByteOrder) *SimpleProtocol {
	protocol := &SimpleProtocol{
		n:    n,
		bo:   byteOrder,
		head: make([]byte, n),
	}

	switch n {
	case 1:
		protocol.encodeHead = func(size int, buffer []byte) {
			buffer[0] = byte(size)
		}
		protocol.decodeHead = func() int {
			return int(protocol.head[0])
		}
	case 2:
		protocol.encodeHead = func(size int, buffer []byte) {
			byteOrder.PutUint16(buffer, uint16(size-n))
		}
		protocol.decodeHead = func() int {
			return int(byteOrder.Uint16(protocol.head))
		}
	case 4:
		protocol.encodeHead = func(size int, buffer []byte) {
			byteOrder.PutUint32(buffer, uint32(size-n))
		}
		protocol.decodeHead = func() int {
			return int(byteOrder.Uint32(protocol.head))
		}
	case 8:
		protocol.encodeHead = func(size int, buffer []byte) {
			byteOrder.PutUint64(buffer, uint64(size-n))
		}
		protocol.decodeHead = func() int {
			return int(byteOrder.Uint64(protocol.head))
		}
	default:
		panic("unsupported packet head size")
	}

	return protocol
}

func (p *SimpleProtocol) Packet(buffer []byte, message Message) ([]byte, error) {
	if cap(buffer) < message.RecommendBufferSize() {
		return message.WriteBuffer(make([]byte, p.n, message.RecommendBufferSize()))
	}
	return message.WriteBuffer(buffer[:p.n])
}

// Write a packet to the conn.
func (p *SimpleProtocol) Write(writer io.Writer, buffer []byte) ([]byte, error) {
	if p.MaxPacketSize > 0 && len(buffer) > p.MaxPacketSize {
		return nil, PacketTooLargeError
	}

	p.encodeHead(len(buffer), buffer)

	if _, err := writer.Write(buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}

// Read a packet into buffer.
func (p *SimpleProtocol) Read(reader io.Reader, buffer []byte) ([]byte, error) {
	if _, err := io.ReadFull(reader, p.head); err != nil {
		return nil, err
	}

	size := p.decodeHead()

	if p.MaxPacketSize > 0 && size > p.MaxPacketSize {
		return nil, PacketTooLargeError
	}

	if cap(buffer) < size {
		buffer = make([]byte, size)
	} else {
		buffer = buffer[0:size]
	}

	if size == 0 {
		return buffer, nil
	}

	if _, err := io.ReadFull(reader, buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}
