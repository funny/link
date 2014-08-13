package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/funny/link"
	"sync"
	"time"
)

var (
	serverAddr  = flag.String("addr", "", "target server address")
	clientNum   = flag.Int("num", 1, "client number")
	messageSize = flag.Int("size", 64, "test message size")
	runTime     = flag.Int("time", 10, "benchmark run time in seconds")
)

// This is an benchmark tool work with the echo_server.
//
// Start echo_server with 'bench' flag
//     go run github.com/funny/examples/echo_server/main.go -bench
//
// Start benchmark with echo_server address
//     go run github.com/funny/examples/benchmark/main.go -addr="127.0.0.1:10010"
func main() {
	flag.Parse()

	var (
		msg       = TestMessage{make([]byte, *messageSize)}
		timeout   = time.Now().Add(time.Second * time.Duration(*runTime))
		initWait  = new(sync.WaitGroup)
		startChan = make(chan int)
		countChan = make(chan int)
	)

	for i := 0; i < *clientNum; i++ {
		initWait.Add(1)
		go client(initWait, startChan, countChan, timeout, msg)
	}

	initWait.Wait()
	close(startChan)

	count := 0
	for i := 0; i < *clientNum; i++ {
		count += <-countChan
	}

	fmt.Printf("Total Num: %d, Total Size: %d\n", count, count*(*messageSize))
}

func client(initWait *sync.WaitGroup, startChan, countChan chan int, timeout time.Time, msg TestMessage) {
	client, err := link.Dial("tcp", *serverAddr, link.NewFixProtocol(2, binary.BigEndian))
	if err != nil {
		panic(err)
	}
	defer client.Close(nil)

	count := 0
	client.OnClose(func(session *link.Session, reason error) {
		if reason != nil {
			println(reason.Error())
		}
		countChan <- count
	})

	client.Start()

	initWait.Done()
	<-startChan

	for timeout.After(time.Now()) {
		if err := client.SyncSend(msg); err != nil {
			break
		}
		count += 1
	}
}

type TestMessage struct {
	Content []byte
}

func (msg TestMessage) RecommendPacketSize() uint {
	return uint(len(msg.Content))
}

func (msg TestMessage) AppendToPacket(packet []byte) []byte {
	return append(packet, msg.Content...)
}
