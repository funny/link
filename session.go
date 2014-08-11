package link

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
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
	sendBuff       []byte
	sendLock       sync.Mutex
	messageHandler MessageHandler

	// About session close
	closeChan     chan int
	closeWait     *sync.WaitGroup
	closeFlag     int32
	closeCallback func(*Session)

	// Put your session state here.
	State interface{}
}

// Create a new session instance.
func NewSession(id uint64, conn net.Conn, protocol PacketProtocol, sendChanSize uint) *Session {
	return &Session{
		id:             id,
		conn:           conn,
		writer:         protocol.NewWriter(),
		reader:         protocol.NewReader(),
		sendChan:       make(chan Message, sendChanSize),
		sendPacketChan: make(chan []byte, sendChanSize),
		closeChan:      make(chan int),
		closeWait:      new(sync.WaitGroup),
		closeFlag:      -1,
	}
}

// Start the session's read write goroutines.
// NOTE: A session always need to started before you use it.
func (session *Session) Start() {
	if atomic.CompareAndSwapInt32(&session.closeFlag, -1, 0) {
		session.closeWait.Add(1)
		go session.writeLoop()

		session.closeWait.Add(1)
		go session.readLoop()
	} else {
		panic(SessionDuplicateStartError)
	}
}

// Loop and wait incoming requests.
func (session *Session) readLoop() {
	defer func() {
		session.closeWait.Done()
		session.Close()
	}()

	var (
		packet []byte
		err    error
	)

	for {
		packet, err = session.reader.ReadPacket(session.conn, packet)
		if err != nil {
			break
		}
		if session.messageHandler != nil {
			session.messageHandler.Handle(session, packet)
		}
	}
}

// Loop and transport responses.
func (session *Session) writeLoop() {
	defer func() {
		session.closeWait.Done()
		session.Close()
	}()
L:
	for {
		select {
		case message := <-session.sendChan:
			if err := session.syncSend(message); err != nil {
				break L
			}
		case packet := <-session.sendPacketChan:
			if err := session.syncSendPacket(packet); err != nil {
				break L
			}
		case <-session.closeChan:
			break L
		}
	}
}

// Sync send a message. This method will block on IO.
func (session *Session) syncSend(message Message) error {
	session.sendLock.Lock()
	defer session.sendLock.Unlock()

	size := message.RecommendPacketSize()

	packet := session.writer.BeginPacket(size, session.sendBuff)
	packet = message.AppendToPacket(packet)
	packet = session.writer.EndPacket(packet)

	session.sendBuff = packet

	err := session.writer.WritePacket(session.conn, packet)
	if err != nil {
		session.Close()
	}

	return err
}

// Sync send a packet. The packet must be properly formatted.
func (session *Session) syncSendPacket(packet []byte) error {
	session.sendLock.Lock()
	defer session.sendLock.Unlock()

	err := session.writer.WritePacket(session.conn, packet)
	if err != nil {
		session.Close()
	}

	return err
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

// Get reader setting.
func (session *Session) ReaderSettings() Settings {
	return session.reader
}

// Get writer setting.
func (session *Session) WriterSettings() Settings {
	return session.writer
}

// Check session is closed or not.
func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) != 0
}

// Set message handler function. A easy way to handle messages.
func (session *Session) OnMessage(callback func(*Session, []byte)) {
	session.messageHandler = messageHandlerFunc{callback}
}

// Set message handler. A complex but more powerful way to handle messages.
func (session *Session) SetMessageHandler(handler MessageHandler) {
	session.messageHandler = handler
}

// Set session close callback.
func (session *Session) OnClose(callback func(*Session)) {
	session.closeCallback = callback
}

// Close session and remove it from api server.
func (session *Session) Close() {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		// if close session without this goroutine
		// deadlock will happen when session close itself.
		go func() {
			session.conn.Close()

			// notify write loop session closed
			close(session.closeChan)

			// wait for read loop and write lopp exit
			session.closeWait.Wait()

			// if this is a server side session
			// remove it from sessin list
			if session.server != nil {
				session.server.delSession(session)
			}

			// trigger the session close event
			if session.closeCallback != nil {
				session.closeCallback(session)
			}
		}()
	}
}

type SendMode uint64

// Example:
// // Async send a message and wait for timeout in 3 seconds, don't close session when timeout.
// session.Send(msg, ASYNC|DO_NOT_CLOSE|TIMEOUT(time.Second * 3))
const (
	SYNC         = SendMode(1 << 0) // Sync send.
	ASYNC        = SendMode(1 << 1) // Async send.
	DO_NOT_CLOSE = SendMode(1 << 2) // Disable auto close when blocking happens.
)

const _TIMEOUT_BITS_ = 60

// Setting the wait blocking timeout.
func TIMEOUT(timeout time.Duration) SendMode {
	return SendMode(timeout << _TIMEOUT_BITS_)
}

// Send a message.
func (session *Session) Send(message Message, mode SendMode) error {
	if session.IsClosed() {
		return SendToClosedError
	}

	switch {
	case mode&SYNC == SYNC:
		session.syncSend(message)
	case mode&ASYNC == ASYNC:
		if timeout := time.Duration(uint64(mode) >> _TIMEOUT_BITS_); timeout != 0 {
			select {
			case session.sendChan <- message:
			case <-time.After(timeout):
				if mode&DO_NOT_CLOSE != DO_NOT_CLOSE {
					session.Close()
					return CloseBlockingError
				}
				return TimeoutBlockingError
			}
		} else {
			select {
			case session.sendChan <- message:
			default:
				if mode&DO_NOT_CLOSE != DO_NOT_CLOSE {
					session.Close()
					return CloseBlockingError
				}
				return DiscardBlockingError
			}
		}
	}

	return nil
}

// Send a packet. The packet must be properly formatted.
func (session *Session) SendPacket(packet []byte, mode SendMode) error {
	if session.IsClosed() {
		return SendToClosedError
	}

	switch {
	case mode&SYNC == SYNC:
		session.syncSendPacket(packet)
	case mode&ASYNC == ASYNC:
		if timeout := time.Duration(uint64(mode) >> _TIMEOUT_BITS_); timeout != 0 {
			select {
			case session.sendPacketChan <- packet:
			case <-time.After(timeout):
				if mode&DO_NOT_CLOSE != DO_NOT_CLOSE {
					session.Close()
					return CloseBlockingError
				}
				return TimeoutBlockingError
			}
		} else {
			select {
			case session.sendPacketChan <- packet:
			default:
				if mode&DO_NOT_CLOSE != DO_NOT_CLOSE {
					session.Close()
					return CloseBlockingError
				}
				return DiscardBlockingError
			}
		}
	}

	return nil
}
