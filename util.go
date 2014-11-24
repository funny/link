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
	session := NewSession(id, conn, protocol, DefaultSendChanSize, DefaultConnBufferSize)
	return session, nil
}

// The easy way to create a connection with timeout setting.
func DialTimeout(network, address string, timeout time.Duration, protocol PacketProtocol) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&dialSessionId, 1)
	session := NewSession(id, conn, protocol, DefaultSendChanSize, DefaultConnBufferSize)
	return session, nil
}

// This type implement the Settings interface.
// It's simple way to make your custome protocol implement Settings interface.
// See FixWriter and FixReader.
type SimpleSettings struct {
	maxsize int
}

// Set max packet size and returns old size limitation.
// Set 0 means unlimit.
func (s *SimpleSettings) MaxPacketSize(maxsize int) (old int) {
	old = s.maxsize
	s.maxsize = maxsize
	return
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
