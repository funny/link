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
	return NewSession(id, conn, protocol.NewCodec(), pool, DefaultConfig)
}

// The easy way to create a connection with timeout setting.
func DialTimeout(network, address string, timeout time.Duration, protocol Protocol, pool *MemPool) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&globalSessionId, 1)
	return NewSession(id, conn, protocol.NewCodec(), pool, DefaultConfig)
}

type Config struct {
	SendChanSize       int
	ReadBufferSize     int
	RequestBufferSize  int
	ResponseBufferSize int
	AsyncSendTimeout   time.Duration
}

var DefaultConfig = Config{
	SendChanSize:       1024,
	ReadBufferSize:     1024,
	RequestBufferSize:  2048,
	ResponseBufferSize: 2048,
	AsyncSendTimeout:   0,
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
func NewSession(id uint64, conn net.Conn, codec Codec, pool *MemPool, config Config) (*Session, error) {
	session := &Session{
		id:                 id,
		codec:              codec,
		conn:               NewConn(conn, config.ReadBufferSize),
		asyncMessageChan:   make(chan asyncMessage, config.SendChanSize),
		asyncBroadcastChan: make(chan asyncBroadcast, config.SendChanSize),
		asyncSendTimeout:   config.AsyncSendTimeout,
		inBuffer:           NewPoolBuffer(0, config.RequestBufferSize, pool),
		outBuffer:          NewPoolBuffer(0, config.ResponseBufferSize, pool),
		closeChan:          make(chan int),
		closeCallbacks:     list.New(),
	}

	go session.sendLoop()

	return session, nil
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

func (session *Session) handshake() error {
	return session.codec.Handshake(session.conn, session.inBuffer)
}

// Sync send a message. This method will block on IO.
func (session *Session) Send(msg Message) error {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()

	var err error

	for {
		session.codec.Prepend(session.outBuffer, msg)

		if err = msg.WriteBuffer(session.outBuffer); err != nil {
			break
		}

		if err = session.codec.Write(session.conn, session.outBuffer); err != nil {
			break
		}

		if frame, ok := msg.(FrameMessage); ok && !frame.IsFinalFrame() {
			msg = frame.NextFrame()
			continue
		}
		break
	}

	if err != nil {
		session.Close()
	}
	return err
}

func (session *Session) sendBuffer(buffer *Buffer) error {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()

	err := session.codec.Write(session.conn, buffer)
	if err != nil {
		session.Close()
	}
	return err
}

// Receive one request.
func (session *Session) Receive(decoder Decoder) (Request, error) {
	session.readMutex.Lock()
	defer session.readMutex.Unlock()

	var (
		frames FrameRequest
		req    Request
		err    error
	)

	for {
		err = session.codec.Read(session.conn, session.inBuffer)
		if err != nil {
			break
		}

		req, err = decoder.Decode(session.inBuffer)
		if err != nil {
			break
		}

		if frame, ok := req.(Frame); ok {
			frames = append(frames, req)
			if !frame.IsFinalFrame() {
				continue
			}
		}
		break
	}

	if err != nil {
		session.Close()
		return nil, err
	}

	if frames != nil {
		return frames, err
	}

	return req, err
}

// Loop and process requests.
func (session *Session) Process(decoder Decoder) error {
	for {
		req, err := session.Receive(decoder)
		if err != nil {
			return err
		}
		if req != nil {
			err = req.Process(session)
		}
		if err != nil {
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
		case buffer := <-session.asyncBroadcastChan:
			buffer.C <- session.sendBuffer(buffer.B.Buffer)
			buffer.B.Free()
		case <-session.closeChan:
			return
		}
	}
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
