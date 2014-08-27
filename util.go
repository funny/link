package link

import (
	"bufio"
	"net"
	"sync/atomic"
	"time"
)

var dialSessionId uint64

// The easy way to setup a server.
func Listen(network, address string, protocol PacketProtocol) (*Server, error) {
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
	id := atomic.AddUint64(&dialSessionId, 1)
	session := NewSession(id, conn, protocol, DefaultSendChanSize, DefaultReadBufferSize)
	return session, nil
}

// The easy way to create a connection with timeout setting.
func DialTimeout(network, address string, timeout time.Duration, protocol PacketProtocol) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&dialSessionId, 1)
	session := NewSession(id, conn, protocol, DefaultSendChanSize, DefaultReadBufferSize)
	return session, nil
}

// This type implement the Settings interface.
// It's simple way to make your custome protocol implement Settings interface.
// See FixWriter and FixReader.
type SimpleSettings struct {
	maxsize uint
}

// Get packet size limit
func (s *SimpleSettings) GetMaxSize() uint {
	return s.maxsize
}

// Limit packet size.
func (s *SimpleSettings) SetMaxSize(maxsize uint) {
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
	size := message.RecommendPacketSize()
	packet := b.writer.BeginPacket(size, nil)
	packet = message.AppendToPacket(packet)
	packet = b.writer.EndPacket(packet)

	sessions.Fetch(func(session *Session) {
		session.TrySendPacket(packet, 0)
	})
}

// Broadcast to sessions. The message only encoded once
// so the performance it's better then send message one by one.
func (b *Broadcaster) MustBroadcast(sessions SessionCollection, message Message) {
	size := message.RecommendPacketSize()
	packet := b.writer.BeginPacket(size, nil)
	packet = message.AppendToPacket(packet)
	packet = b.writer.EndPacket(packet)

	sessions.Fetch(func(session *Session) {
		session.SendPacket(packet)
	})
}

// Buffered connection.
type BufferConn struct {
	net.Conn
	reader *bufio.Reader
}

func NewBufferConn(conn net.Conn, size int) *BufferConn {
	return &BufferConn{
		conn,
		bufio.NewReaderSize(conn, size),
	}
}

func (conn *BufferConn) Read(d []byte) (int, error) {
	return conn.reader.Read(d)
}
