package link

import (
	"errors"
	"github.com/funny/sync"
	"net"
	"sync/atomic"
)

// Errors
var (
	SendToClosedError     = errors.New("Send to closed session")
	PacketTooLargeError   = errors.New("Packet too large")
	AsyncSendTimeoutError = errors.New("Async send timeout")
)

// The easy way to setup a server.
func Listen(network, address string, protocol Protocol, pool *MemPool) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, protocol, pool, DefaultConfig), nil
}

// Server.
type Server struct {
	Config

	// About network
	listener    net.Listener
	protocol    Protocol
	pool        *MemPool
	broadcaster *Broadcaster

	// About sessions
	maxSessionId uint64
	sessions     map[uint64]*Session
	sessionMutex sync.Mutex

	// About server start and stop
	stopFlag int32
	stopWait sync.WaitGroup

	State interface{} // server state.
}

// Create a server.
func NewServer(listener net.Listener, protocol Protocol, pool *MemPool, config Config) *Server {
	server := &Server{
		listener: listener,
		protocol: protocol,
		pool:     pool,
		sessions: make(map[uint64]*Session),
		Config:   config,
	}
	codec := protocol.NewCodec()
	server.broadcaster = NewBroadcaster(codec, pool, server.sessionFetcher)
	return server
}

// Get listener address.
func (server *Server) Listener() net.Listener {
	return server.listener
}

// Get protocol.
func (server *Server) Protocol() Protocol {
	return server.protocol
}

// Broadcast.
func (server *Server) Broadcast(msg Message) ([]BroadcastWork, error) {
	return server.broadcaster.Broadcast(msg)
}

// Accept incoming connection once.
func (server *Server) accept() (*Session, error) {
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

// Loop and accept incoming connections. The callback will called asynchronously when each session start.
func (server *Server) Serve(handler func(*Session)) error {
	for {
		session, err := server.accept()
		if err != nil {
			if server.Stop() {
				return err
			}
			return nil
		}
		go func() {
			if err := session.Handshake(); err != nil {
				session.Close()
				return
			}
			handler(session)
		}()
	}
	return nil
}

// Stop server.
func (server *Server) Stop() bool {
	if atomic.CompareAndSwapInt32(&server.stopFlag, 0, 1) {
		server.listener.Close()
		server.closeSessions()
		server.stopWait.Wait()
		return true
	}
	return false
}

func (server *Server) newSession(id uint64, conn net.Conn) *Session {
	session := NewSession(id, conn, server.protocol.NewCodec(), server.pool, server.Config)
	server.putSession(session)
	return session
}

// Put a session into session list.
func (server *Server) putSession(session *Session) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()

	session.AddCloseCallback(server, func() {
		server.delSession(session)
	})
	server.sessions[session.id] = session
	server.stopWait.Add(1)
}

// Delete a session from session list.
func (server *Server) delSession(session *Session) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()

	session.RemoveCloseCallback(server)
	delete(server.sessions, session.id)
	server.stopWait.Done()
}

// Copy sessions for close.
func (server *Server) copySessions() []*Session {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()

	sessions := make([]*Session, 0, len(server.sessions))
	for _, session := range server.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// Fetch sessions.
func (server *Server) sessionFetcher(num *int32, callback func(*Session)) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()

	*num = int32(len(server.sessions))

	for _, session := range server.sessions {
		callback(session)
	}
}

// Close all sessions.
func (server *Server) closeSessions() {
	// copy session to avoid deadlock
	sessions := server.copySessions()
	for _, session := range sessions {
		session.Close()
	}
}
