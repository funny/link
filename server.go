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
	serverStopHook   func(*Server)
	sessionStartHook func(*Session)
	sessionCloseHook func(*Session)

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

// Set server stop hook. Hook will invoked when server stop and all session closed.
func (server *Server) SetServerStopHook(hook func(*Server)) {
	server.serverStopHook = hook
}

// Set session start hook. Hook will invoked when a new session start.
func (server *Server) SetSessionStartHook(hook func(*Session)) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()
	server.sessionStartHook = hook
}

// Set session close hook. Hook will invoked when a session closed.
func (server *Server) SetSessionCloseHook(hook func(*Session)) {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()
	server.sessionCloseHook = hook
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

		if server.serverStopHook != nil {
			server.serverStopHook(server)
		}
	}
}

// Broadcast to sessions. The response only encoded one time
// so performance better than send response one by one.
func (server *Server) Broadcast(sessions SessionList, response Response) {
	size := response.RecommendPacketSize()

	packet := server.writer.BeginPacket(size, nil)
	packet = response.AppendToPacket(packet)
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

	session.SetCloseCallback(server.closeSession)

	startHook := server.putSession(session)
	if startHook != nil {
		startHook(session)
	}

	session.Start()
}

// Close  and remove a session from server.
func (server *Server) closeSession(session *Session) {
	closeHook := server.delSession(session)
	if closeHook != nil {
		closeHook(session)
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

	return server.sessionStartHook
}

// Delete a session from session list
func (server *Server) delSession(session *Session) func(*Session) {
	if atomic.LoadInt32(&server.stopFlag) == 0 {
		server.sessionMutex.Lock()
		defer server.sessionMutex.Unlock()

		delete(server.sessions, session.id)
	}

	server.stopWait.Done()

	return server.sessionCloseHook
}

// Close all sessions.
func (server *Server) closeSessions() {
	server.sessionMutex.Lock()
	defer server.sessionMutex.Unlock()

	for _, session := range server.sessions {
		session.Close()
	}
}
