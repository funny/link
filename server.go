package link

import (
	"net"
	"sync"
	"time"
)

const sessionMapNum = 100

type sessionMap struct {
	sync.RWMutex
	sessions map[uint64]*Session
}

type Server struct {
	listener  net.Listener
	codecType CodecType

	// About sessions
	maxSessionId uint64
	sessionMaps  [sessionMapNum]sessionMap

	// About server start and stop
	stopOnce sync.Once
	stopWait sync.WaitGroup

	// Server state
	State interface{}
}

func NewServer(listener net.Listener, codecType CodecType) *Server {
	server := &Server{
		listener:  listener,
		codecType: codecType,
	}
	for i := 0; i < sessionMapNum; i++ {
		server.sessionMaps[i].sessions = make(map[uint64]*Session)
	}
	return server
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
	smap := &server.sessionMaps[sessionId%sessionMapNum]
	smap.RLock()
	defer smap.RUnlock()

	session, _ := smap.sessions[sessionId]
	return session
}

func (server *Server) newSession(conn net.Conn) *Session {
	session := NewSession(conn, server.codecType)
	server.putSession(session)
	return session
}

func (server *Server) putSession(session *Session) {
	smap := &server.sessionMaps[session.id%sessionMapNum]
	smap.Lock()
	defer smap.Unlock()

	session.AddCloseCallback(server, func() { server.delSession(session) })
	smap.sessions[session.id] = session
	server.stopWait.Add(1)
}

func (server *Server) delSession(session *Session) {
	smap := &server.sessionMaps[session.id%sessionMapNum]
	smap.Lock()
	defer smap.Unlock()

	session.RemoveCloseCallback(server)
	delete(smap.sessions, session.id)
	server.stopWait.Done()
}

func (server *Server) copySessions(n int) []*Session {
	smap := &server.sessionMaps[n]
	smap.Lock()
	defer smap.Unlock()

	sessions := make([]*Session, 0, len(smap.sessions))
	for _, session := range smap.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (server *Server) closeSessions() {
	// copy session to avoid deadlock
	for i := 0; i < sessionMapNum; i++ {
		sessions := server.copySessions(i)
		for _, session := range sessions {
			session.Close()
		}
	}
}
