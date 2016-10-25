package link

import (
	"sync"
	"sync/atomic"
)

const sessionMapNum = 32

type Manager struct {
	sessionMaps [sessionMapNum]sessionMap
	disposeOnce sync.Once
	disposeWait sync.WaitGroup
}

type sessionMap struct {
	sync.RWMutex
	sessions    map[uint64]*Session
	disposeFlag uint32
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
			atomic.StoreUint32(&smap.disposeFlag, 1)
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

func (manager *Manager) GetSession(sessionID uint64) *Session {
	smap := &manager.sessionMaps[sessionID%sessionMapNum]
	smap.RLock()
	defer smap.RUnlock()

	session, _ := smap.sessions[sessionID]
	return session
}

func (manager *Manager) putSession(session *Session) {
	smap := &manager.sessionMaps[session.id%sessionMapNum]
	smap.Lock()
	defer smap.Unlock()

	smap.sessions[session.id] = session
	manager.disposeWait.Add(1)
}

func (manager *Manager) delSession(session *Session) {
	smap := &manager.sessionMaps[session.id%sessionMapNum]

	if atomic.LoadUint32(&smap.disposeFlag) == 1 {
		manager.disposeWait.Done()
		return
	}

	smap.Lock()
	defer smap.Unlock()

	delete(smap.sessions, session.id)
	manager.disposeWait.Done()
}
