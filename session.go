package link

import (
	"container/list"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Session.
type Session struct {
	id uint64

	// About network
	conn     net.Conn
	protocol PacketProtocol
	writer   PacketWriter
	reader   PacketReader

	// About send and receive
	sendChan       chan Message
	sendPacketChan chan []byte
	sendBuff       []byte
	sendMutex      sync.Mutex
	OnSendFailed   func(*Session, error)

	// About session close
	closeChan               chan int
	closeFlag               int32
	closeReason             interface{}
	closeEventListenerMutex sync.Mutex
	closeEventListeners     *list.List

	// Put your session state here.
	State interface{}
}

// Create a new session instance.
func NewSession(id uint64, conn net.Conn, protocol PacketProtocol, sendChanSize uint, readBufferSize int) *Session {
	if readBufferSize > 0 {
		conn = NewBufferConn(conn, readBufferSize)
	}

	session := &Session{
		id:                  id,
		conn:                conn,
		protocol:            protocol,
		writer:              protocol.NewWriter(),
		reader:              protocol.NewReader(),
		sendChan:            make(chan Message, sendChanSize),
		sendPacketChan:      make(chan []byte, sendChanSize),
		closeChan:           make(chan int),
		closeEventListeners: list.New(),
	}

	go session.sendLoop()

	return session
}

// Loop and transport responses.
func (session *Session) sendLoop() {
	for {
		select {
		case message := <-session.sendChan:
			if err := session.Send(message); err != nil {
				if session.OnSendFailed != nil {
					session.OnSendFailed(session, err)
				} else {
					session.Close(err)
				}
				return
			}
		case packet := <-session.sendPacketChan:
			if err := session.SendPacket(packet); err != nil {
				if session.OnSendFailed != nil {
					session.OnSendFailed(session, err)
				} else {
					session.Close(err)
				}
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

// Get packet protocol.
func (session *Session) Protocol() PacketProtocol {
	return session.protocol
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

// Close session.
func (session *Session) Close(reason interface{}) {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		session.closeReason = reason

		session.conn.Close()

		// exit send loop and cancel async send
		close(session.closeChan)

		session.dispatchCloseEvent()
	}
}

// Loop and read message. NOTE: The callback argument point to internal read buffer.
func (session *Session) ReadLoop(handler func([]byte)) {
	var buffer []byte
	for {
		msg, err := session.Read(buffer)
		if err != nil {
			session.Close(err)
			break
		}
		handler(msg)
		buffer = msg[0:0]
	}
}

// Read message once (Can reuse buffer).
func (session *Session) Read(buffer []byte) ([]byte, error) {
	if len(buffer) != 0 {
		buffer = buffer[0:0]
	}
	msg, err := session.reader.ReadPacket(session.conn, buffer)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// Packet a message.
func (session *Session) Packet(message Message, buff []byte) (packet []byte, err error) {
	size := message.RecommendPacketSize()
	packet = session.writer.BeginPacket(size, buff)
	packet, err = message.AppendToPacket(packet)
	if err != nil {
		return nil, err
	}
	packet = session.writer.EndPacket(packet)
	return
}

// Sync send a message. This method will block on IO.
func (session *Session) Send(message Message) (err error) {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()
	session.sendBuff, err = session.Packet(message, session.sendBuff)
	if err != nil {
		return err
	}
	return session.writer.WritePacket(session.conn, session.sendBuff)
}

// Sync send a packet. The packet must be properly formatted.
// Please see Session.Packet().
func (session *Session) SendPacket(packet []byte) error {
	return session.writer.WritePacket(session.conn, packet)
}

// Async send a message. This method will never block.
// If blocking happens, this method returns BlockingError.
func (session *Session) TrySend(message Message, timeout time.Duration) error {
	if session.IsClosed() {
		return SendToClosedError
	}

	select {
	case session.sendChan <- message:
	case <-session.closeChan:
		return SendToClosedError
	case <-time.After(timeout):
		return BlockingError
	}

	return nil
}

// Try send a message. This method will never block.
// If blocking happens, this method returns BlockingError.
// The packet must be properly formatted.
// Please see Session.Packet().
func (session *Session) TrySendPacket(packet []byte, timeout time.Duration) error {
	if session.IsClosed() {
		return SendToClosedError
	}

	select {
	case session.sendPacketChan <- packet:
	case <-session.closeChan:
		return SendToClosedError
	case <-time.After(timeout):
		return BlockingError
	}

	return nil
}

// The session close event listener interface.
type SessionCloseEventListener interface {
	OnSessionClose(*Session)
}

// Add close event listener.
func (session *Session) AddCloseEventListener(listener SessionCloseEventListener) {
	session.closeEventListeners.PushBack(listener)
}

// Remove close event listener.
func (session *Session) RemoveCloseEventListener(listener SessionCloseEventListener) {
	for i := session.closeEventListeners.Front(); i != nil; i = i.Next() {
		if i.Value == listener {
			session.closeEventListeners.Remove(i)
			return
		}
	}
}

// Dispatch close event.
func (session *Session) dispatchCloseEvent() {
	for i := session.closeEventListeners.Front(); i != nil; i = i.Next() {
		i.Value.(SessionCloseEventListener).OnSessionClose(session)
	}
}
