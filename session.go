package link

import (
	"bufio"
	"net"
	"sync"
	"sync/atomic"
)

// Buffered connection.
type BufferConn struct {
	net.Conn
	reader *bufio.Reader
}

func NewBufferConn(conn net.Conn, size int) *BufferConn {
	return &BufferConn{
		conn,
		bufio.NewReaderSize(conn, size),
	}
}

func (conn *BufferConn) Read(d []byte) (int, error) {
	return conn.reader.Read(d)
}

// Session.
type Session struct {
	id     uint64
	server *Server

	// About network
	conn   net.Conn
	writer PacketWriter
	reader PacketReader

	// About send and receive
	sendChan       chan Message
	sendPacketChan chan []byte
	readBuff       []byte
	sendBuff       []byte
	sendLock       sync.Mutex
	messageHandler MessageHandler

	// About session close
	closeChan   chan int
	closeFlag   int32
	closeReason interface{}

	// Put your session state here.
	State interface{}
}

// Create a new session instance.
func NewSession(id uint64, conn net.Conn, protocol PacketProtocol, sendChanSize uint, readBufferSize int) *Session {
	if readBufferSize > 0 {
		conn = NewBufferConn(conn, readBufferSize)
	}

	session := &Session{
		id:             id,
		conn:           conn,
		writer:         protocol.NewWriter(),
		reader:         protocol.NewReader(),
		sendChan:       make(chan Message, sendChanSize),
		sendPacketChan: make(chan []byte, sendChanSize),
		closeChan:      make(chan int),
	}

	go session.sendLoop()

	return session
}

func (server *Server) newSession(id uint64, conn net.Conn) *Session {
	session := NewSession(id, conn, server.protocol, server.sendChanSize, server.readBufferSize)
	session.server = server
	session.server.putSession(session)
	return session
}

// Loop and read message.
func (session *Session) ReadLoop(handler func([]byte)) {
	for {
		msg, err := session.Read()
		if err != nil {
			session.Close(err)
			break
		}
		handler(msg)
	}
}

// Read message once.
func (session *Session) Read() ([]byte, error) {
	var err error
	session.readBuff, err = session.reader.ReadPacket(session.conn, session.readBuff)
	if err != nil {
		return nil, err
	}
	return session.readBuff, nil
}

// Loop and transport responses.
func (session *Session) sendLoop() {
	for {
		select {
		case message := <-session.sendChan:
			if err := session.SyncSend(message); err != nil {
				session.Close(err)
				return
			}
		case packet := <-session.sendPacketChan:
			if err := session.SyncSendPacket(packet); err != nil {
				session.Close(err)
				return
			}
		case <-session.closeChan:
			return
		}
	}
}

// Get session id.
func (session *Session) Id() uint64 {
	return session.id
}

// Get local address.
func (session *Session) Conn() net.Conn {
	return session.conn
}

// Get session owner.
func (session *Session) Server() *Server {
	return session.server
}

// Get reader settings.
func (session *Session) ReaderSettings() Settings {
	return session.reader
}

// Get writer settings.
func (session *Session) WriterSettings() Settings {
	return session.writer
}

// Check session is closed or not.
func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) != 0
}

// Get session close reason.
func (session *Session) CloseReason() interface{} {
	return session.closeReason
}

// Close session and remove it from api server.
func (session *Session) Close(reason interface{}) {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		session.closeReason = reason

		session.conn.Close()

		// exit send loop
		close(session.closeChan)

		// if this is a server side session
		// remove it from sessin list
		if session.server != nil {
			session.server.delSession(session)
		}
	}
}

// Sync send a message. This method will block on IO.
// Use in carefully.
func (session *Session) SyncSend(message Message) error {
	session.sendLock.Lock()
	defer session.sendLock.Unlock()

	size := message.RecommendPacketSize()

	packet := session.writer.BeginPacket(size, session.sendBuff)
	packet = message.AppendToPacket(packet)
	packet = session.writer.EndPacket(packet)

	session.sendBuff = packet

	return session.writer.WritePacket(session.conn, packet)
}

// Sync send a packet. Use in carefully.
// The packet must be properly formatted.
// If you didn't know what it means, please see Channel.Broadcast().
// Use in carefully.
func (session *Session) SyncSendPacket(packet []byte) error {
	session.sendLock.Lock()
	defer session.sendLock.Unlock()

	return session.writer.WritePacket(session.conn, packet)
}

// Async send a message. This method will never block.
// If channel blocked session will be closed.
func (session *Session) Send(message Message) error {
	if session.IsClosed() {
		return SendToClosedError
	}

	select {
	case session.sendChan <- message:
		return nil
	default:
		return BlockingError
	}
}

// Async send a packet. This method will block on IO.
// The packet must be properly formatted.
// If you didn't know what it means, please see Channel.Broadcast().
// Use in carefully.
func (session *Session) SendPacket(packet []byte) error {
	if session.IsClosed() {
		return SendToClosedError
	}

	select {
	case session.sendPacketChan <- packet:
		return nil
	default:
		return BlockingError
	}
}
