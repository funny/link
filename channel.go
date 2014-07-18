package packnet

import "sync"

// The channel type. Used to maintain a group of session. Normally used for broadcast classify purpose.
type Channel struct {
	mutex    *sync.RWMutex
	sessions map[uint64]channelSession
}

type channelSession struct {
	*Session
	KickCallback func()
}

// Create a channel instance.
func NewChannel() *Channel {
	return &Channel{
		mutex:    new(sync.RWMutex),
		sessions: make(map[uint64]channelSession),
	}
}

// Join the channel. The kickClallback will called when the session kick out from the channel.
func (channel *Channel) Join(session *Session, kickCallback func()) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	channel.sessions[session.Id()] = channelSession{session, kickCallback}
}

// Exit the channel.
func (channel *Channel) Exit(session *Session) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	delete(channel.sessions, session.Id())
}

// Kick out a session from the channel.
func (channel *Channel) Kick(sessionId uint64) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	if session, exists := channel.sessions[sessionId]; exists {
		delete(channel.sessions, sessionId)
		if session.KickCallback != nil {
			session.KickCallback()
		}
	}
}

// Fetch the sessions. Implement the SessionList interface.
// So the channel can use to send broadcast by Server.Broadcast().
func (channel *Channel) Fetch(callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	for _, sesssion := range channel.sessions {
		callback(sesssion.Session)
	}
}
