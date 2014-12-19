package link

import (
	"encoding/binary"
	"io"
)

// The packet spliting protocol like Erlang's {packet, N}.
// Each packet has a fix length packet header to present packet length.
type SimpleProtocol struct {
	n             int
	bo            binary.ByteOrder
	bf            BufferFactory
	head          []byte
	encodeHead    func(int, []byte)
	decodeHead    func() int
	MaxPacketSize int
}

// Create a {packet, N} protocol.
// The n means how many bytes of the packet header.
// The 'bo' used to define packet header's byte order.
func PacketN(n int, byteOrder binary.ByteOrder, bufferFactory BufferFactory) *SimpleProtocol {
	protocol := &SimpleProtocol{
		n:    n,
		bo:   byteOrder,
		bf:   bufferFactory,
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

// Get buffer factory.
func (p *SimpleProtocol) BufferFactory() BufferFactory {
	return p.bf
}

func (p *SimpleProtocol) Prepare(buffer Buffer, message Message) {
	buffer.PrepareWrite(p.n + message.RecommendBufferSize())
	buffer.Ignore(p.n)
}

// Write a packet to the conn.
func (p *SimpleProtocol) Write(writer io.Writer, buffer Buffer) error {
	var (
		size = buffer.Len()
		buff = buffer.Data()
	)

	if p.MaxPacketSize > 0 && size > p.MaxPacketSize {
		return PacketTooLargeError
	}

	p.encodeHead(size, buff)

	if _, err := writer.Write(buff); err != nil {
		return err
	}

	return nil
}

// Read a packet into buffer.
func (p *SimpleProtocol) Read(reader io.Reader, buffer Buffer) error {
	if _, err := io.ReadFull(reader, p.head); err != nil {
		return err
	}

	size := p.decodeHead()

	if p.MaxPacketSize > 0 && size > p.MaxPacketSize {
		return PacketTooLargeError
	}

	buffer.PrepareRead(size)

	if size == 0 {
		return nil
	}

	if _, err := io.ReadFull(reader, buffer.Data()); err != nil {
		return err
	}

	return nil
}
