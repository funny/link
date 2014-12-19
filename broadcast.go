package link

import "github.com/funny/sync"

// The session collection use to fetch session and send broadcast.
type SessionCollection interface {
	Protocol() Protocol
	Fetch(func(*Session))
}

// Broadcast to sessions. The message only encoded once
// so the performance better then send message one by one.
func Broadcast(sessions SessionCollection, message Message) error {
	buffer, err := sessions.Protocol().Packet(make([]byte, 0), message)
	if err != nil {
		return err
	}
	sessions.Fetch(func(session *Session) {
		session.TrySendPacket(buffer, 0)
	})
	return nil
}

// Broadcast to sessions. The message only encoded once
// so the performance better then send message one by one.
func MustBroadcast(sessions SessionCollection, message Message) error {
	buffer, err := sessions.Protocol().Packet(make([]byte, 0), message)
	if err != nil {
		return err
	}
	sessions.Fetch(func(session *Session) {
		session.SendPacket(buffer)
	})
	return nil
}

// The channel type. Used to maintain a group of session.
// Normally used for broadcast classify purpose.
type Channel struct {
	mutex    sync.RWMutex
	protocol Protocol
	sessions map[uint64]channelSession
}

type channelSession struct {
	*Session
	KickCallback func()
}

// Create a channel instance.
func NewChannel(protocol Protocol) *Channel {
	return &Channel{
		protocol: protocol,
		sessions: make(map[uint64]channelSession),
	}
}

// How mush sessions in this channel.
func (channel *Channel) Len() int {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()

	return len(channel.sessions)
}

// Join the channel. The kickCallback will called when the session kick out from the channel.
func (channel *Channel) Join(session *Session, kickCallback func()) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	session.AddCloseEventListener(channel)
	channel.sessions[session.Id()] = channelSession{session, kickCallback}
}

// Implement the SessionCloseEventListener interface.
func (channel *Channel) OnSessionClose(session *Session) {
	channel.Exit(session)
}

// Exit the channel.
func (channel *Channel) Exit(session *Session) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	session.RemoveCloseEventListener(channel)
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

// Fetch the sessions. NOTE: Invoke Kick() or Exit() in fetch callback will dead lock.
func (channel *Channel) Fetch(callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()

	for _, sesssion := range channel.sessions {
		callback(sesssion.Session)
	}
}

// Get channel protocol.
func (channel *Channel) Protocol() Protocol {
	return channel.protocol
}
