package link

// The session collection use to fetch session and send broadcast.
type SessionCollection interface {
	Fetch(func(*Session))
	Protocol() Protocol
}

// Broadcast to sessions. The message only encoded once
// so the performance better then send message one by one.
func Broadcast(sessions SessionCollection, message Message) error {
	buffer := sessions.Protocol().BufferFactory().NewOutBuffer()
	sessions.Protocol().Prepare(buffer, message)
	if err := message.WriteBuffer(buffer); err != nil {
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
	buffer := sessions.Protocol().BufferFactory().NewOutBuffer()
	sessions.Protocol().Prepare(buffer, message)
	if err := message.WriteBuffer(buffer); err != nil {
		return err
	}
	sessions.Fetch(func(session *Session) {
		session.SendPacket(buffer)
	})
	return nil
}
