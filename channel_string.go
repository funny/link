// DO NOT EDIT
// GENERATE BY 'go run channel_gen.go StringChannel string channel_string.go'
package link

import (
	"sync"
)

type StringChannel struct {
	mutex    sync.RWMutex
	sessions map[string]*Session

	// channel state
	State interface{}
}

func NewStringChannel() *StringChannel {
	return &StringChannel{
		sessions: make(map[string]*Session),
	}
}

func (channel *StringChannel) Len() int {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	return len(channel.sessions)
}

// Fetch the sessions. NOTE: Dead lock happends if invoke Exit() in fetch callback.
func (channel *StringChannel) Fetch(callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	for _, session := range channel.sessions {
		callback(session)
	}
}

func (channel *StringChannel) Get(key string) *Session {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	session, _ := channel.sessions[key]
	return session
}

func (channel *StringChannel) Put(key string, session *Session) {
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

func (channel *StringChannel) remove(key string, session *Session) {
	session.RemoveCloseCallback(channel)
	delete(channel.sessions, key)
}

func (channel *StringChannel) Remove(key string) bool {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	session, exists := channel.sessions[key]
	if exists {
		channel.remove(key, session)
	}
	return exists
}

func (channel *StringChannel) Close() {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	for key, session := range channel.sessions {
		channel.remove(key, session)
	}
}
