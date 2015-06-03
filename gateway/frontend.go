package gateway

import (
	"sync"
	"time"

	"github.com/funny/link"
	"github.com/funny/link/stream"
)

type Frontend struct {
	server     *link.Server
	maxLinkId  uint64
	links      map[uint64]*frontendLink
	linksMutex sync.RWMutex
	handshaker func(*link.Session) (uint64, error)
}

func NewFrontend(server *link.Server, handshaker func(*link.Session) (uint64, error)) *Frontend {
	front := &Frontend{
		server:     server,
		links:      make(map[uint64]*frontendLink),
		handshaker: handshaker,
	}
	go front.serveClient()
	return front
}

func (front *Frontend) serveClient() {
	front.server.Serve(func(session *link.Session) {
		linkId, err := front.handshaker(session)
		if err != nil {
			return
		}

		front.linksMutex.RLock()
		defer front.linksMutex.RUnlock()

		if flink, exists := front.links[linkId]; exists {
			flink.AddClient(session)
		}
	})

	front.linksMutex.Lock()
	defer front.linksMutex.Unlock()

	for _, flink := range front.links {
		flink.Close()
	}
}

func (front *Frontend) Stop() {
	front.server.Stop()
}

func (front *Frontend) AddBackend(network, address string, protocol *stream.Protocol) (uint64, error) {
	session, err := link.DialTimeout(network, address, time.Second*3, protocol)
	if err != nil {
		return 0, err
	}

	front.linksMutex.Lock()
	defer front.linksMutex.Unlock()

	front.maxLinkId += 1
	front.links[front.maxLinkId] = newFrontLink(session)
	return front.maxLinkId, nil
}
