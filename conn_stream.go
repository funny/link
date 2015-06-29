package link

import (
	"bufio"
	"net"
)

var (
	_ ServerProtocol = &StreamProtocol{}
	_ ClientProtocol = &StreamProtocol{}
	_ Listener       = &StreamListener{}
	_ Conn           = &StreamConn{}
)

type StreamCodecType interface {
	NewStreamCodec(r *bufio.Reader, w *bufio.Writer) StreamCodec
}

type StreamCodec interface {
	DecodeStream(msg interface{}) error
	EncodeStream(msg interface{}) error
}

type StreamProtocol struct {
	CodecType        StreamCodecType
	ReadBufferSize   int
	WriteBufferSize  int
	ClientHandshaker func(*StreamConn) error
	ServerHandshaker func(*StreamConn) error
}

func Stream(codecType StreamCodecType) *StreamProtocol {
	return &StreamProtocol{codecType, 8192, 8192, nil, nil}
}

func (protocol *StreamProtocol) NewListener(listener net.Listener) (Listener, error) {
	return NewStreamListener(listener, protocol), nil
}

func (protocol *StreamProtocol) NewClientConn(conn net.Conn) (Conn, error) {
	lconn := NewStreamConn(conn, protocol)
	if protocol.ClientHandshaker != nil {
		if err := protocol.ClientHandshaker(lconn); err != nil {
			lconn.Close()
			return nil, err
		}
	}
	return lconn, nil
}

type StreamListener struct {
	listener net.Listener
	protocol *StreamProtocol
}

func NewStreamListener(listener net.Listener, protocol *StreamProtocol) *StreamListener {
	return &StreamListener{listener, protocol}
}

func (l *StreamListener) Accept() (Conn, error) {
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

func (l *StreamListener) Addr() net.Addr { return l.listener.Addr() }
func (l *StreamListener) Close() error   { return l.listener.Close() }

type StreamConn struct {
	Conn   net.Conn
	Codec  StreamCodec
	Reader *bufio.Reader
	Writer *bufio.Writer
}

func NewStreamConn(conn net.Conn, protocol *StreamProtocol) *StreamConn {
	r := bufio.NewReaderSize(conn, protocol.ReadBufferSize)
	w := bufio.NewWriterSize(conn, protocol.WriteBufferSize)
	return &StreamConn{
		Conn:   conn,
		Codec:  protocol.CodecType.NewStreamCodec(r, w),
		Reader: r,
		Writer: w,
	}
}

func (conn *StreamConn) Send(msg interface{}) error {
	return conn.Codec.EncodeStream(msg)
}

func (conn *StreamConn) Receive(msg interface{}) error {
	return conn.Codec.DecodeStream(msg)
}

func (conn *StreamConn) Close() error         { return conn.Conn.Close() }
func (conn *StreamConn) LocalAddr() net.Addr  { return conn.Conn.LocalAddr() }
func (conn *StreamConn) RemoteAddr() net.Addr { return conn.Conn.RemoteAddr() }
