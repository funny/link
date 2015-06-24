package link

import (
	"github.com/funny/binary"
	"net"
)

var (
	_ PacketServerProtocol = &PacketProtocol{}
	_ PacketClientProtocol = &PacketProtocol{}
	_ IPacketListener      = &PacketListener{}
	_ IPacketConn          = &PacketConn{}
)

var (
	Line     = binary.SplitByLine
	Zero     = binary.SplitByZero
	Uvarint  = binary.SplitByUvarint
	Uint8    = binary.SplitByUint8
	Uint16BE = binary.SplitByUint16BE
	Uint16LE = binary.SplitByUint16LE
	Uint24BE = binary.SplitByUint24BE
	Uint24LE = binary.SplitByUint24LE
	Uint32BE = binary.SplitByUint32BE
	Uint32LE = binary.SplitByUint32LE
	Uint40BE = binary.SplitByUint40BE
	Uint40LE = binary.SplitByUint40LE
	Uint48BE = binary.SplitByUint48BE
	Uint48LE = binary.SplitByUint48LE
	Uint56BE = binary.SplitByUint56BE
	Uint56LE = binary.SplitByUint56LE
	Uint64BE = binary.SplitByUint64BE
	Uint64LE = binary.SplitByUint64LE
)

type PacketProtocol struct {
	Spliter          binary.Spliter
	ReadBufferSize   int
	WriteBufferSize  int
	ClientHandshaker func(*PacketConn) error
	ServerHandshaker func(*PacketConn) error
}

func Packet(spliter binary.Spliter) *PacketProtocol {
	return &PacketProtocol{spliter, 8192, 8192, nil, nil}
}

func (protocol *PacketProtocol) NewPacketListener(listener net.Listener) (IPacketListener, error) {
	return NewPacketListener(listener, protocol), nil
}

func (protocol *PacketProtocol) NewPacketClientConn(conn net.Conn) (IPacketConn, error) {
	lconn := NewPacketConn(conn, protocol)
	if protocol.ClientHandshaker != nil {
		if err := protocol.ClientHandshaker(lconn); err != nil {
			lconn.Close()
			return nil, err
		}
	}
	return lconn, nil
}

func (protocol *PacketProtocol) NewListener(listener net.Listener) (Listener, error) {
	return protocol.NewPacketListener(listener)
}

func (protocol *PacketProtocol) NewClientConn(conn net.Conn) (Conn, error) {
	return protocol.NewPacketClientConn(conn)
}

type PacketListener struct {
	listener net.Listener
	protocol *PacketProtocol
}

func NewPacketListener(listener net.Listener, protocol *PacketProtocol) *PacketListener {
	return &PacketListener{listener, protocol}
}

func (l *PacketListener) AcceptPacket() (IPacketConn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return NewPacketConn(conn, l.protocol), nil
}

func (l *PacketListener) Handshake(conn Conn) error {
	if l.protocol.ServerHandshaker != nil {
		return l.protocol.ServerHandshaker(conn.(*PacketConn))
	}
	return nil
}

func (l *PacketListener) Accept() (Conn, error) { return l.AcceptPacket() }
func (l *PacketListener) Addr() net.Addr        { return l.listener.Addr() }
func (l *PacketListener) Close() error          { return l.listener.Close() }

type PacketConn struct {
	Conn    net.Conn
	Spliter binary.Spliter
	Reader  *binary.Reader
	Writer  *binary.Writer
}

func NewPacketConn(conn net.Conn, protocol *PacketProtocol) *PacketConn {
	return &PacketConn{
		Conn:    conn,
		Spliter: protocol.Spliter,
		Reader:  binary.NewBufioReader(conn, protocol.ReadBufferSize),
		Writer:  binary.NewBufioWriter(conn, protocol.WriteBufferSize),
	}
}

func (conn *PacketConn) ReadPacket() ([]byte, error) {
	b := conn.Reader.ReadPacket(conn.Spliter)
	return b, conn.Reader.Error()
}

func (conn *PacketConn) WritePacket(msg []byte) error {
	conn.Writer.WritePacket(msg, conn.Spliter)
	return conn.Writer.Flush()
}

func (conn *PacketConn) Close() error         { return conn.Conn.Close() }
func (conn *PacketConn) LocalAddr() net.Addr  { return conn.Conn.LocalAddr() }
func (conn *PacketConn) RemoteAddr() net.Addr { return conn.Conn.RemoteAddr() }
