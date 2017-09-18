package link

import "sync"

const sessionMapNum = 32

type Manager struct {
	sessionMaps [sessionMapNum]sessionMap
	disposeOnce sync.Once
	disposeWait sync.WaitGroup
}

type sessionMap struct {
	sync.RWMutex
	sessions map[uint64]*Session
	disposed bool
}

func NewManager() *Manager {
	manager := &Manager{}
	for i := 0; i < len(manager.sessionMaps); i++ {
		manager.sessionMaps[i].sessions = make(map[uint64]*Session)
	}
	return manager
}

func (manager *Manager) Dispose() {
	manager.disposeOnce.Do(func() {
		for i := 0; i < sessionMapNum; i++ {
			smap := &manager.sessionMaps[i]
			smap.Lock()
			smap.disposed = true
			for _, session := range smap.sessions {
				session.Close()
			}
			smap.Unlock()
		}
		manager.disposeWait.Wait()
	})
}

func (manager *Manager) NewSession(codec Codec, sendChanSize int) *Session {
	session := newSession(manager, codec, sendChanSize)
	manager.putSession(session)
	return session
}

func (manager *Manager) GetSession(sessionID uint64) (session *Session) {
	smap := &manager.sessionMaps[sessionID%sessionMapNum]
	smap.RLock()
	session, _ = smap.sessions[sessionID]
	smap.RUnlock()
	return
}

func (manager *Manager) putSession(session *Session) {
	smap := &manager.sessionMaps[session.id%sessionMapNum]

	if smap.disposed {
		session.Close()
		return
	}

	smap.Lock()
	smap.sessions[session.id] = session
	smap.Unlock()
	manager.disposeWait.Add(1)
}

func (manager *Manager) delSession(session *Session) {
	smap := &manager.sessionMaps[session.id%sessionMapNum]

	smap.Lock()
	delete(smap.sessions, session.id)
	smap.Unlock()
	manager.disposeWait.Done()
}
