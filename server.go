package link

import (
	"io"
	"net"
	"strings"
	"time"
)

type Server struct {
	manager      *Manager
	listener     net.Listener
	protocol     Protocol
	sendChanSize int
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

func (server *Server) Accept() (*Session, error) {
	var tempDelay time.Duration
	for {
		conn, err := server.listener.Accept()
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
		return server.manager.NewSession(
			server.protocol.NewCodec(conn),
			server.sendChanSize,
		), nil
	}
}

func (server *Server) Stop() {
	server.listener.Close()
	server.manager.Dispose()
}
