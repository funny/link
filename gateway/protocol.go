package gateway

import (
	"net"

	"github.com/funny/link"
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
	*link.StreamProtocol
}

func NewBackend() *Backend {
	return &Backend{link.Stream()}
}

func (backend *Backend) NewListener(listener net.Listener) (link.Listener, error) {
	lsn, _ := backend.StreamProtocol.NewListener(listener)
	srv := link.NewServer(lsn, link.SelfCodec())
	return NewBackendListener(srv), nil
}

type broadcast struct {
	Codec link.PacketCodec
}

func NewChannel(codecType link.PacketCodecType) *link.Channel {
	return link.NewCustomChannel(broadcast{codecType.NewPacketCodec()})
}

func (b broadcast) Broadcast(msg interface{}, fetcher link.SessionFetcher) error {
	data, err := b.Codec.EncodePacket(msg)
	if err != nil {
		return err
	}

	var server *BackendListener
	ids := make([]uint64, 0, 10)
	fetcher(func(session *link.Session) {
		conn := session.Conn().(*BackendConn)
		ids = append(ids, conn.id)
		if server == nil {
			server = conn.link.listener
		}
	})

	server.broadcast(ids, data)
	return nil
}
