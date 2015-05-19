package link

import (
	"bufio"
	"container/list"
	"github.com/funny/sync"
	"net"
	"sync/atomic"
	"time"
)

var globalSessionId uint64

// The easy way to create a connection.
func Dial(network, address string, protocol Protocol, pool *MemPool) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&globalSessionId, 1)
	session := NewSession(id, conn, protocol.NewCodec(), pool, DefaultConfig)
	err = session.Handshake()
	if err != nil {
		return nil, err
	}
	return session, nil
}

// The easy way to create a connection with timeout setting.
func DialTimeout(network, address string, timeout time.Duration, protocol Protocol, pool *MemPool) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&globalSessionId, 1)
	session := NewSession(id, conn, protocol.NewCodec(), pool, DefaultConfig)
	err = session.Handshake()
	if err != nil {
		return nil, err
	}
	return session, nil
}

type Config struct {
	SendChanSize     int
	ReadBufferSize   int
	InBufferSize     int
	OutBufferSize    int
	AsyncSendTimeout time.Duration
}

var DefaultConfig = Config{
	SendChanSize:     1024,
	ReadBufferSize:   1024,
	InBufferSize:     2048,
	OutBufferSize:    2048,
	AsyncSendTimeout: 0,
}

// Session.
type Session struct {
	id uint64

	// About network
	codec Codec
	conn  *Conn

	// About send and receive
	readMutex          sync.Mutex
	sendMutex          sync.Mutex
	asyncMessageChan   chan asyncMessage
	asyncBroadcastChan chan asyncBroadcast
	asyncSendTimeout   time.Duration
	inBuffer           *Buffer
	outBuffer          *Buffer

	// About session close
	closeChan       chan int
	closeFlag       int32
	closeEventMutex sync.Mutex
	closeCallbacks  *list.List

	// Put your session state here.
	State interface{}
}

// Buffered connection.
type Conn struct {
	conn   net.Conn
	Reader *bufio.Reader
}

func NewConn(conn net.Conn, readBufferSize int) *Conn {
	return &Conn{
		conn,
		bufio.NewReaderSize(conn, readBufferSize),
	}
}

func (conn *Conn) Write(b []byte) (int, error) {
	return conn.conn.Write(b)
}

func (conn *Conn) Read(b []byte) (int, error) {
	return conn.Reader.Read(b)
}

func (conn *Conn) Close() error {
	return conn.conn.Close()
}

// Create a new session instance.
func NewSession(id uint64, conn net.Conn, codec Codec, pool *MemPool, config Config) *Session {
	session := &Session{
		id:                 id,
		codec:              codec,
		conn:               NewConn(conn, config.ReadBufferSize),
		asyncMessageChan:   make(chan asyncMessage, config.SendChanSize),
		asyncBroadcastChan: make(chan asyncBroadcast, config.SendChanSize),
		asyncSendTimeout:   config.AsyncSendTimeout,
		inBuffer:           NewPoolBuffer(0, config.InBufferSize, pool),
		outBuffer:          NewPoolBuffer(0, config.OutBufferSize, pool),
		closeChan:          make(chan int),
		closeCallbacks:     list.New(),
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
	return session.conn.conn
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

func (session *Session) Handshake() error {
	if codec, ok := session.codec.(Handshake); ok {
		return codec.Handshake(session.conn, session.inBuffer)
	}
	return nil
}

// Sync send a message. This method will block on IO.
func (session *Session) Send(msg Message) error {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()

	if err := session.codec.SendMessage(session.conn, session.outBuffer, msg); err != nil {
		session.Close()
		return err
	}

	return nil
}

// Process one request.
func (session *Session) ProcessOnce(handler RequestHandler) error {
	session.readMutex.Lock()
	defer session.readMutex.Unlock()

	if err := session.codec.ProcessRequest(session.conn, session.inBuffer, handler); err != nil {
		session.Close()
		return err
	}

	return nil
}

// Loop and process requests.
func (session *Session) Process(handler RequestHandler) error {
	for {
		if err := session.ProcessOnce(handler); err != nil {
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

type asyncBroadcast struct {
	C chan<- error
	B *broadcast
}

// Loop and transport responses.
func (session *Session) sendLoop() {
	for {
		select {
		case message := <-session.asyncMessageChan:
			message.C <- session.Send(message.M)
		case broadcast := <-session.asyncBroadcastChan:
			broadcast.C <- session.sendBroadcast(broadcast.B.Buffer)
			broadcast.B.Free()
		case <-session.closeChan:
			return
		}
	}
}

func (session *Session) sendBroadcast(buf *Buffer) error {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()

	err := session.codec.SendBroadcast(session.conn, buf)
	if err != nil {
		session.Close()
	}
	return err
}

// Async send a response.
func (session *Session) AsyncSend(msg Message) AsyncWork {
	c := make(chan error, 1)

	if session.IsClosed() {
		c <- SendToClosedError
		return AsyncWork{c}
	}

	select {
	case session.asyncMessageChan <- asyncMessage{c, msg}:
	default:
		if session.asyncSendTimeout == 0 {
			c <- AsyncSendTimeoutError
			session.Close()
		} else {
			go func() {
				select {
				case session.asyncMessageChan <- asyncMessage{c, msg}:
				case <-session.closeChan:
					c <- SendToClosedError
				case <-time.After(session.asyncSendTimeout):
					session.Close()
					c <- AsyncSendTimeoutError
				}
			}()
		}
	}

	return AsyncWork{c}
}

// Async send a packet.
func (session *Session) asyncBroadcast(bc *broadcast) BroadcastWork {
	c := make(chan error, 1)

	if session.IsClosed() {
		c <- SendToClosedError
		return BroadcastWork{session, c}
	}

	select {
	case session.asyncBroadcastChan <- asyncBroadcast{c, bc}:
	default:
		if session.asyncSendTimeout == 0 {
			c <- AsyncSendTimeoutError
			session.Close()
		} else {
			go func() {
				select {
				case session.asyncBroadcastChan <- asyncBroadcast{c, bc}:
				case <-session.closeChan:
					c <- SendToClosedError
				case <-time.After(session.asyncSendTimeout):
					session.Close()
					c <- AsyncSendTimeoutError
				}
			}()
		}
	}

	return BroadcastWork{session, c}
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
