package gateway

import (
	"bytes"
	"math/rand"
	"runtime/pprof"
	"sync"
	"testing"
	"time"

	"github.com/funny/binary"
	"github.com/funny/link"
	"github.com/funny/link/packet"
	"github.com/funny/link/stream"
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

func StartTestBackend(t *testing.T, handler func(*link.Session)) *link.Server {
	backend, err := link.Listen("tcp", "0.0.0.0:0", NewBackend(
		1024, 1024, 1024,
	))
	unitest.NotError(t, err)
	go backend.Serve(handler)
	return backend
}

func StartTestGateway(t *testing.T, backendAddr string) *Frontend {
	server, err := link.Listen("tcp", "0.0.0.0:0", packet.New(
		binary.SplitByUint32BE, 1024, 1024, 1024,
	))
	unitest.NotError(t, err)

	var linkIds []uint64

	gateway := NewFrontend(server, func(_ *link.Session) (uint64, error) {
		return linkIds[rand.Intn(len(linkIds))], nil
	})

	for i := 0; i < 10; i++ {
		id, err := gateway.AddBackend("tcp",
			backendAddr,
			stream.New(1024, 1024, 1024),
		)
		unitest.NotError(t, err)
		linkIds = append(linkIds, id)
	}

	return gateway
}

func Test_Gateway_Simple(t *testing.T) {
	backend := StartTestBackend(t, func(session *link.Session) {
		var msg RAW
		for {
			if err := session.Receive(&msg); err != nil {
				break
			}
			if err := session.Send(msg); err != nil {
				break
			}
		}
	})

	gateway := StartTestGateway(t, backend.Listener().Addr().String())
	gatewayAddr := gateway.server.Listener().Addr().String()

	client, err := link.Dial("tcp", gatewayAddr, packet.New(
		binary.SplitByUint32BE, 1024, 1024, 1024,
	))
	unitest.NotError(t, err)
	for i := 0; i < 10000; i++ {
		msg1 := RandBytes(1024)
		err1 := client.Send(packet.RAW(msg1))
		unitest.NotError(t, err1)

		var msg2 packet.RAW
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

func Test_Gateway_Complex(t *testing.T) {
	backend := StartTestBackend(t, func(session *link.Session) {
		var msg RAW
		for {
			if err := session.Receive(&msg); err != nil {
				break
			}
			if err := session.Send(msg); err != nil {
				break
			}
		}
	})

	gateway := StartTestGateway(t, backend.Listener().Addr().String())
	gatewayAddr := gateway.server.Listener().Addr().String()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			client, err := link.Dial("tcp", gatewayAddr, packet.New(
				binary.SplitByUint32BE, 1024, 1024, 1024,
			))
			unitest.NotError(t, err)

			for j := 0; j < 500; j++ {
				msg1 := RandBytes(1024)
				err1 := client.Send(packet.RAW(msg1))
				unitest.NotError(t, err1)

				var msg2 packet.RAW
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
		channel       *link.Channel
		broadcastWait sync.WaitGroup
		clientWait    sync.WaitGroup
	)

	clientWait.Add(clientNum)
	backend := StartTestBackend(t, func(session *link.Session) {
		channel.Join(session, nil)
		clientWait.Done()
		for {
			var msg RAW
			if err := session.Receive(&msg); err != nil {
				break
			}
			broadcastWait.Done()
		}
	})
	channel = link.NewChannel(backend.Listener().Protocol())

	go func() {
		clientWait.Wait()
		for i := 0; i < 10000; i++ {
			msg := RandBytes(10)
			channel.Broadcast(RAW(msg))
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

			client, err := link.Dial("tcp", gatewayAddr, packet.New(
				binary.SplitByUvarint, 1024, 1024, 1024,
			))
			unitest.NotError(t, err)

			for j := 0; j < 10000; j++ {
				var msg packet.RAW
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
