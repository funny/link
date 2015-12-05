// +build generate

//go:generate go run channel_gen.go Int32Channel int32 channel_int32.go
//go:generate go run channel_gen.go Uint32Channel uint32 channel_uint32.go
//go:generate go run channel_gen.go Int64Channel int64 channel_int64.go
//go:generate go run channel_gen.go Uint64Channel uint64 channel_uint64.go
//go:generate go run channel_gen.go StringChannel string channel_string.go
package link

import (
	"sync"
)

type Channel struct {
	mutex    sync.RWMutex
	sessions map[KEY]*Session

	// channel state
	State interface{}
}

func NewChannel() *Channel {
	return &Channel{
		sessions: make(map[KEY]*Session),
	}
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

func (channel *Channel) Get(key KEY) *Session {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	session, _ := channel.sessions[key]
	return session
}

func (channel *Channel) Put(key KEY, session *Session) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	if session, exists := channel.sessions[key]; exists {
		channel.remove(key, session)
	}
	session.AddCloseCallback(channel, func() {
		channel.Remove(key)
	})
	channel.sessions[key] = session
}

func (channel *Channel) remove(key KEY, session *Session) {
	session.RemoveCloseCallback(channel)
	delete(channel.sessions, key)
}

func (channel *Channel) Remove(key KEY) bool {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	session, exists := channel.sessions[key]
	if exists {
		channel.remove(key, session)
	}
	return exists
}

func (channel *Channel) Close() {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	for key, session := range channel.sessions {
		channel.remove(key, session)
	}
}
