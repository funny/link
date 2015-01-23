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
func Dial(network, address string) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&dialSessionId, 1)
	session := NewSession(id, conn, DefaultProtocol, CLIENT_SIDE, DefaultSendChanSize, DefaultConnBufferSize)
	return session, nil
}

// The easy way to create a connection with timeout setting.
func DialTimeout(network, address string, timeout time.Duration) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&dialSessionId, 1)
	session := NewSession(id, conn, DefaultProtocol, CLIENT_SIDE, DefaultSendChanSize, DefaultConnBufferSize)
	return session, nil
}

type Decoder func(*InBuffer) error

// Session.
type Session struct {
	id uint64

	// About network
	conn     net.Conn
	protocol ProtocolState

	// About send and receive
	readMutex           sync.Mutex
	sendMutex           sync.Mutex
	asyncSendChan       chan asyncMessage
	asyncSendBufferChan chan asyncBuffer
	inBufferMutex       sync.Mutex
	inBuffer            *InBuffer
	outBuffer           *OutBuffer

	// About session close
	closeChan       chan int
	closeFlag       int32
	closeEventMutex sync.Mutex
	closeCallbacks  *list.List

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
func NewSession(id uint64, conn net.Conn, protocol Protocol, side ProtocolSide, sendChanSize int, readBufferSize int) *Session {
	if readBufferSize > 0 {
		conn = newBufferConn(conn, readBufferSize)
	}

	protocolState := protocol.New(conn, side)
	if protocolState == nil {
		return nil
	}

	session := &Session{
		id:                  id,
		conn:                conn,
		protocol:            protocolState,
		asyncSendChan:       make(chan asyncMessage, sendChanSize),
		asyncSendBufferChan: make(chan asyncBuffer, sendChanSize),
		inBuffer:            newInBuffer(),
		outBuffer:           newOutBuffer(),
		closeChan:           make(chan int),
		closeCallbacks:      list.New(),
	}

	go session.sendLoop()

	return session
}

// Get session id.
func (session *Session) Id() uint64 {
	return session.id
}

// Get session connection.
func (session *Session) Conn() net.Conn {
	return session.conn
}

// Check session is closed or not.
func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) != 0
}

// Close session.
func (session *Session) Close() {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		session.conn.Close()

		// exit send loop and cancel async send
		close(session.closeChan)

		session.invokeCloseCallbacks()

		session.inBuffer.free()
		session.outBuffer.free()
	}
}

// Sync send a message from encoder func. This method will block on IO.
func (session *Session) SendFunc(e func(*OutBuffer) error) error {
	return session.Send(encoder(e))
}

// Sync send a message. This method will block on IO.
func (session *Session) Send(message Message) error {
	session.inBufferMutex.Lock()
	defer session.inBufferMutex.Unlock()

	var err error

	buffer := session.outBuffer
	session.protocol.PrepareOutBuffer(buffer, message.OutBufferSize())

	err = message.WriteOutBuffer(buffer)
	if err == nil {
		err = session.sendBuffer(buffer)
	}

	buffer.reset()
	return err
}

func (session *Session) sendBuffer(buffer *OutBuffer) error {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()

	return session.protocol.Write(session.conn, buffer)
}

// Process one request.
func (session *Session) ProcessOnce(decoder Decoder) error {
	session.readMutex.Lock()
	defer session.readMutex.Unlock()

	buffer := session.inBuffer
	err := session.protocol.Read(session.conn, buffer)
	if err != nil {
		buffer.reset()
		session.Close()
		return err
	}

	err = decoder(buffer)
	buffer.reset()

	return nil
}

// Process request.
func (session *Session) Process(decoder Decoder) error {
	for {
		if err := session.ProcessOnce(decoder); err != nil {
			return err
		}
	}
	return nil
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

type asyncBuffer struct {
	C chan<- error
	B *OutBuffer
}

// Loop and transport responses.
func (session *Session) sendLoop() {
	for {
		select {
		case buffer := <-session.asyncSendBufferChan:
			buffer.C <- session.sendBuffer(buffer.B)
			buffer.B.broadcastFree()
		case message := <-session.asyncSendChan:
			message.C <- session.Send(message.M)
		case <-session.closeChan:
			return
		}
	}
}

// Async send a message from encoder.
func (session *Session) AsyncSendFunc(e func(*OutBuffer) error) AsyncWork {
	return session.AsyncSend(encoder(e))
}

// Async send a message.
func (session *Session) AsyncSend(message Message) AsyncWork {
	c := make(chan error, 1)
	if session.IsClosed() {
		c <- SendToClosedError
	} else {
		select {
		case session.asyncSendChan <- asyncMessage{c, message}:
		default:
			go func() {
				select {
				case session.asyncSendChan <- asyncMessage{c, message}:
				case <-session.closeChan:
					c <- SendToClosedError
				case <-time.After(time.Second * 5):
					session.Close()
					c <- AsyncSendTimeoutError
				}
			}()
		}
	}
	return AsyncWork{c}
}

// Async send a packet.
func (session *Session) asyncSendBuffer(buffer *OutBuffer) AsyncWork {
	c := make(chan error, 1)
	if session.IsClosed() {
		c <- SendToClosedError
	} else {
		select {
		case session.asyncSendBufferChan <- asyncBuffer{c, buffer}:
		default:
			go func() {
				select {
				case session.asyncSendBufferChan <- asyncBuffer{c, buffer}:
				case <-session.closeChan:
					c <- SendToClosedError
				case <-time.After(time.Second * 5):
					session.Close()
					c <- AsyncSendTimeoutError
				}
			}()
		}
	}
	return AsyncWork{c}
}

type closeCallback struct {
	Handler interface{}
	Func    func()
}

// Add close callback.
func (session *Session) AddCloseCallback(handler interface{}, callback func()) {
	if session.IsClosed() {
		return
	}

	session.closeEventMutex.Lock()
	defer session.closeEventMutex.Unlock()

	session.closeCallbacks.PushBack(closeCallback{handler, callback})
}

// Remove close callback.
func (session *Session) RemoveCloseCallback(handler interface{}) {
	if session.IsClosed() {
		return
	}

	session.closeEventMutex.Lock()
	defer session.closeEventMutex.Unlock()

	for i := session.closeCallbacks.Front(); i != nil; i = i.Next() {
		if i.Value.(closeCallback).Handler == handler {
			session.closeCallbacks.Remove(i)
			return
		}
	}
}

// Dispatch close event.
func (session *Session) invokeCloseCallbacks() {
	session.closeEventMutex.Lock()
	defer session.closeEventMutex.Unlock()

	for i := session.closeCallbacks.Front(); i != nil; i = i.Next() {
		callback := i.Value.(closeCallback)
		callback.Func()
	}
}
