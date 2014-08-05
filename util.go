package packnet

import (
	"net"
	"time"
)

var dialSessionId uint64

// The easy way to setup a server.
func ListenAndServe(network, address string, protocol PacketProtocol) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	return NewServer(listener, protocol), nil
}

// The easy way to create a connection.
func Dial(network, address string, protocol PacketProtocol) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	dialSessionId += 1

	session := NewSession(dialSessionId, conn, protocol, DefaultSendChanSize)

	return session, nil
}

// The easy way to create a connection with timeout setting.
func DialTimeout(network, address string, timeout time.Duration, protocol PacketProtocol) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}

	dialSessionId += 1

	session := NewSession(dialSessionId, conn, protocol, DefaultSendChanSize)

	return session, nil
}

// A simple send queue. Can used for buffered send.
// For example, sometimes you have many Send() call during a request processing.
// You can use the send queue to buffer those messages then call Send() once after request processing done.
// The send queue type implemented Message interface. So you can pass it as the Send() method argument.
type SendQueue struct {
	messages []Message
}

// Push a message into send queue but not send it immediately.
func (q *SendQueue) Send(message Message) {
	q.messages = append(q.messages, message)
}

// Implement the Message interface.
func (q *SendQueue) RecommendPacketSize() uint {
	size := uint(0)
	for _, message := range q.messages {
		size += message.RecommendPacketSize()
	}
	return size
}

// Implement the Message interface.
func (q *SendQueue) AppendToPacket(packet []byte) []byte {
	for _, message := range q.messages {
		packet = message.AppendToPacket(packet)
	}
	return packet
}
