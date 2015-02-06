package rpc

import (
	"encoding/json"
	"errors"
	"github.com/funny/link"
	"strings"
	"sync"
	"sync/atomic"
)

type Client struct {
	session *link.Session
	seqNum  uint32
	mutex   sync.Mutex
	request map[uint32]*rpcRequestState
}

type rpcRequestState struct {
	Reply interface{}
	C     chan error
}

func Dial(network, address string) (*Client, error) {
	session, err := link.Dial(network, address)
	if err != nil {
		return nil, err
	}
	client := &Client{
		session: session,
		request: make(map[uint32]*rpcRequestState),
	}
	go session.Process(func(msg *link.InBuffer) error {
		seqNum := msg.ReadUint32LE()

		client.mutex.Lock()
		state := client.request[seqNum]
		delete(client.request, seqNum)
		client.mutex.Unlock()

		if err := msg.ReadString(int(msg.ReadUint32LE())); err != "" {
			state.C <- errors.New(err)
			return nil
		}

		if err := json.NewDecoder(msg).Decode(state.Reply); err != nil {
			state.C <- err
			return nil
		}

		state.C <- nil
		return nil
	})
	return client, nil
}

func (client *Client) Call(serviceMethod string, args, reply interface{}) error {
	var (
		names  = strings.Split(serviceMethod, ".")
		seqNum = atomic.AddUint32(&client.seqNum, 1)
		c      = make(chan error, 1)
	)

	client.mutex.Lock()
	client.request[seqNum] = &rpcRequestState{
		Reply: reply,
		C:     c,
	}
	client.mutex.Unlock()

	client.session.AsyncSend(link.MessageFunc(func(buffer *link.OutBuffer) error {
		buffer.WriteUvarint(uint64(len(names[0])))
		buffer.WriteString(names[0])
		buffer.WriteUvarint(uint64(len(names[1])))
		buffer.WriteString(names[1])
		buffer.WriteUint32LE(seqNum)
		return json.NewEncoder(buffer).Encode(args)
	}), 0)

	return <-c
}

func (client *Client) Close() {
	client.session.Close()
}

type rpcRequest struct {
	SeqNum  uint32
	Service string
	Method  string
	Args    interface{}
}

func (req *rpcRequest) RecommendBufferSize() int {
	return 1024
}
