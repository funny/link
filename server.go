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
	DefaultProtocol       = PacketN(4, LittleEndian) //Default protocol for utility APIs.
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
	listener       net.Listener
	protocol       Protocol
	broadcastState ProtocolState

	// About sessions
	maxSessionId uint64
	sessions     map[uint64]*Session
	sessionMutex sync.Mutex

	// About server start and stop
	stopFlag   int32
	stopWait   sync.WaitGroup
	stopReason interface{}

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
	server.broadcastState = protocol.New(server)
	return server
}

// Get listener address.
func (server *Server) Listener() net.Listener {
	return server.listener
}

// Check server is stoppped
func (server *Server) IsStopped() bool {
	return atomic.LoadInt32(&server.stopFlag) == 1
}

// Get server stop reason.
func (server *Server) StopReason() interface{} {
	return server.stopReason
}

// Loop and accept incoming connections. The callback will called asynchronously when each session start.
func (server *Server) Handle(handler func(*Session)) {
	for {
		session, err := server.Accept()
		if err != nil {
			server.Stop(err)
			break
		}
		go handler(session)
	}
}

// Accept incoming connection once.
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

// Implement the SessionCloseEventListener interface.
func (server *Server) OnSessionClose(session *Session) {
	server.delSession(session)
}

// Stop server.
func (server *Server) Stop(reason interface{}) {
	if atomic.CompareAndSwapInt32(&server.stopFlag, 0, 1) {
		server.stopReason = reason

		server.listener.Close()

		// close all sessions
		server.closeSessions()
		server.stopWait.Wait()
	}
}

func (server *Server) newSession(id uint64, conn net.Conn) *Session {
	session := NewSession(id, conn, server.protocol, server.SendChanSize, server.ReadBufferSize)
	server.putSession(session)
	return session
}

// Put a session into session list.
func (server *Server) putSession(session *Session) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()

	session.AddCloseEventListener(server)
	server.sessions[session.id] = session
	server.stopWait.Add(1)
}

// Delete a session from session list.
func (server *Server) delSession(session *Session) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()

	session.RemoveCloseEventListener(server)
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

// Close all sessions.
func (server *Server) closeSessions() {
	sessions := server.copySessions()
	for _, session := range sessions {
		session.Close(nil)
	}
}

// Get server protocol.
func (server *Server) Protocol() Protocol {
	return server.protocol
}

// Get broadcast protocol.
// Implement SessionCollection interface.
func (server *Server) BroadcastState() ProtocolState {
	return server.broadcastState
}

// Fetch sessions.
// Implement SessionCollection interface.
func (server *Server) FetchSession(callback func(*Session)) {
	sessions := server.copySessions()
	for _, session := range sessions {
		callback(session)
	}
}
