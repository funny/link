package packet

import (
	"net"

	"github.com/funny/binary"
	"github.com/funny/link"
	"github.com/funny/link/stream"
)

type Protocol struct {
	*stream.Protocol
	Spliter          binary.Spliter
	ClientHandshaker func(link.Conn) error
	ServerHandshaker func(link.Conn) error
}

func New(spliter binary.Spliter, readBufferSize, writeBufferSize, sendChanSize int) *Protocol {
	return &Protocol{stream.New(readBufferSize, writeBufferSize, sendChanSize), spliter, nil, nil}
}

func (protocol *Protocol) NewListener(listener net.Listener) link.Listener {
	return &Listener{listener, protocol}
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
	Spliter binary.Spliter
	*stream.Conn
}

func NewConn(conn net.Conn, config *Protocol) *Conn {
	return &Conn{
		Spliter: config.Spliter,
		Conn:    stream.NewConn(conn, config.Protocol),
	}
}

func (conn *Conn) Receive(msg interface{}) error {
	if spliter, ok := conn.Spliter.(binary.HeadSpliter); ok {
		if fast, ok := msg.(FastInMessage); ok {
			r := spliter.Limit(conn.Reader)
			if conn.Reader.Error() != nil {
				return conn.Reader.Error()
			}
			return fast.Unmarshal(r)
		}
	}
	return msg.(InMessage).Unmarshal(conn.Reader.ReadPacket(conn.Spliter))
}

func (conn *Conn) Send(msg interface{}) error {
	if spliter, ok := conn.Spliter.(binary.HeadSpliter); ok {
		if fast, ok := msg.(FastOutMessage); ok {
			spliter.WriteHead(conn.Writer, fast.MarshalSize())
			if err := fast.Marshal(conn.Writer); err != nil {
				return err
			}
			return conn.Writer.Flush()
		}
	}
	b, err := msg.(OutMessage).Marshal()
	if err != nil {
		return err
	}
	conn.Writer.WritePacket(b, conn.Spliter)
	return conn.Writer.Flush()
}
