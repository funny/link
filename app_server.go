package link

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	ConnConfig
	SessionConfig
}

type Server struct {
	listener *Listener

	// About sessions
	Config
	maxSessionId uint64
	sessions     map[uint64]*Session
	sessionMutex sync.Mutex

	// About server start and stop
	stopFlag int32
	stopChan chan int
	stopWait sync.WaitGroup

	State interface{} // server state.
}

func NewServer(listener *Listener, config Config) *Server {
	server := &Server{
		listener: listener,
		sessions: make(map[uint64]*Session),
		stopChan: make(chan int),
		Config:   config,
	}
	go server.checkAlive()
	return server
}

func (server *Server) Listener() net.Listener {
	return server.listener.l
}

func (server *Server) Broadcast(msg OutMessage) ([]BroadcastWork, error) {
	return Broadcast(msg, server.sessionFetcher)
}

func (server *Server) Accept() (*Session, error) {
	conn, err := server.listener.Accept()
	if err != nil {
		return nil, err
	}
	session := server.newSession(
		atomic.AddUint64(&server.maxSessionId, 1),
		conn,
	)
	return session, nil
}

func (server *Server) Serve(handler func(*Session)) error {
	for {
		session, err := server.Accept()
		if err != nil {
			if server.Stop() {
				return err
			}
			return nil
		}
		go handler(session)
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

func (server *Server) sessionFetcher(callback func(*Session)) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()

	for _, session := range server.sessions {
		callback(session)
	}
}

func (server *Server) newSession(id uint64, conn *Conn) *Session {
	session := NewSession(id, conn, server.SessionConfig)
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

func (server *Server) checkAlive() {
	tick := time.NewTicker(time.Second)
	for {
		select {
		case <-tick.C:
			now := time.Now()
			server.sessionFetcher(func(session *Session) {
				if session.IsTimeout(now) {
					go session.Close()
				}
			})
		case <-server.stopChan:
			tick.Stop()
			return
		}
	}
}
