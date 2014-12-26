package link

import "github.com/funny/sync"

// The session collection use to fetch session and send broadcast.
type SessionCollection interface {
	Protocol() Protocol
	FetchSession(func(*Session))
}

type BroadcastWork struct {
	Session *Session
	AsyncWork
}

// Broadcast to sessions. The message only encoded once
// so the performance is better than send message one by one.
func Broadcast(sessions SessionCollection, message Message) ([]BroadcastWork, error) {
	buffer := NewOutBuffer()
	packet, err := sessions.Protocol().Packet(message, buffer)
	if err != nil {
		return nil, err
	}
	works := make([]BroadcastWork, 0, 10)
	buffer.isBroadcast = true
	sessions.FetchSession(func(session *Session) {
		buffer.broadcastUse()
		works = append(works, BroadcastWork{
			session,
			session.AsyncSendPacket(packet),
		})
	})
	return works, nil
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

// Get channel protocol.
// Implement SessionCollection interface.
func (channel *Channel) Protocol() Protocol {
	return channel.protocol
}

// Fetch the sessions. NOTE: Invoke Kick() or Exit() in fetch callback will dead lock.
// Implement SessionCollection interface.
func (channel *Channel) FetchSession(callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()

	for _, sesssion := range channel.sessions {
		callback(sesssion.Session)
	}
}
