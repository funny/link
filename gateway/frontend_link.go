package gateway

import (
	"sync"
	"sync/atomic"

	"github.com/funny/link"
	"github.com/funny/link/packet"
)

type frontendLink struct {
	session      *link.Session
	clients      map[uint64]*link.Session
	clientWaits  map[uint64]*link.Session
	clientWaitId uint64
	clientMutex  sync.RWMutex
	closeFlag    int32
}

func newFrontLink(session *link.Session) *frontendLink {
	flink := &frontendLink{
		session:     session,
		clients:     make(map[uint64]*link.Session),
		clientWaits: make(map[uint64]*link.Session),
	}

	session.AddCloseCallback(flink, func() {
		flink.Close()
	})

	go func() {
		var msg = gatewayMsg{}
		for {
			if err := session.Receive(&msg); err != nil {
				break
			}
			switch msg.Command {
			case CMD_MSG:
				flink.dispathMsg(msg.ClientId, msg.Data)
			case CMD_BRD:
				flink.broadcast(msg.ClientIds, msg.Data)
			case CMD_NEW_2:
				flink.newClient(msg.ClientId, msg.ClientIds[0])
			case CMD_DEL:
				flink.delClient(msg.ClientId, false)
			}
		}
		flink.Close()
	}()

	return flink
}

func (flink *frontendLink) Close() {
	if atomic.CompareAndSwapInt32(&flink.closeFlag, 0, 1) {
		flink.clientMutex.Lock()
		defer flink.clientMutex.Unlock()

		flink.session.RemoveCloseCallback(flink)
		flink.session.Close()

		for _, client := range flink.clients {
			client.Close()
		}
	}
}

func (flink *frontendLink) AddClient(session *link.Session) {
	flink.clientMutex.Lock()
	defer flink.clientMutex.Unlock()

	flink.clientWaitId += 1
	flink.clientWaits[flink.clientWaitId] = session

	addr := session.Conn().RemoteAddr()
	flink.session.Send(&gatewayMsg{
		Command: CMD_NEW_1, ClientId: flink.clientWaitId,
		Data: []byte(addr.Network() + "," + addr.String()),
	})
}

func (flink *frontendLink) newClient(waitId, clientId uint64) {
	flink.clientMutex.Lock()
	defer flink.clientMutex.Unlock()

	session, exists := flink.clientWaits[waitId]
	if !exists {
		return
	}

	delete(flink.clientWaits, waitId)
	flink.clients[clientId] = session

	go func() {
		var (
			inMsg  packet.RAW
			outMsg = gatewayMsg{Command: CMD_NEW_3, ClientId: clientId}
		)

		if err := flink.session.Send(&outMsg); err != nil {
			flink.Close()
			return
		}

		outMsg.Command = CMD_MSG
		for {
			if err := session.Receive(&inMsg); err != nil {
				flink.delClient(clientId, true)
				break
			}
			outMsg.Data = inMsg
			if err := flink.session.Send(&outMsg); err != nil {
				flink.Close()
				break
			}
		}
	}()
}

func (flink *frontendLink) delClient(id uint64, feedback bool) {
	flink.clientMutex.Lock()
	defer flink.clientMutex.Unlock()

	if _, exists := flink.clients[id]; exists {
		delete(flink.clients, id)
		if feedback {
			flink.session.AsyncSend(&gatewayMsg{
				Command: CMD_DEL, ClientId: id,
			})
		}
	}
}

func (flink *frontendLink) dispathMsg(clientId uint64, data []byte) {
	flink.clientMutex.RLock()
	defer flink.clientMutex.RUnlock()

	if client, exists := flink.clients[clientId]; exists {
		client.AsyncSend(packet.RAW(data))
	}
}

func (flink *frontendLink) broadcast(clientIds []uint64, data []byte) {
	flink.clientMutex.RLock()
	defer flink.clientMutex.RUnlock()

	for i := 0; i < len(clientIds); i++ {
		if client, exists := flink.clients[clientIds[i]]; exists {
			client.AsyncSend(packet.RAW(data))
		}
	}
}
