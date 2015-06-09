package gateway

import (
	"net"

	"github.com/funny/link"
	"github.com/funny/link/stream"
)

/*
Example:

	import (
		"github.com/funny/link"
		"github.com/funny/link/gateway"
	)

	server, _ := link.Listen("tcp", "0.0.0.0:0", gateway.NewBackend(1024,1024,1024))
*/
type Backend struct {
	protocol *stream.Protocol
}

func NewBackend(readBufferSize, writeBufferSize, sendChanSize int) *Backend {
	return &Backend{stream.New(readBufferSize, writeBufferSize, sendChanSize)}
}

func (backend *Backend) NewListener(listener net.Listener) link.Listener {
	return NewBackendListener(link.NewServer(backend.protocol.NewListener(listener)))
}

type Broadcast struct {
}

func (_ Broadcast) Broadcast(msg interface{}, fetcher link.SessionFetcher) error {
	var server *BackendListener
	ids := make([]uint64, 0, 10)
	fetcher(func(session *link.Session) {
		conn := session.Conn().(*BackendConn)
		ids = append(ids, conn.id)
		if server == nil {
			server = conn.link.listener
		}
	})
	server.broadcast(ids, msg)
	return nil
}
