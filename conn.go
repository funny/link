package link

import (
	"github.com/funny/binary"
	"net"
	"time"
)

type ConnConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	AliveTimeout    time.Duration
}

type Listener struct {
	l net.Listener
	ConnConfig
}

func NewListener(l net.Listener, config ConnConfig) *Listener {
	return &Listener{
		l:          l,
		ConnConfig: config,
	}
}

func (listener *Listener) Accept() (*Conn, error) {
	c, err := listener.l.Accept()
	if err != nil {
		return nil, err
	}
	return NewConn(c, listener.ConnConfig), nil
}

func (listener *Listener) Close() error {
	return listener.l.Close()
}

type Conn struct {
	c net.Conn
	r *binary.Reader
	w *binary.Writer
}

func NewConn(c net.Conn, config ConnConfig) *Conn {
	return &Conn{
		c: c,
		r: binary.NewBufioReader(c, config.ReadBufferSize),
		w: binary.NewBufioWriter(c, config.WriteBufferSize),
	}
}

func (conn *Conn) Close() (err error) {
	if conn.w.Error() == nil && conn.r.Error() == nil {
		conn.c.SetDeadline(time.Now().Add(time.Second * 3))
		conn.w.Flush()
	}
	return conn.c.Close()
}

func (conn *Conn) LocalAddr() net.Addr    { return conn.c.LocalAddr() }
func (conn *Conn) RemoteAddr() net.Addr   { return conn.c.RemoteAddr() }
func (conn *Conn) Reader() *binary.Reader { return conn.r }
func (conn *Conn) Writer() *binary.Writer { return conn.w }
