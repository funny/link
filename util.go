package link

import (
	"bufio"
	"bytes"
	"encoding/json"
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
// For example, sometimes you have many Session.Send() call during a request processing.
// You can use the send queue to buffer those messages then call Session.Send() once after request processing done.
// The send queue type implemented Message interface. So you can pass it as the Session.Send() method argument.
type SendQueue struct {
	messages []Message
}

// Push a message into send queue but not send it immediately.
func (q *SendQueue) Push(message Message) {
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
func (q *SendQueue) AppendToPacket(packet []byte) ([]byte, error) {
	var err error
	for _, message := range q.messages {
		packet, err = message.AppendToPacket(packet)
		if err != nil {
			return nil, err
		}
	}
	return packet, nil
}

// The session collection use to fetch session and send broadcast.
type SessionCollection interface {
	Fetch(func(*Session))
}

// A broadcast sender. The broadcast message only encoded once
// so the performance it's better then send message one by one.
type Broadcaster struct {
	writer PacketWriter
}

// Craete a broadcaster.
func NewBroadcaster(writer PacketWriter) *Broadcaster {
	return &Broadcaster{
		writer: writer,
	}
}

func (b *Broadcaster) packet(message Message) (packet []byte, err error) {
	size := message.RecommendPacketSize()
	packet = b.writer.BeginPacket(size, nil)
	packet, err = message.AppendToPacket(packet)
	if err != nil {
		return nil, err
	}
	packet = b.writer.EndPacket(packet)
	return
}

// Broadcast to sessions. The message only encoded once
// so the performance it's better then send message one by one.
func (b *Broadcaster) Broadcast(sessions SessionCollection, message Message) error {
	packet, err := b.packet(message)
	if err != nil {
		return err
	}
	sessions.Fetch(func(session *Session) {
		session.TrySendPacket(packet, 0)
	})
	return nil
}

// Broadcast to sessions. The message only encoded once
// so the performance it's better then send message one by one.
func (b *Broadcaster) MustBroadcast(sessions SessionCollection, message Message) error {
	packet, err := b.packet(message)
	if err != nil {
		return err
	}
	sessions.Fetch(func(session *Session) {
		session.SendPacket(packet)
	})
	return nil
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

// Binary message
type Binary []byte

// Implement the Message interface.
func (bin Binary) RecommendPacketSize() uint {
	return uint(len(bin))
}

// Implement the Message interface.
func (bin Binary) AppendToPacket(packet []byte) ([]byte, error) {
	return append(packet, bin...), nil
}

// JSON message
type JSON struct {
	V    interface{}
	Size uint
}

// Implement the Message interface.
func (j JSON) RecommendPacketSize() uint {
	return j.Size
}

// Implement the Message interface.
func (j JSON) AppendToPacket(packet []byte) ([]byte, error) {
	w := bytes.NewBuffer(packet)
	e := json.NewEncoder(w)
	err := e.Encode(j.V)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
