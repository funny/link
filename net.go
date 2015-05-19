package link

import (
	"bufio"
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
	c    net.Conn
	r    *bufio.Reader
	w    *bufio.Writer
	rb   [10]byte
	wb   [10]byte
	rerr error
	werr error
}

func NewConn(c net.Conn, config ConnConfig) *Conn {
	return &Conn{
		c: c,
		r: bufio.NewReaderSize(c, config.ReadBufferSize),
		w: bufio.NewWriterSize(c, config.WriteBufferSize),
	}
}

func (conn *Conn) Close() (err error) {
	if conn.werr == nil && conn.rerr == nil {
		conn.c.SetDeadline(time.Now().Add(time.Second * 3))
		conn.Flush()
	}
	return conn.c.Close()
}

func (conn *Conn) Reset(c net.Conn) {
	conn.c = c
	conn.r.Reset(c)
	conn.w.Reset(c)
	conn.rerr = nil
	conn.werr = nil
}

func (conn *Conn) LocalAddr() net.Addr   { return conn.c.LocalAddr() }
func (conn *Conn) RemoteAddr() net.Addr  { return conn.c.RemoteAddr() }
func (conn *Conn) Reader() *bufio.Reader { return conn.r }
func (conn *Conn) Writer() *bufio.Writer { return conn.w }
func (conn *Conn) RError() error         { return conn.rerr }
func (conn *Conn) WError() error         { return conn.werr }
