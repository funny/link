package link

import (
	"sync"
)

const sessionGroupNum = 1 << 5

type Manager struct {
	sessgroups  [sessionGroupNum]sessgroup
	disposeOnce sync.Once
	disposeWait sync.WaitGroup
}

type sessgroup struct {
	sync.RWMutex
	sessmap  map[uint64]*Session
	disposed bool
}

func NewManager() *Manager {
	manager := &Manager{}
	for i := 0; i < len(manager.sessgroups); i++ {
		manager.sessgroups[i].sessmap = make(map[uint64]*Session)
	}
	return manager
}

func (manager *Manager) Dispose() {
	manager.disposeOnce.Do(func() {
		for i := 0; i < sessionGroupNum; i++ {
			sessg := &manager.sessgroups[i]
			sessg.Lock()
			sessg.disposed = true
			for _, session := range sessg.sessmap {
				session.Close()
			}
			sessg.Unlock()
		}
		manager.disposeWait.Wait()
	})
}

func (manager *Manager) NewSession(codec Codec, sendChanSize int) *Session {
	session := newSession(manager, codec, sendChanSize)
	manager.addSession(session)
	return session
}

func (manager *Manager) GetSession(sessionID uint64) (session *Session) {
	if sessionID%sessionGroupNum == 0 {
		return nil
	}
	sessg := &manager.sessgroups[sessionID%sessionGroupNum]
	sessg.RLock()
	session, _ = sessg.sessmap[sessionID]
	sessg.RUnlock()
	return
}

func (manager *Manager) addSession(session *Session) {
	if session.id%sessionGroupNum == 0 {
		return
	}
	sessg := &manager.sessgroups[session.id%sessionGroupNum]

	if sessg.disposed {
		session.Close()
		return
	}

	sessg.Lock()
	sessg.sessmap[session.id] = session
	sessg.Unlock()
	manager.disposeWait.Add(1)
}

func (manager *Manager) delSession(session *Session) {
	if session.id%sessionGroupNum == 0 {
		return
	}
	sessg := &manager.sessgroups[session.id%sessionGroupNum]

	sessg.Lock()
	delete(sessg.sessmap, session.id)
	sessg.Unlock()
	manager.disposeWait.Done()
}
