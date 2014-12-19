package link

import (
	"bufio"
	"container/list"
	"github.com/funny/sync"
	"net"
	"sync/atomic"
	"time"
)

var dialSessionId uint64

// The easy way to create a connection.
func Dial(network, address string, protocol Protocol) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&dialSessionId, 1)
	session := NewSession(id, conn, protocol, DefaultSendChanSize, DefaultConnBufferSize)
	return session, nil
}

// The easy way to create a connection with timeout setting.
func DialTimeout(network, address string, timeout time.Duration, protocol Protocol) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&dialSessionId, 1)
	session := NewSession(id, conn, protocol, DefaultSendChanSize, DefaultConnBufferSize)
	return session, nil
}

// Session.
type Session struct {
	id uint64

	// About network
	conn     net.Conn
	protocol Protocol

	// About send and receive
	sendChan       chan Message
	sendPacketChan chan *Buffer
	readMutex      sync.Mutex
	sendMutex      sync.Mutex
	OnSendFailed   func(*Session, error)

	// About session close
	closeChan           chan int
	closeFlag           int32
	closeReason         interface{}
	closeEventMutex     sync.Mutex
	closeEventListeners *list.List

	// Put your session state here.
	State interface{}
}

// Buffered connection.
type bufferConn struct {
	net.Conn
	reader *bufio.Reader
}

func newBufferConn(conn net.Conn, readBufferSize int) *bufferConn {
	return &bufferConn{
		conn,
		bufio.NewReaderSize(conn, readBufferSize),
	}
}

func (conn *bufferConn) Read(d []byte) (int, error) {
	return conn.reader.Read(d)
}

// Create a new session instance.
func NewSession(id uint64, conn net.Conn, protocol Protocol, sendChanSize int, readBufferSize int) *Session {
	if readBufferSize > 0 {
		conn = newBufferConn(conn, readBufferSize)
	}

	session := &Session{
		id:                  id,
		conn:                conn,
		protocol:            protocol,
		sendChan:            make(chan Message, sendChanSize),
		sendPacketChan:      make(chan *Buffer, sendChanSize),
		closeChan:           make(chan int),
		closeEventListeners: list.New(),
	}

	go session.sendLoop()

	return session
}

// Get session id.
func (session *Session) Id() uint64 {
	return session.id
}

// Get local address.
func (session *Session) Conn() net.Conn {
	return session.conn
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

// Read a message.
func (session *Session) Read() (*Buffer, error) {
	var buffer = new(Buffer)
	if err := session.ReadReuse(buffer); err != nil {
		return nil, err
	}
	return buffer, nil
}

// Read a message with buffer reuse.
// You can reuse a buffer for reading or just set buffer as nil is OK.
// About the buffer reusing, please see Read() and Handle().
func (session *Session) ReadReuse(buffer *Buffer) error {
	if buffer == nil {
		panic(NilBufferError)
	}

	session.readMutex.Lock()
	defer session.readMutex.Unlock()

	return session.protocol.Read(session.conn, buffer)
}

// Packet a message.
func (session *Session) Packet(message Message) (*Buffer, error) {
	var buffer = new(Buffer)
	if err := session.protocol.Packet(buffer, message); err != nil {
		return nil, err
	}
	return buffer, nil
}

// Packet a message with buffer reuse.
func (session *Session) PacketReuse(message Message, buffer *Buffer) error {
	if buffer == nil {
		panic(NilBufferError)
	}
	return session.protocol.Packet(buffer, message)
}

// Sync send a message. Equals Packet() and SendPacket(). This method will block on IO.
func (session *Session) Send(message Message) error {
	return session.SendReuse(message, &Buffer{})
}

// Sync send a message with buffer reuse.
// Equals Packet() and SendPacket().
// NOTE 1: This method will block on IO.
// NOTE 2: You can reuse a buffer for sending or just set buffer as nil is OK.
// About the buffer reusing, please see Send() and sendLoop().
func (session *Session) SendReuse(message Message, buffer *Buffer) error {
	if err := session.PacketReuse(message, buffer); err != nil {
		return err
	}
	return session.SendPacket(buffer)
}

// Sync send a packet. The packet must be properly formatted.
// Please see Packet().
func (session *Session) SendPacket(packet *Buffer) error {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()
	return session.protocol.Write(session.conn, packet)
}

// Loop and read message. NOTE: The callback argument point to internal read buffer.
func (session *Session) Handle(handler func(*Buffer)) {
	var buffer = &Buffer{}
	for {
		if err := session.ReadReuse(buffer); err != nil {
			session.Close(err)
			break
		}
		handler(buffer)
	}
}

// Loop and transport responses.
func (session *Session) sendLoop() {
	var buffer = &Buffer{}
	for {
		select {
		case message := <-session.sendChan:
			if err := session.SendReuse(message, buffer); err != nil {
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

// Try async send a message.
// If send chan block until timeout happens, this method returns BlockingError.
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

// Try async send a packet.
// If send chan block until timeout happens, this method returns BlockingError.
// The packet must be properly formatted. Please see Session.Packet().
func (session *Session) TrySendPacket(packet *Buffer, timeout time.Duration) error {
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
	if session.IsClosed() {
		return
	}

	session.closeEventMutex.Lock()
	defer session.closeEventMutex.Unlock()

	session.closeEventListeners.PushBack(listener)
}

// Remove close event listener.
func (session *Session) RemoveCloseEventListener(listener SessionCloseEventListener) {
	if session.IsClosed() {
		return
	}

	session.closeEventMutex.Lock()
	defer session.closeEventMutex.Unlock()

	for i := session.closeEventListeners.Front(); i != nil; i = i.Next() {
		if i.Value == listener {
			session.closeEventListeners.Remove(i)
			return
		}
	}
}

// Dispatch close event.
func (session *Session) dispatchCloseEvent() {
	session.closeEventMutex.Lock()
	defer session.closeEventMutex.Unlock()

	for i := session.closeEventListeners.Front(); i != nil; i = i.Next() {
		i.Value.(SessionCloseEventListener).OnSessionClose(session)
	}
}
