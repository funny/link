package rpc

import (
	"github.com/funny/jsonrpc"
	"github.com/funny/unitest"
	"net"
	"sync"
	"testing"
)

type Arith int

type Args struct {
	A, B int
}

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func Test_RPC(t *testing.T) {
	server, err := NewServer("tcp", "0.0.0.0:12345")
	unitest.NotError(t, err)

	err2 := server.Register(new(Arith))
	unitest.NotError(t, err2)

	go server.Serve()

	client, err := Dial("tcp", "127.0.0.1:12345")
	unitest.NotError(t, err)

	var reply int
	err3 := client.Call("Arith.Multiply", &Args{7, 8}, &reply)
	unitest.NotError(t, err3)

	unitest.Pass(t, reply == 56)
}

func Benchmark_1(b *testing.B) {
	b.StopTimer()
	var server = jsonrpc.NewServer()
	server.Register(new(Arith))
	var wg sync.WaitGroup
	wg.Add(1)
	address := ""
	go func() {
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			b.Log("Server TCP:", err)
			return
		}
		address = lis.Addr().String()
		wg.Done()
		server.Accept(lis)
	}()
	wg.Wait()
	var client, err = jsonrpc.Dial("tcp", address)
	if err != nil {
		b.Fatal("Dial:", err)
	}
	b.StartTimer()

	var reply int
	for i := 0; i < b.N; i++ {
		client.Call("Arith.Multiply", &Args{7, 8}, &reply)
	}
}

func Benchmark_2(b *testing.B) {
	b.StopTimer()
	server, err := NewServer("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal("NewServer:", err)
	}
	server.Register(new(Arith))
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		server.Serve()
	}()
	wg.Wait()
	client, err := Dial("tcp", server.server.Listener().Addr().String())
	if err != nil {
		b.Fatal("Dial:", err)
	}
	b.StartTimer()

	var reply int
	for i := 0; i < b.N; i++ {
		client.Call("Arith.Multiply", &Args{7, 8}, &reply)
	}
}
