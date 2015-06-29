package link

import (
	"sync"
	"sync/atomic"
)

type Server struct {
	listener Listener

	// About sessions
	maxSessionId uint64
	sessions     map[uint64]*Session
	sessionMutex sync.Mutex

	// About server start and stop
	stopFlag int32
	stopChan chan int
	stopWait sync.WaitGroup

	// Server state
	State interface{}
}

func NewServer(listener Listener) *Server {
	server := &Server{
		listener: listener,
		sessions: make(map[uint64]*Session),
		stopChan: make(chan int),
	}
	return server
}

func (server *Server) Listener() Listener {
	return server.listener
}

func (server *Server) Loop(handler func(*Session)) error {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			if server.Stop() {
				return err
			}
			return nil
		}
		go func() {
			if server.listener.Handshake(conn) != nil {
				return
			}
			handler(server.newSession(conn))
		}()
	}
	return nil
}

func (server *Server) Stop() bool {
	if atomic.CompareAndSwapInt32(&server.stopFlag, 0, 1) {
		server.listener.Close()
		close(server.stopChan)
		server.closeSessions()
		server.stopWait.Wait()
		return true
	}
	return false
}

func (server *Server) newSession(conn Conn) *Session {
	session := NewSession(conn)
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
