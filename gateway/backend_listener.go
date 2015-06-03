package gateway

import (
	"io"
	"net"
	"sync"

	"github.com/funny/link"
)

type BackendListener struct {
	server     *link.Server
	protocol   *Backend
	acceptChan chan link.Conn
	linkMutex  sync.RWMutex
	maxLinkId  uint64
	links      map[uint64]*backendLink
}

func NewBackendListener(server *link.Server) *BackendListener {
	backend := &BackendListener{
		server:     server,
		links:      make(map[uint64]*backendLink),
		acceptChan: make(chan link.Conn, 2000),
	}
	go backend.serve()
	return backend
}

func (this *BackendListener) serve() {
	this.server.Serve(func(session *link.Session) {
		this.linkMutex.Lock()
		defer this.linkMutex.Unlock()

		this.maxLinkId += 1
		this.links[this.maxLinkId] = newBackendLink(this.maxLinkId, this, session)
	})
}

func (this *BackendListener) delLink(id uint64) {
	this.linkMutex.Lock()
	defer this.linkMutex.Unlock()

	if _, exists := this.links[id]; exists {
		delete(this.links, id)
	}
}

func (this *BackendListener) broadcast(ids []uint64, data []byte) {
	this.linkMutex.RLock()
	defer this.linkMutex.RUnlock()

	for _, link := range this.links {
		link.session.AsyncSend(&gatewayMsg{
			Command: CMD_BRD, ClientIds: ids, Data: data,
		})
	}
}

func (this *BackendListener) Handshake(_ link.Conn) error {
	return nil
}

func (this *BackendListener) Protocol() link.ServerProtocol {
	return this.protocol
}

// link.Listener.Accept()
func (this *BackendListener) Accept() (link.Conn, error) {
	conn, ok := <-this.acceptChan
	if !ok {
		return nil, io.EOF
	}
	return conn, nil
}

// link.Listener.Close()
func (this *BackendListener) Close() error {
	this.server.Stop()

	this.linkMutex.Lock()
	defer this.linkMutex.Unlock()

	for _, link := range this.links {
		link.Close(false)
	}
	return nil
}

// link.Listener.Addr()
func (this *BackendListener) Addr() net.Addr {
	return this.server.Listener().Addr()
}
