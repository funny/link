package packnet

import "sync"

// The channel type. Used to maintain a group of session. Normally used for broadcast classify purpose.
type Channel struct {
	server         *Server
	mutex          sync.RWMutex
	broadcastBuff  []byte
	broadcastMutex sync.RWMutex
	sessions       map[uint64]channelSession
}

type channelSession struct {
	*Session
	KickCallback func()
}

// Create a channel instance.
func (server *Server) NewChannel() *Channel {
	return &Channel{
		server:   server,
		sessions: make(map[uint64]channelSession),
	}
}

// How mush sessions in this channel.
func (channel *Channel) Len() int {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	return len(channel.sessions)
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

// Fetch the sessions. NOTE: Invoke Kick() or Exit() in fetch callback will dead lock.
func (channel *Channel) Fetch(callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	for _, sesssion := range channel.sessions {
		callback(sesssion.Session)
	}
}

// Broadcast to sessions. The message only encoded one time
// so the performance it's better than send message one by one.
func (channel *Channel) Broadcast(message Message) {
	channel.broadcastMutex.Lock()
	defer channel.broadcastMutex.Unlock()

	size := message.RecommendPacketSize()

	packet := channel.server.writer.BeginPacket(size, channel.broadcastBuff)
	packet = message.AppendToPacket(packet)
	packet = channel.server.writer.EndPacket(packet)

	channel.broadcastBuff = packet

	channel.Fetch(func(session *Session) {
		session.sendPacket(packet)
	})
}
