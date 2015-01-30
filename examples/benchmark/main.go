package main

import (
	"flag"
	"fmt"
	"github.com/funny/link"
	"github.com/funny/sync"
	"net"
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
		msg        = link.Bytes(make([]byte, *messageSize))
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

func client(initWait *sync.WaitGroup, startChan chan int, resultChan chan ClientResult, timeout time.Time, message link.Message) {
	conn, err := net.DialTimeout("tcp", *serverAddr, time.Second*3)
	if err != nil {
		panic(err)
	}

	conn = &CountConn{conn, 0, 0}
	client, _ := link.NewSession(0, conn, link.DefaultProtocol, link.CLIENT_SIDE, link.DefaultSendChanSize, *bufferSize)
	defer client.Close()

	sendCount := 0
	recvCount := 0
	sendDone := false

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		initWait.Done()
		for {
			err := client.ProcessOnce(func(*link.InBuffer) error {
				return nil
			})
			recvCount += 1
			if sendDone && recvCount == sendCount {
				break
			}
			if err != nil {
				println(err.Error())
				break
			}
		}
	}()

	<-startChan

	for timeout.After(time.Now()) {
		if err := client.Send(message); err != nil {
			break
		}
		sendCount += 1
	}
	sendDone = true
	wg.Wait()

	resultChan <- ClientResult{
		sendCount,
		conn.(*CountConn).ReadCount,
		conn.(*CountConn).WriteCount,
	}
}
