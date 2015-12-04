package link

import (
	"net"
	"sync"
	"time"
)

type Server struct {
	listener  net.Listener
	codecType CodecType

	// About sessions
	maxSessionId uint64
	sessions     map[uint64]*Session
	sessionMutex sync.RWMutex

	// About server start and stop
	stopOnce sync.Once
	stopWait sync.WaitGroup

	// Server state
	State interface{}
}

func NewServer(listener net.Listener, codecType CodecType) *Server {
	return &Server{
		listener:  listener,
		codecType: codecType,
		sessions:  make(map[uint64]*Session),
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
			return nil, err
		}
		tempDelay = 0
		return server.newSession(conn), nil
	}
}

func (server *Server) Stop() {
	server.stopOnce.Do(func() {
		server.listener.Close()
		server.closeSessions()
		server.stopWait.Wait()
	})
}

func (server *Server) GetSession(sessionId uint64) *Session {
	server.sessionMutex.RLock()
	defer server.sessionMutex.RUnlock()
	session, _ := server.sessions[sessionId]
	return session
}

func (server *Server) newSession(conn net.Conn) *Session {
	session := NewSession(conn, server.codecType)
	server.putSession(session)
	return session
}

func (server *Server) putSession(session *Session) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()

	session.AddCloseCallback(server, func() { server.delSession(session) })
	server.sessions[session.id] = session
	server.stopWait.Add(1)
}

func (server *Server) delSession(session *Session) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()

	session.RemoveCloseCallback(server)
	delete(server.sessions, session.id)
	server.stopWait.Done()
}

func (server *Server) copySessions() []*Session {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()

	sessions := make([]*Session, 0, len(server.sessions))
	for _, session := range server.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (server *Server) closeSessions() {
	// copy session to avoid deadlock
	sessions := server.copySessions()
	for _, session := range sessions {
		session.Close()
	}
}
