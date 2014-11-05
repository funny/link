package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/funny/link"
	"net"
	"sync"
	"time"
)

var (
	serverAddr  = flag.String("addr", "", "target server address")
	clientNum   = flag.Int("num", 1, "client number")
	messageSize = flag.Int("size", 64, "test message size")
	runTime     = flag.Int("time", 10, "benchmark run time in seconds")
	bufferSize  = flag.Int("buffer", 1024, "read buffer size")
)

type ClientResult struct {
	SendCount  int
	ReadCount  int
	WriteCount int
}

// This is an benchmark tool work with the echo_server.
//
// Start echo_server with 'bench' flag
//     go run echo_server/main.go -bench
//
// Start benchmark with echo_server address
//     go run benchmark/main.go -addr="127.0.0.1:10010"
func main() {
	flag.Parse()

	var (
		msg        = make(link.Binary, *messageSize)
		timeout    = time.Now().Add(time.Second * time.Duration(*runTime))
		initWait   = new(sync.WaitGroup)
		startChan  = make(chan int)
		resultChan = make(chan ClientResult)
	)

	for i := 0; i < *clientNum; i++ {
		initWait.Add(1)
		go client(initWait, startChan, resultChan, timeout, msg)
	}

	initWait.Wait()
	close(startChan)

	var (
		sendCount  = 0
		readCount  = 0
		writeCount = 0
	)
	for i := 0; i < *clientNum; i++ {
		result := <-resultChan
		sendCount += result.SendCount
		readCount += result.ReadCount
		writeCount += result.WriteCount
	}

	fmt.Printf(
		"Send Count: %d, Total Size: %d, Read Count: %d, Write Count: %d\n",
		sendCount, sendCount*(*messageSize), readCount, writeCount,
	)
}

type CountConn struct {
	net.Conn
	ReadCount  int
	WriteCount int
}

func (conn *CountConn) Read(p []byte) (int, error) {
	conn.ReadCount += 1
	return conn.Conn.Read(p)
}

func (conn *CountConn) Write(p []byte) (int, error) {
	conn.WriteCount += 1
	return conn.Conn.Write(p)
}

func client(initWait *sync.WaitGroup, startChan chan int, resultChan chan ClientResult, timeout time.Time, msg link.Binary) {
	conn, err := net.DialTimeout("tcp", *serverAddr, time.Second*3)
	if err != nil {
		panic(err)
	}

	conn = &CountConn{conn, 0, 0}
	client := link.NewSession(0, conn, link.PacketN(2, binary.BigEndian), link.DefaultSendChanSize, *bufferSize)
	defer client.Close(nil)

	go func() {
		initWait.Done()
		for {
			if _, err := client.Read(); err != nil {
				client.Close(err)
				break
			}
		}
	}()

	<-startChan

	count := 0
	for timeout.After(time.Now()) {
		if err := client.Send(msg); err != nil {
			break
		}
		count += 1
	}

	if client.CloseReason() != nil {
		println(client.CloseReason().(error).Error())
	}

	resultChan <- ClientResult{
		count,
		conn.(*CountConn).ReadCount,
		conn.(*CountConn).WriteCount,
	}
}
