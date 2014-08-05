package packnet

import (
	"net"
	"sync"
	"sync/atomic"
)

// Default send chan buffer size for sessions.
var DefaultSendChanSize uint = 1024

// Server.
type Server struct {
	// About network
	listener net.Listener
	protocol PacketProtocol
	writer   PacketWriter

	// About sessions
	sendChanSize uint
	maxSessionId uint64
	sessions     map[uint64]*Session
	sessionMutex sync.Mutex

	// About server start and stop
	stopChan chan int
	stopFlag int32
	stopWait *sync.WaitGroup

	// Server events.
	serverStopCallback   func(*Server)
	sessionStartCallback func(*Session)
	sessionCloseCallback func(*Session)

	// Put your server state here.
	State interface{}
}

// Create a server.
func NewServer(listener net.Listener, protocol PacketProtocol) *Server {
	return &Server{
		listener:     listener,
		protocol:     protocol,
		writer:       protocol.NewWriter(),
		sendChanSize: DefaultSendChanSize,
		maxSessionId: 0,
		sessions:     make(map[uint64]*Session),
		stopChan:     make(chan int),
		stopWait:     new(sync.WaitGroup),
		stopFlag:     -1,
	}
}

// Get listener address.
func (server *Server) Listener() net.Listener {
	return server.listener
}

// Set session send channel buffer size setting.
// New setting will effect on new sessions.
func (server *Server) SetSendChanSize(size uint) {
	server.sendChanSize = size
}

// Get current session send chan buffer size setting.
func (server *Server) GetSendChanSize() uint {
	return server.sendChanSize
}

// Set server stop callback. The callback will invoked when server stop and all session closed.
func (server *Server) OnServerStop(callback func(*Server)) {
	server.serverStopCallback = callback
}

// Set session start callback. The callback  will invoked when a new session start.
func (server *Server) OnSessionStart(callback func(*Session)) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()
	server.sessionStartCallback = callback
}

func (server *Server) getSessionStartCallback() func(*Session) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()
	return server.sessionStartCallback
}

// Set session close callback. The callback  will invoked when a session closed.
func (server *Server) OnSessionClose(callback func(*Session)) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()
	server.sessionCloseCallback = callback
}

// Start server.
func (server *Server) Start() {
	if atomic.CompareAndSwapInt32(&server.stopFlag, -1, 0) {
		go server.acceptLoop()
	} else {
		panic(ServerDuplicateStartError)
	}
}

// Stop server.
func (server *Server) Stop() {
	if atomic.CompareAndSwapInt32(&server.stopFlag, 0, 1) {
		// wait for accept loop exit
		server.listener.Close()
		<-server.stopChan

		// wait for all session exit
		server.closeSessions()
		server.stopWait.Wait()

		if server.serverStopCallback != nil {
			server.serverStopCallback(server)
		}
	}
}

// Loop and accept connections until get an error.
func (server *Server) acceptLoop() {
	defer func() {
		close(server.stopChan)
		server.Stop()
	}()

	for {
		conn, err := server.listener.Accept()
		if err != nil {
			break
		}
		go server.startSession(conn)
	}
}

// Start a session to present the connection.
func (server *Server) startSession(conn net.Conn) {
	session := NewSession(
		atomic.AddUint64(&server.maxSessionId, 1),
		conn,
		server.protocol.NewWriter(),
		server.protocol.NewReader(),
		server.sendChanSize,
	)

	// init the session state
	startCallback := server.getSessionStartCallback()
	if startCallback != nil {
		startCallback(session)
	}

	// session maybe closed in start callback
	if !session.IsClosed() {
		server.putSession(session)
		session.Start(server.closeSession)
	}
}

// Close  and remove a session from server.
func (server *Server) closeSession(session *Session) {
	closeCallback := server.delSession(session)
	if closeCallback != nil {
		closeCallback(session)
	}
}

// Put a session into session list
func (server *Server) putSession(session *Session) {
	if atomic.LoadInt32(&server.stopFlag) == 0 {
		server.sessionMutex.Lock()
		defer server.sessionMutex.Unlock()

		server.sessions[session.id] = session
	}

	server.stopWait.Add(1)
}

// Delete a session from session list
func (server *Server) delSession(session *Session) func(*Session) {
	if atomic.LoadInt32(&server.stopFlag) == 0 {
		server.sessionMutex.Lock()
		defer server.sessionMutex.Unlock()

		delete(server.sessions, session.id)
	}

	server.stopWait.Done()

	return server.sessionCloseCallback
}

// Close all sessions.
func (server *Server) closeSessions() {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()

	for _, session := range server.sessions {
		session.Close()
	}
}
