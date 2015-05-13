package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/funny/link"
	"github.com/funny/link/protocol/fixhead"
	"io"
	"net"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

var (
	serverAddr  = flag.String("addr", "127.0.0.1:10010", "echo server address")
	clientNum   = flag.Int("num", 1, "client number")
	messageSize = flag.Int("size", 64, "test message size")
	runTime     = flag.Int("time", 10, "benchmark run time in seconds")
	asyncChan   = flag.Int("async", 10000, "async send chan size, 0 == sync send")
	proces      = flag.Int("procs", 1, "how many benchmark process")
	waitMaster  = flag.Bool("wait", false, "DO NOT USE")
)

type ClientResult struct {
	SendCount  int
	ReadCount  int
	WriteCount int
}

type childProc struct {
	Cmd    *exec.Cmd
	Stdin  io.WriteCloser
	Stdout bytes.Buffer
}

const OutputFormat = "Send Count: %d, Total Size: %d, Read Count: %d, Write Count: %d\n"

// This is an benchmark tool work with the echo_server.
//
// Start echo_server with 'bench' flag
//     go run echo_server.go -bench
//
// Start benchmark with echo_server address
//     go run echo_benchmark.go
//     go run echo_benchmark.go -num=100
//     go run echo_benchmark.go -size=1024
//     go run echo_benchmark.go -time=20
//     go run echo_benchmark.go -addr="127.0.0.1:10010"
func main() {
	flag.Parse()

	if MultiProcess() {
		return
	}

	var (
		msg        = link.Bytes(make([]byte, *messageSize))
		timeout    = time.Now().Add(time.Second * time.Duration(*runTime))
		initWait   = new(sync.WaitGroup)
		startChan  = make(chan int)
		resultChan = make(chan ClientResult)
		pool       = link.NewMemPool(10, 1, 10)
	)

	link.DefaultConfig.RequestBufferSize = 1024
	link.DefaultConfig.ResponseBufferSize = 1024
	link.DefaultConfig.SendChanSize = *asyncChan

	for i := 0; i < *clientNum; i++ {
		initWait.Add(1)
		conn, err := net.DialTimeout("tcp", *serverAddr, time.Second*3)
		if err != nil {
			panic(err)
		}
		go client(initWait, &CountConn{Conn: conn}, startChan, resultChan, timeout, msg, pool)
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

	fmt.Printf(OutputFormat,
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

func client(initWait *sync.WaitGroup, conn *CountConn, startChan chan int, resultChan chan ClientResult, timeout time.Time, msg link.Message, pool *link.MemPool) {
	client, _ := link.NewSession(0, conn, fixhead.Uint16BE, pool, link.DefaultConfig)
	defer client.Close()

	recvTrigger := make(chan int, 1024)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		initWait.Done()
		for {
			if t := <-recvTrigger; t == 0 {
				break
			}
			_, err := client.Receive(link.DecodeFunc(func(*link.Buffer) (link.Request, error) {
				return nil, nil
			}))
			if err != nil {
				println("receive:", err.Error())
				break
			}
		}
	}()

	<-startChan

	sendCount := 0

	for timeout.After(time.Now()) {
		recvTrigger <- 1
		var err error
		if *asyncChan == 0 {
			err = client.Send(msg)
		} else {
			client.AsyncSend(msg)
		}
		if err != nil {
			println("send:", err.Error())
			break
		}
		sendCount += 1
	}
	recvTrigger <- 0
	wg.Wait()

	resultChan <- ClientResult{
		sendCount,
		conn.ReadCount,
		conn.WriteCount,
	}
}

func MultiProcess() bool {
	if *proces <= 1 {
		if *waitMaster {
			fmt.Scanln()
		}
		return false
	}

	cmds := make([]*childProc, *proces)
	for i := 0; i < *proces; i++ {
		cmd := exec.Command(
			"go",
			"run",
			"echo_benchmark.go",
			"addr="+*serverAddr,
			"num="+strconv.Itoa(*clientNum / *proces),
			"size="+strconv.Itoa(*messageSize),
			"time="+strconv.Itoa(*runTime),
			"async="+strconv.Itoa(*asyncChan),
			"wait",
		)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			panic("get stdin pipe failed: " + err.Error())
		}
		cmds[i] = &childProc{Cmd: cmd, Stdin: stdin}
		cmd.Stdout = &cmds[i].Stdout
		cmd.Start()
	}

	for i := 0; i < *proces; i++ {
		cmds[i].Stdin.Write([]byte{'\n'})
	}

	for i := 0; i < *proces; i++ {
		err := cmds[i].Cmd.Wait()
		if err != nil {
			println("wait proc failed:", err.Error())
		}
	}

	var (
		totalSendCount  = 0
		totalTotalSize  = 0
		totalReadCount  = 0
		totalWriteCount = 0
	)

	for i := 0; i < *proces; i++ {
		output := cmds[i].Stdout.String()
		fmt.Print(output)

		var (
			sendCount  = 0
			totalSize  = 0
			readCount  = 0
			writeCount = 0
		)

		fmt.Sscanf(output, OutputFormat,
			&sendCount,
			&totalSize,
			&readCount,
			&writeCount,
		)

		totalSendCount += sendCount
		totalTotalSize += totalSize
		totalReadCount += readCount
		totalWriteCount += writeCount
	}

	fmt.Println("--------------------")
	fmt.Printf(OutputFormat,
		totalSendCount, totalTotalSize, totalReadCount, totalWriteCount,
	)

	return true
}
