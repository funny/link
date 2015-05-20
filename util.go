package link

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
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
