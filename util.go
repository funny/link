package link

import (
	"net"
	"sync"
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

// This type implement the Setable interface.
// It's simple way to make your custome protocol implement Setable interface.
// See FixWriter and FixReader.
type SimpleSetting struct {
	timeout time.Duration
	maxsize uint
}

// Get timeout setting.
func (s *SimpleSetting) GetTimeout() time.Duration {
	return s.timeout
}

// Set timeout.
func (s *SimpleSetting) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

// Get packet size limit
func (s *SimpleSetting) GetMaxSize() uint {
	return s.maxsize
}

// Limit packet size.
func (s *SimpleSetting) SetMaxSize(maxsize uint) {
	s.maxsize = maxsize
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

// A broadcast sender. The broadcast message only encoded once
// so the performance it's better then send message one by one.
type Broadcaster struct {
	mutex  sync.RWMutex
	buff   []byte
	writer PacketWriter
}

// The session collection use to fetch session and send broadcast.
type SessionCollection interface {
	Fetch(func(*Session))
}

// Craete a broadcaster.
func (server *Server) NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		writer: server.protocol.NewWriter(),
	}
}

// Broadcast to sessions. The message only encoded once
// so the performance it's better then send message one by one.
func (b *Broadcaster) Broadcast(sessions SessionCollection, message Message) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	size := message.RecommendPacketSize()

	packet := b.writer.BeginPacket(size, b.buff)
	packet = message.AppendToPacket(packet)
	packet = b.writer.EndPacket(packet)

	b.buff = packet

	sessions.Fetch(func(session *Session) {
		session.SendPacket(packet)
	})
}
