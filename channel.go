package link

import (
	"sync"
)

type SessionFetcher func(func(*Session))

type BroadcastProtocol interface {
	Broadcast(msg interface{}, fetcher SessionFetcher) error
}

type defaultBroadcast struct{}

func (_ defaultBroadcast) Broadcast(msg interface{}, fetcher SessionFetcher) error {
	fetcher(func(session *Session) { session.AsyncSend(msg) })
	return nil
}

type Channel struct {
	protocol BroadcastProtocol
	mutex    sync.RWMutex
	sessions map[uint64]*Session

	// channel state
	State interface{}
}

func NewChannel() *Channel {
	return NewCustomChannel(defaultBroadcast{})
}

func NewCustomChannel(protocol BroadcastProtocol) *Channel {
	channel := &Channel{
		protocol: protocol,
		sessions: make(map[uint64]*Session),
	}
	return channel
}

func (channel *Channel) Broadcast(msg interface{}) error {
	return channel.protocol.Broadcast(msg, channel.Fetch)
}

func (channel *Channel) Len() int {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	return len(channel.sessions)
}

// Fetch the sessions. NOTE: Dead lock happends if invoke Exit() in fetch callback.
func (channel *Channel) Fetch(callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	for _, session := range channel.sessions {
		callback(session)
	}
}

func (channel *Channel) Join(session *Session) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	session.AddCloseCallback(channel, func() { channel.Exit(session) })
	channel.sessions[session.Id()] = session
}

func (channel *Channel) delSession(session *Session) bool {
	_, exists := channel.sessions[session.Id()]
	if exists {
		session.RemoveCloseCallback(channel)
		delete(channel.sessions, session.Id())
	}
	return exists
}

func (channel *Channel) Exit(session *Session) bool {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	return channel.delSession(session)
}

func (channel *Channel) Close() {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	for _, session := range channel.sessions {
		channel.delSession(session)
	}
}
