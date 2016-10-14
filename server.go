package link

import (
	"io"
	"net"
	"strings"
	"time"
)

func Accept(listener net.Listener) (net.Conn, error) {
	var tempDelay time.Duration
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil, io.EOF
			}
			return nil, err
		}
		return conn, nil
	}
}

type Server struct {
	manager      *Manager
	listener     net.Listener
	protocol     Protocol
	sendChanSize int
}

type Handler interface {
	HandleSession(session *Session, ctx Context, err error)
}

var _ Handler = HandlerFunc(nil)

type HandlerFunc func(session *Session, ctx Context, err error)

func (hf HandlerFunc) HandleSession(session *Session, ctx Context, err error) {
	hf(session, ctx, err)
}

func NewServer(l net.Listener, p Protocol, sendChanSize int) *Server {
	return &Server{
		manager:      NewManager(),
		listener:     l,
		protocol:     p,
		sendChanSize: sendChanSize,
	}
}

func (server *Server) Listener() net.Listener {
	return server.listener
}

func (server *Server) Serve(handler Handler) error {
	for {
		conn, err := Accept(server.listener)
		if err != nil {
			return err
		}

		go func() {
			codec, ctx, err := server.protocol.NewCodec(conn)
			if err != nil {
				conn.Close()
				return
			}
			session := server.manager.NewSession(codec, server.sendChanSize)
			handler.HandleSession(session, ctx, nil)
		}()
	}
}

func (server *Server) GetSession(sessionID uint64) *Session {
	return server.manager.GetSession(sessionID)
}

func (server *Server) Stop() {
	server.listener.Close()
	server.manager.Dispose()
}
