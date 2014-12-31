package link

import (
	"encoding/binary"
	"io"
)

var (
	BigEndian    = ByteOrder(binary.BigEndian)
	LittleEndian = ByteOrder(binary.LittleEndian)

	packet1BE = newSimpleProtocol(1, BigEndian)
	packet1LE = newSimpleProtocol(1, LittleEndian)
	packet2BE = newSimpleProtocol(2, BigEndian)
	packet2LE = newSimpleProtocol(2, LittleEndian)
	packet4BE = newSimpleProtocol(4, BigEndian)
	packet4LE = newSimpleProtocol(4, LittleEndian)
	packet8BE = newSimpleProtocol(8, BigEndian)
	packet8LE = newSimpleProtocol(8, LittleEndian)
)

type ByteOrder binary.ByteOrder

type Packet struct {
	*OutBuffer
}

// Packet spliting protocol.
// You can implement custom packet protocol for special protocol.
type Protocol interface {
	// Create a instance.
	New() Protocol

	// Packet a message.
	Packet(message Message) (Packet, error)

	// Write a packet.
	Write(writer io.Writer, packet Packet) error

	// Read a packet.
	Read(reader io.Reader) (*InBuffer, error)
}

// Create a {packet, N} protocol.
// The n means how many bytes of the packet header.
// n must is 1、2、4 or 8.
func PacketN(n int, byteOrder ByteOrder) Protocol {
	switch n {
	case 1:
		switch byteOrder {
		case BigEndian:
			return packet1BE
		case LittleEndian:
			return packet1LE
		}
	case 2:
		switch byteOrder {
		case BigEndian:
			return packet2BE
		case LittleEndian:
			return packet2LE
		}
	case 4:
		switch byteOrder {
		case BigEndian:
			return packet4BE
		case LittleEndian:
			return packet4LE
		}
	case 8:
		switch byteOrder {
		case BigEndian:
			return packet8BE
		case LittleEndian:
			return packet8LE
		}
	}
	panic("unsupported packet head size")
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

// Create simple protocol.
func newSimpleProtocol(n int, byteOrder binary.ByteOrder) *SimpleProtocol {
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

// SimpleProtocol no need to create new instance.
func (p *SimpleProtocol) New() Protocol {
	return p
}

// Write a packet.
func (p *SimpleProtocol) Packet(message Message) (Packet, error) {
	buffer := NewOutBuffer()
	buffer.Prepare(message.RecommendBufferSize())
	buffer.Data = buffer.Data[:p.n]
	err := message.WriteBuffer(buffer)
	return Packet{buffer}, err
}

// Write a packet.
func (p *SimpleProtocol) Write(writer io.Writer, packet Packet) error {
	if p.MaxPacketSize > 0 && len(packet.Data) > p.MaxPacketSize {
		return PacketTooLargeError
	}
	p.encodeHead(packet.Data)
	if _, err := writer.Write(packet.Data); err != nil {
		return err
	}
	return nil
}

// Read a packet.
func (p *SimpleProtocol) Read(reader io.Reader) (*InBuffer, error) {
	// head
	buffer := NewInBuffer()
	buffer.Prepare(p.n)
	if _, err := io.ReadFull(reader, buffer.Data); err != nil {
		return nil, err
	}
	size := p.decodeHead(buffer.Data)
	if p.MaxPacketSize > 0 && size > p.MaxPacketSize {
		return nil, PacketTooLargeError
	}
	// body
	buffer.Prepare(size)
	if size == 0 {
		return buffer, nil
	}
	if _, err := io.ReadFull(reader, buffer.Data); err != nil {
		return nil, err
	}
	return buffer, nil
}
