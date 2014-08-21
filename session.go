package link

import (
	"net"
	"sync"
	"sync/atomic"
)

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
	readBUff       []byte
	sendBuff       []byte
	sendLock       sync.Mutex
	messageHandler MessageHandler

	// About session close
	closeChan   chan int
	closeFlag   int32
	closeReason error

	// Put your session state here.
	State interface{}
}

// Create a new session instance.
func NewSession(id uint64, conn net.Conn, protocol PacketProtocol, sendChanSize uint) *Session {
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
	session := NewSession(id, conn, server.protocol, server.sendChanSize)
	session.server = server
	session.server.putSession(session)
	return session
}

// Loop and read message.
func (session *Session) ReadLoop(handler func([]byte)) {
	for {
		msg := session.Read()
		if msg == nil {
			break
		}
		handler(msg)
	}
}

// Read message once.
func (session *Session) Read() []byte {
	if session.IsClosed() {
		return nil
	}
	var err error
	session.readBUff, err = session.reader.ReadPacket(session.conn, session.readBUff)
	if err != nil {
		session.Close(err)
		return nil
	}
	return session.readBUff
}

// Loop and transport responses.
func (session *Session) sendLoop() {
	for {
		select {
		case message := <-session.sendChan:
			session.SyncSend(message)
		case packet := <-session.sendPacketChan:
			session.SyncSendPacket(packet)
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
func (session *Session) CloseReason() error {
	return session.closeReason
}

// Close session and remove it from api server.
func (session *Session) Close(reason error) {
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

	err := session.writer.WritePacket(session.conn, packet)
	if err != nil {
		session.Close(err)
	}

	return err
}

// Sync send a packet. Use in carefully.
// The packet must be properly formatted.
// If you didn't know what it means, please see Channel.Broadcast().
// Use in carefully.
func (session *Session) SyncSendPacket(packet []byte) error {
	session.sendLock.Lock()
	defer session.sendLock.Unlock()

	err := session.writer.WritePacket(session.conn, packet)
	if err != nil {
		session.Close(err)
	}

	return err
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
		session.Close(BlockingError)
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
		session.Close(BlockingError)
		return BlockingError
	}
}
