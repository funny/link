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

var (
	DefaultSendChanSize   = 1024                     // Default session send chan buffer size.
	DefaultConnBufferSize = 1024                     // Default session read buffer size.
	DefaultProtocol       = PacketN(4, LittleEndian) // Default protocol for utility APIs.
)

// The easy way to setup a server.
func Listen(network, address string) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, DefaultProtocol), nil
}

// Server.
type Server struct {
	// About network
	listener    net.Listener
	protocol    Protocol
	broadcaster *Broadcaster

	// About sessions
	maxSessionId uint64
	sessions     map[uint64]*Session
	sessionMutex sync.Mutex

	// About server start and stop
	stopFlag int32
	stopWait sync.WaitGroup

	SendChanSize   int         // Session send chan buffer size.
	ReadBufferSize int         // Session read buffer size.
	State          interface{} // server state.
}

// Create a server.
func NewServer(listener net.Listener, protocol Protocol) *Server {
	server := &Server{
		listener:       listener,
		protocol:       protocol,
		sessions:       make(map[uint64]*Session),
		SendChanSize:   DefaultSendChanSize,
		ReadBufferSize: DefaultConnBufferSize,
	}
	protocolState, _ := protocol.New(server, SERVER_SIDE)
	server.broadcaster = NewBroadcaster(protocolState, server.fetchSession)
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

// Broadcast to channel. The message only encoded once
// so the performance is better than send message one by one.
func (server *Server) BroadcastFunc(e func(*OutBuffer) error) ([]BroadcastWork, error) {
	return server.broadcaster.BroadcastFunc(e)
}

// Broadcast to channel. The message only encoded once
// so the performance is better than send message one by one.
func (server *Server) Broadcast(message Message) ([]BroadcastWork, error) {
	return server.broadcaster.Broadcast(message)
}

// Accept incoming connection once.
func (server *Server) Accept() (*Session, error) {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			return nil, err
		}
		session := server.newSession(
			atomic.AddUint64(&server.maxSessionId, 1),
			conn,
		)
		if session != nil {
			return session, nil
		}
	}
}

// Loop and accept incoming connections. The callback will called asynchronously when each session start.
func (server *Server) Serve(handler func(*Session)) error {
	for {
		session, err := server.Accept()
		if err != nil {
			server.Stop()
			return err
		}
		go handler(session)
	}
	return nil
}

// Stop server.
func (server *Server) Stop() {
	if atomic.CompareAndSwapInt32(&server.stopFlag, 0, 1) {
		server.listener.Close()
		server.closeSessions()
		server.stopWait.Wait()
	}
}

func (server *Server) newSession(id uint64, conn net.Conn) *Session {
	session, _ := NewSession(id, conn, server.protocol, SERVER_SIDE, server.SendChanSize, server.ReadBufferSize)
	if session == nil {
		return nil
	}
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
func (server *Server) fetchSession(callback func(*Session)) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()

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
