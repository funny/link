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
	readMutex   sync.Mutex
	sendMutex   sync.Mutex
	packetChan  chan asyncPacket
	messageChan chan asyncMessage

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
		packetChan:          make(chan asyncPacket, sendChanSize),
		messageChan:         make(chan asyncMessage, sendChanSize),
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
func (session *Session) Read() (*InBuffer, error) {
	session.readMutex.Lock()
	defer session.readMutex.Unlock()

	return session.protocol.Read(session.conn)
}

// Sync send a message. Equals Packet() then SendPacket(). This method will block on IO.
func (session *Session) Send(message Message) error {
	packet, err := session.Packet(message)
	if err != nil {
		return err
	}
	err = session.SendPacket(packet)
	packet.Free()
	return err
}

// Packet a message. The packet buffer need to free by manual.
func (session *Session) Packet(message Message) (Packet, error) {
	return session.protocol.Packet(message)
}

// Sync send a packet. See Packet() method.
func (session *Session) SendPacket(packet Packet) error {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()
	return session.protocol.Write(session.conn, packet)
}

// Loop and read message.
func (session *Session) Handle(handler func(*InBuffer)) {
	for {
		buffer, err := session.Read()
		if err != nil {
			session.Close(err)
			break
		}
		handler(buffer)
		buffer.Free()
	}
}

// Async work.
type AsyncWork struct {
	c <-chan error
}

// Wait work done. Returns error when work failed.
func (aw AsyncWork) Wait() error {
	return <-aw.c
}

type asyncMessage struct {
	C chan<- error
	M Message
}

type asyncPacket struct {
	C chan<- error
	P Packet
}

// Loop and transport responses.
func (session *Session) sendLoop() {
	for {
		select {
		case packet := <-session.packetChan:
			packet.C <- session.SendPacket(packet.P)
			packet.P.broadcastFree()
		case message := <-session.messageChan:
			message.C <- session.Send(message.M)
		case <-session.closeChan:
			return
		}
	}
}

// Async send a message.
func (session *Session) AsyncSend(message Message) AsyncWork {
	c := make(chan error, 1)
	if session.IsClosed() {
		c <- SendToClosedError
	} else {
		select {
		case session.messageChan <- asyncMessage{c, message}:
		case <-session.closeChan:
			c <- SendToClosedError
		default:
			go func() {
				select {
				case session.messageChan <- asyncMessage{c, message}:
				case <-time.After(time.Second * 5):
					session.Close(AsyncSendTimeoutError)
					c <- AsyncSendTimeoutError
				}
			}()
		}
	}
	return AsyncWork{c}
}

// Async send a packet.
func (session *Session) AsyncSendPacket(packet Packet) AsyncWork {
	c := make(chan error, 1)
	if session.IsClosed() {
		c <- SendToClosedError
	} else {
		select {
		case session.packetChan <- asyncPacket{c, packet}:
		case <-session.closeChan:
			c <- SendToClosedError
		default:
			go func() {
				select {
				case session.packetChan <- asyncPacket{c, packet}:
				case <-time.After(time.Second * 5):
					session.Close(AsyncSendTimeoutError)
					c <- AsyncSendTimeoutError
				}
			}()
		}
	}
	return AsyncWork{c}
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
