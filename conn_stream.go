package link

import (
	"bufio"
	"net"
)

var (
	_ StreamServerProtocol = &StreamProtocol{}
	_ StreamClientProtocol = &StreamProtocol{}
	_ IStreamListener      = &StreamListener{}
	_ IStreamConn          = &StreamConn{}
)

type StreamProtocol struct {
	ReadBufferSize   int
	WriteBufferSize  int
	ClientHandshaker func(*StreamConn) error
	ServerHandshaker func(*StreamConn) error
}

func Stream() *StreamProtocol {
	return &StreamProtocol{8192, 8192, nil, nil}
}

func (protocol *StreamProtocol) NewStreamListener(listener net.Listener) (IStreamListener, error) {
	return NewStreamListener(listener, protocol), nil
}

func (protocol *StreamProtocol) NewStreamClientConn(conn net.Conn) (IStreamConn, error) {
	lconn := NewStreamConn(conn, protocol)
	if protocol.ClientHandshaker != nil {
		if err := protocol.ClientHandshaker(lconn); err != nil {
			lconn.Close()
			return nil, err
		}
	}
	return lconn, nil
}

func (protocol *StreamProtocol) NewListener(listener net.Listener) (Listener, error) {
	return protocol.NewStreamListener(listener)
}

func (protocol *StreamProtocol) NewClientConn(conn net.Conn) (Conn, error) {
	return protocol.NewStreamClientConn(conn)
}

type StreamListener struct {
	listener net.Listener
	protocol *StreamProtocol
}

func NewStreamListener(listener net.Listener, protocol *StreamProtocol) *StreamListener {
	return &StreamListener{listener, protocol}
}

func (l *StreamListener) AcceptStream() (IStreamConn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return NewStreamConn(conn, l.protocol), nil
}

func (l *StreamListener) Handshake(conn Conn) error {
	if l.protocol.ServerHandshaker != nil {
		return l.protocol.ServerHandshaker(conn.(*StreamConn))
	}
	return nil
}

func (l *StreamListener) Accept() (Conn, error) { return l.AcceptStream() }
func (l *StreamListener) Addr() net.Addr        { return l.listener.Addr() }
func (l *StreamListener) Close() error          { return l.listener.Close() }

type StreamConn struct {
	Conn   net.Conn
	Reader *bufio.Reader
	Writer *bufio.Writer
}

func NewStreamConn(conn net.Conn, protocol *StreamProtocol) *StreamConn {
	return &StreamConn{
		Conn:   conn,
		Reader: bufio.NewReaderSize(conn, protocol.ReadBufferSize),
		Writer: bufio.NewWriterSize(conn, protocol.WriteBufferSize),
	}
}

func (conn *StreamConn) UpStream() *bufio.Reader   { return conn.Reader }
func (conn *StreamConn) DownStream() *bufio.Writer { return conn.Writer }
func (conn *StreamConn) Close() error              { return conn.Conn.Close() }
func (conn *StreamConn) LocalAddr() net.Addr       { return conn.Conn.LocalAddr() }
func (conn *StreamConn) RemoteAddr() net.Addr      { return conn.Conn.RemoteAddr() }
