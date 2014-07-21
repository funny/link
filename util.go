package packnet

import (
	"net"
	"time"
)

// The easy way to setup a server.
func ListenAndServe(network, address string, protocol PacketProtocol) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	return NewServer(listener, protocol), nil
}

// The easy way to create a connection.
func Dial(network, address string, protocol PacketProtocol, id uint64, sendChanSize uint) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	session := NewSession(id, conn, protocol.NewWriter(), protocol.NewReader(), sendChanSize)

	return session, nil
}

// The easy way to create a connection with timeout setting.
func DialTimeout(network, address string, timeout time.Duration, protocol PacketProtocol, id uint64, sendChanSize uint) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}

	session := NewSession(id, conn, protocol.NewWriter(), protocol.NewReader(), sendChanSize)

	return session, nil
}

type SendQueue struct {
	session  *Session
	messages []Message
}

func (q *SendQueue) Send(message Message) {
	q.messages = append(q.messages, message)
}

func (q *SendQueue) RecommendPacketSize() uint {
	size := uint(0)
	for _, message := range q.messages {
		size += message.RecommendPacketSize()
	}
	return size
}

func (q *SendQueue) AppendToPacket(packet []byte) []byte {
	for _, message := range q.messages {
		packet = message.AppendToPacket(packet)
	}
	return packet
}
