package gateway

import (
	"bytes"
	"math/rand"
	"runtime/pprof"
	"sync"
	"testing"
	"time"

	"github.com/funny/link"
	"github.com/funny/unitest"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandBytes(n int) []byte {
	n = rand.Intn(n)
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(rand.Intn(255))
	}
	return b
}

func StartEchoBackend() (*link.Server, error) {
	backend, err := link.Serve("tcp://0.0.0.0:0", NewBackend(), link.Raw())
	if err != nil {
		return nil, err
	}
	go backend.Loop(func(session *link.Session) {
		var msg []byte
		for {
			if err := session.Receive(&msg); err != nil {
				break
			}
			if err := session.Send(msg); err != nil {
				break
			}
		}
	})
	return backend, nil
}

func StartTestGateway(t *testing.T, backendAddr string) *Frontend {
	listener, err := link.ListenPacket("tcp://0.0.0.0:0", link.Packet(link.Uint16BE))
	unitest.NotError(t, err)

	var linkIds []uint64

	gateway := NewFrontend(listener, func(_ *link.Session) (uint64, error) {
		return linkIds[rand.Intn(len(linkIds))], nil
	})

	for i := 0; i < 1; i++ {
		id, err := gateway.AddBackend("tcp://"+backendAddr, link.Stream())
		unitest.NotError(t, err)
		linkIds = append(linkIds, id)
	}

	return gateway
}

func Test_Simple(t *testing.T) {
	backend, err := StartEchoBackend()
	unitest.NotError(t, err)

	gateway := StartTestGateway(t, backend.Listener().Addr().String())
	gatewayAddr := gateway.server.Listener().Addr().String()

	client, err := link.Connect("tcp://"+gatewayAddr, link.Packet(link.Uint16BE), link.Raw())
	unitest.NotError(t, err)

	for i := 0; i < 10000; i++ {
		msg1 := RandBytes(1024)
		err1 := client.Send(msg1)
		unitest.NotError(t, err1)

		var msg2 []byte
		err2 := client.Receive(&msg2)
		unitest.NotError(t, err2)

		if bytes.Equal(msg1, msg2) == false {
			t.Log(i, msg1, msg2)
			t.Fail()
		}
		unitest.Pass(t, bytes.Equal(msg1, msg2))
	}
	client.Close()

	gateway.Stop()
	backend.Stop()

	time.Sleep(time.Second * 2)
	MakeSureSessionGoroutineExit(t)
}

func Test_Complex(t *testing.T) {
	backend, err := StartEchoBackend()
	unitest.NotError(t, err)

	gateway := StartTestGateway(t, backend.Listener().Addr().String())
	gatewayAddr := gateway.server.Listener().Addr().String()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			client, err := link.Connect("tcp://"+gatewayAddr, link.Packet(link.Uint16BE), link.Raw())
			unitest.NotError(t, err)

			for j := 0; j < 500; j++ {
				msg1 := RandBytes(1024)
				err1 := client.Send(msg1)
				unitest.NotError(t, err1)

				var msg2 []byte
				err2 := client.Receive(&msg2)
				unitest.NotError(t, err2)

				if bytes.Equal(msg1, msg2) == false {
					t.Log(j, msg1, msg2)
					t.Fail()
				}
				unitest.Pass(t, bytes.Equal(msg1, msg2))
			}

			client.Close()
		}()
	}
	wg.Wait()

	gateway.Stop()
	backend.Stop()

	time.Sleep(time.Second * 2)
	MakeSureSessionGoroutineExit(t)
}

func Test_Broadcast(t *testing.T) {
	var (
		clientNum     = 20
		packetNum     = 2000
		channel       = NewChannel(link.Raw())
		broadcastMsg  []byte
		broadcastWait sync.WaitGroup
		clientWait    sync.WaitGroup
	)

	backend, err := link.Serve("tcp://0.0.0.0:0", NewBackend(), link.Raw())
	unitest.NotError(t, err)

	go backend.Loop(func(session *link.Session) {
		channel.Join(session)
		clientWait.Done()
		for {
			var msg []byte
			if err := session.Receive(&msg); err != nil {
				break
			}
			unitest.Pass(t, bytes.Equal(msg, broadcastMsg))
			broadcastWait.Done()
		}
	})

	clientWait.Add(clientNum)
	go func() {
		clientWait.Wait()
		for i := 0; i < packetNum; i++ {
			broadcastMsg = RandBytes(1024)
			channel.Broadcast(broadcastMsg)
			broadcastWait.Add(clientNum)
			broadcastWait.Wait()
		}
	}()

	gateway := StartTestGateway(t, backend.Listener().Addr().String())
	gatewayAddr := gateway.server.Listener().Addr().String()

	var wg sync.WaitGroup
	for i := 0; i < clientNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			client, err := link.Connect("tcp://"+gatewayAddr, link.Packet(link.Uint16BE), link.Raw())
			unitest.NotError(t, err)

			for j := 0; j < packetNum; j++ {
				var msg []byte
				err := client.Receive(&msg)
				unitest.NotError(t, err)

				err = client.Send(msg)
				unitest.NotError(t, err)
			}

			client.Close()
		}()
	}
	wg.Wait()

	gateway.Stop()
	backend.Stop()

	time.Sleep(time.Second * 2)
	MakeSureSessionGoroutineExit(t)
}

func MakeSureSessionGoroutineExit(t *testing.T) {
	buff := new(bytes.Buffer)
	goroutines := pprof.Lookup("goroutine")

	if err := goroutines.WriteTo(buff, 2); err != nil {
		t.Fatalf("Dump goroutine failed: %v", err)
	}

	if n := bytes.Index(buff.Bytes(), []byte("asyncSendLoop")); n >= 0 {
		t.Log(buff.String())
		t.Fatalf("Some session goroutine running")
	}
}
