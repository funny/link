package link

import (
	"encoding/binary"
	"io"
)

var (
	BigEndian    = binary.BigEndian
	LittleEndian = binary.LittleEndian
)

type Packet *OutBuffer

// Packet spliting protocol.
// You can implement custom packet protocol for special protocol.
type Protocol interface {
	// Packet a message into buffer. The buffer maybe grows.
	Packet(message Message, buffer *OutBuffer) (Packet, error)

	// Write a packet. The buffer maybe grows.
	Write(writer io.Writer, packet Packet) error

	// Read a packet. The buffer maybe grows.
	Read(reader io.Reader, buffer *InBuffer) error
}

// The packet spliting protocol like Erlang's {packet, N}.
// Each packet has a fix length packet header to present packet length.
type SimpleProtocol struct {
	n             int
	bo            binary.ByteOrder
	encodeHead    func([]byte)
	decodeHead    func([]byte) int
	MaxPacketSize int
}

// Create a {packet, N} protocol.
// The n means how many bytes of the packet header.
func PacketN(n int, byteOrder binary.ByteOrder) *SimpleProtocol {
	protocol := &SimpleProtocol{
		n:  n,
		bo: byteOrder,
	}

	switch n {
	case 1:
		protocol.encodeHead = func(buffer []byte) {
			buffer[0] = byte(len(buffer) - n)
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(buffer[0])
		}
	case 2:
		protocol.encodeHead = func(buffer []byte) {
			byteOrder.PutUint16(buffer, uint16(len(buffer)-n))
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(byteOrder.Uint16(buffer))
		}
	case 4:
		protocol.encodeHead = func(buffer []byte) {
			byteOrder.PutUint32(buffer, uint32(len(buffer)-n))
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(byteOrder.Uint32(buffer))
		}
	case 8:
		protocol.encodeHead = func(buffer []byte) {
			byteOrder.PutUint64(buffer, uint64(len(buffer)-n))
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(byteOrder.Uint64(buffer))
		}
	default:
		panic("unsupported packet head size")
	}

	return protocol
}

// Write a packet. The buffer maybe grows.
func (p *SimpleProtocol) Packet(message Message, buffer *OutBuffer) (Packet, error) {
	buffer.Prepare(message.RecommendBufferSize())
	buffer.Data = buffer.Data[:p.n]
	err := message.WriteBuffer(buffer)
	return Packet(buffer), err
}

// Write a packet. The buffer maybe grows.
func (p *SimpleProtocol) Write(writer io.Writer, buffer Packet) error {
	if p.MaxPacketSize > 0 && len(buffer.Data) > p.MaxPacketSize {
		return PacketTooLargeError
	}
	p.encodeHead(buffer.Data)
	if _, err := writer.Write(buffer.Data); err != nil {
		return err
	}
	return nil
}

// Read a packet. The buffer maybe grows.
func (p *SimpleProtocol) Read(reader io.Reader, buffer *InBuffer) error {
	// head
	buffer.Prepare(p.n)
	if _, err := io.ReadFull(reader, buffer.Data); err != nil {
		return err
	}
	size := p.decodeHead(buffer.Data)
	if p.MaxPacketSize > 0 && size > p.MaxPacketSize {
		return PacketTooLargeError
	}
	// body
	buffer.Prepare(size)
	if size == 0 {
		return nil
	}
	if _, err := io.ReadFull(reader, buffer.Data); err != nil {
		return err
	}
	return nil
}
