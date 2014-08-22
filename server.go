package link

import (
	"net"
	"sync"
	"sync/atomic"
)

// Default send chan buffer size for sessions.
var DefaultSendChanSize uint = 1024

// Default read buffer size for session.
var DefaultReadBufferSize int = 1024

// Server.
type Server struct {
	// About network
	listener net.Listener
	protocol PacketProtocol

	// About sessions
	sendChanSize   uint
	readBufferSize int
	maxSessionId   uint64
	sessions       map[uint64]*Session
	sessionMutex   sync.Mutex

	// About server start and stop
	stopFlag   int32
	stopWait   *sync.WaitGroup
	stopReason interface{}

	// Put your server state here.
	State interface{}
}

// Create a server.
func NewServer(listener net.Listener, protocol PacketProtocol) *Server {
	return &Server{
		listener:       listener,
		protocol:       protocol,
		sendChanSize:   DefaultSendChanSize,
		readBufferSize: DefaultReadBufferSize,
		maxSessionId:   0,
		sessions:       make(map[uint64]*Session),
		stopWait:       new(sync.WaitGroup),
	}
}

// Get listener address.
func (server *Server) Listener() net.Listener {
	return server.listener
}

// Set session send channel buffer size.
// New setting will effect on new sessions.
func (server *Server) SetSendChanSize(size uint) {
	server.sendChanSize = size
}

// Get current session send chan buffer size setting.
func (server *Server) GetSendChanSize() uint {
	return server.sendChanSize
}

// Get current session read buffer size setting.
func (server *Server) SetReadBufferSize(size int) {
	server.readBufferSize = size
}

// Set session read buffer size.
// New setting will effect on new sessions.
func (server *Server) GetReadBufferSize() int {
	return server.readBufferSize
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
func (server *Server) AcceptLoop(handler func(*Session)) {
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

// Put a session into session list.
func (server *Server) putSession(session *Session) {
	server.sessionMutex.Lock()
	server.sessions[session.id] = session
	session.server.stopWait.Add(1)
	server.sessionMutex.Unlock()
}

// Delete a session from session list.
func (server *Server) delSession(session *Session) {
	server.sessionMutex.Lock()
	delete(server.sessions, session.id)
	session.server.stopWait.Done()
	server.sessionMutex.Unlock()
}

// Close all sessions.
func (server *Server) closeSessions() {
	server.sessionMutex.Lock()
	sessions := make([]*Session, 0, len(server.sessions))
	for _, session := range server.sessions {
		sessions = append(sessions, session)
	}
	server.sessionMutex.Unlock()

	for _, session := range sessions {
		session.Close(nil)
	}
}
