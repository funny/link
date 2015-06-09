package stream

import (
	"net"
	"time"

	"github.com/funny/binary"
	"github.com/funny/link"
)

type Protocol struct {
	ReadBufferSize   int
	WriteBufferSize  int
	SendChanSize     int
	ClientHandshaker func(link.Conn) error
	ServerHandshaker func(link.Conn) error
}

func New(readBufferSize, writeBufferSize, sendChanSize int) *Protocol {
	return &Protocol{readBufferSize, writeBufferSize, sendChanSize, nil, nil}
}

func (protocol *Protocol) NewListener(listener net.Listener) link.Listener {
	return NewListener(listener, protocol)
}

func (protocol *Protocol) NewClientConn(conn net.Conn) (link.Conn, error) {
	lconn := NewConn(conn, protocol)
	if protocol.ClientHandshaker != nil {
		if err := protocol.ClientHandshaker(lconn); err != nil {
			lconn.Close()
			return nil, err
		}
	}
	return lconn, nil
}

type Listener struct {
	listener net.Listener
	protocol *Protocol
}

func NewListener(listener net.Listener, protocol *Protocol) link.Listener {
	return &Listener{listener, protocol}
}

func (l *Listener) Accept() (link.Conn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return NewConn(conn, l.protocol), nil
}

func (l *Listener) Handshake(conn link.Conn) error {
	if l.protocol.ServerHandshaker != nil {
		return l.protocol.ServerHandshaker(conn)
	}
	return nil
}

func (l *Listener) Protocol() link.ServerProtocol { return l.protocol }
func (l *Listener) Addr() net.Addr                { return l.listener.Addr() }
func (l *Listener) Close() error                  { return l.listener.Close() }

type Conn struct {
	c      net.Conn
	config link.SessionConfig
	Reader *binary.Reader
	Writer *binary.Writer
}

func NewConn(conn net.Conn, config *Protocol) *Conn {
	return &Conn{
		c:      conn,
		config: link.SessionConfig{config.SendChanSize},
		Reader: binary.NewBufioReader(conn, config.ReadBufferSize),
		Writer: binary.NewBufioWriter(conn, config.WriteBufferSize),
	}
}

func (conn *Conn) Close() error {
	if conn.c.SetDeadline(time.Now().Add(time.Second*3)) == nil {
		conn.Writer.Flush()
	}
	return conn.c.Close()
}

func (conn *Conn) Receive(msg interface{}) error {
	if err := msg.(InMessage).Unmarshal(conn.Reader); err != nil {
		return err
	}
	return conn.Reader.Error()
}

func (conn *Conn) Send(msg interface{}) error {
	if err := msg.(OutMessage).Marshal(conn.Writer); err != nil {
		return err
	}
	return conn.Writer.Flush()
}

func (conn *Conn) Config() link.SessionConfig { return conn.config }
func (conn *Conn) LocalAddr() net.Addr        { return conn.c.LocalAddr() }
func (conn *Conn) RemoteAddr() net.Addr       { return conn.c.RemoteAddr() }
