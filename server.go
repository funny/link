package packnet

import (
	"net"
	"sync"
	"sync/atomic"
)

type SessionList interface {
	Fetch(func(*Session))
}

// Server.
type Server struct {
	// About network
	listener net.Listener
	protocol PacketProtocol
	writer   PacketWriter

	// About sessions
	sendChanBuff uint
	maxSessionId uint64
	sessions     map[uint64]*Session
	sessionMutex sync.Mutex

	// About server start and stop
	stopChan chan int
	stopFlag int32
	stopWait *sync.WaitGroup
	started  bool

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
		sendChanBuff: 1024,
		maxSessionId: 0,
		sessions:     make(map[uint64]*Session),
		stopChan:     make(chan int),
		stopWait:     new(sync.WaitGroup),
	}
}

// Get listener address.
func (server *Server) Listener() net.Listener {
	return server.listener
}

// Set session send channel buffer size setting.
// New setting will effect on new sessions.
func (server *Server) SetSendChanBuff(size uint) {
	server.sendChanBuff = size
}

// Get current session send chan buffer size setting.
func (server *Server) GetSendChanBuff() uint {
	return server.sendChanBuff
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

// Set session close callback. The callback  will invoked when a session closed.
func (server *Server) OnSessionClose(callback func(*Session)) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()
	server.sessionCloseCallback = callback
}

// Start server.
func (server *Server) Start() {
	if !server.started {
		server.started = true
		go server.acceptLoop()
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

// Broadcast to sessions. The message only encoded one time
// so performance better than send message one by one.
func (server *Server) Broadcast(sessions SessionList, message Message) {
	size := message.RecommendPacketSize()

	packet := server.writer.BeginPacket(size, nil)
	packet = message.AppendToPacket(packet)
	packet = server.writer.EndPacket(packet)

	sessions.Fetch(func(session *Session) {
		session.SendPacket(packet)
	})
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
	atomic.AddUint64(&server.maxSessionId, 1)

	session := NewSession(
		server.maxSessionId,
		conn,
		server.protocol.NewWriter(),
		server.protocol.NewReader(),
		server.GetSendChanBuff(),
	)

	session.OnClose(server.closeSession)

	startCallback := server.putSession(session)
	if startCallback != nil {
		startCallback(session)
	}

	session.Start()
}

// Close  and remove a session from server.
func (server *Server) closeSession(session *Session) {
	closeCallback := server.delSession(session)
	if closeCallback != nil {
		closeCallback(session)
	}
}

// Put a session into session list
func (server *Server) putSession(session *Session) func(*Session) {
	if atomic.LoadInt32(&server.stopFlag) == 0 {
		server.sessionMutex.Lock()
		defer server.sessionMutex.Unlock()

		server.sessions[session.id] = session
	}

	server.stopWait.Add(1)

	return server.sessionStartCallback
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
