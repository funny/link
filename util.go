package link

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	DefaultConfig = Config{
		ConnConfig{
			ReadBufferSize:  2048,
			WriteBufferSize: 2048,
		},
		SessionConfig{
			AutoFlush:         true,
			AsyncSendTimeout:  0,
			AsyncSendChanSize: 1000,
		},
	}
)

var (
	globalSessionInited int32
	gloablSessionId     uint64
	globalSessions      map[uint64]*Session
	globalSessionMutex  sync.Mutex
)

func globalSessionInit() {
	if atomic.CompareAndSwapInt32(&globalSessionInited, 0, 1) {
		globalSessions = make(map[uint64]*Session)
		go globalSessionCheckAlive()
	}
}

func newGlobalSession(conn *Conn) *Session {
	globalSessionInit()

	id := atomic.AddUint64(&gloablSessionId, 1)
	session := NewSession(id, conn, DefaultConfig.SessionConfig)

	globalSessionMutex.Lock()
	defer globalSessionMutex.Unlock()
	globalSessions[id] = session

	session.AddCloseCallback(globalSessions, func() {
		globalSessionMutex.Lock()
		defer globalSessionMutex.Unlock()
		delete(globalSessions, session.Id())
	})

	return session
}

func globalSessionFetcher(callback func(session *Session)) {
	globalSessionMutex.Lock()
	defer globalSessionMutex.Unlock()

	for _, session := range globalSessions {
		callback(session)
	}
}

func globalSessionCheckAlive() {
	tick := time.NewTicker(time.Second)
	for range tick.C {
		now := time.Now()
		globalSessionFetcher(func(session *Session) {
			if session.IsTimeout(now) {
				go session.Close()
			}
		})
	}
}

func Listen(network, address string) (*Listener, error) {
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewListener(l, DefaultConfig.ConnConfig), nil
}

func Serve(network, address string) (*Server, error) {
	listener, err := Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, DefaultConfig), nil
}

func Dial(network, address string) (*Conn, error) {
	c, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewConn(c, DefaultConfig.ConnConfig), nil
}

func Connect(network, address string) (*Session, error) {
	conn, err := Dial(network, address)
	if err != nil {
		return nil, err
	}
	return newGlobalSession(conn), nil
}

func DialTimeout(network, address string, timeout time.Duration) (*Conn, error) {
	c, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	return NewConn(c, DefaultConfig.ConnConfig), nil
}

func ConnectTimeout(network, address string, timeout time.Duration) (*Session, error) {
	conn, err := DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	return newGlobalSession(conn), nil
}

type JSON struct {
	V interface{}
	S Spliter
}

func (j JSON) Send(conn *Conn) error {
	if limiter, ok := j.S.(Limiter); ok {
		return json.NewDecoder(limiter.Limit(conn)).Decode(j.V)
	}
	b, err := json.Marshal(j.V)
	if err != nil {
		return err
	}
	conn.WritePacket(b, j.S)
	return nil
}

func (j JSON) Receive(conn *Conn) error {
	b := conn.ReadPacket(j.S)
	return json.Unmarshal(b, j.V)
}

type GOB struct {
	V interface{}
	S Spliter
}

func (g GOB) Send(conn *Conn) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(g.V); err != nil {
		return err
	}
	conn.WritePacket(buf.Bytes(), g.S)
	return nil
}

func (g GOB) Receive(conn *Conn) error {
	if limiter, ok := g.S.(Limiter); ok {
		return gob.NewDecoder(limiter.Limit(conn)).Decode(g.V)
	}
	b := conn.ReadPacket(g.S)
	r := bytes.NewReader(b)
	return gob.NewDecoder(r).Decode(g.V)
}
