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
	"sync/atomic"
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

const OutputFormat = "Send Count: %d, Recv Count: %d, Read Count: %d, Write Count: %d\n"

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
		conns      = make([]*CountConn, 0, *clientNum)
	)

	link.DefaultConfig.InBufferSize = 1024
	link.DefaultConfig.OutBufferSize = 1024
	link.DefaultConfig.SendChanSize = *asyncChan

	for i := 0; i < *clientNum; i++ {
		conn, err := net.DialTimeout("tcp", *serverAddr, time.Second*3)
		if err != nil {
			panic(err)
		}
		initWait.Add(2)
		countConn := &CountConn{Conn: conn}
		conns = append(conns, countConn)
		go client(initWait, countConn, startChan, resultChan, timeout, msg, pool)
	}

	initWait.Wait()
	close(startChan)

	time.Sleep(time.Second * time.Duration(*runTime))
	var sum CountConn
	for i := 0; i < *clientNum; i++ {
		conn := conns[i]
		conn.Close()
		sum.SendCount += conn.SendCount
		sum.RecvCount += conn.RecvCount
		sum.ReadCount += conn.ReadCount
		sum.WriteCount += conn.WriteCount
	}

	fmt.Printf(OutputFormat, sum.SendCount, sum.RecvCount, sum.ReadCount, sum.WriteCount)
}

type CountConn struct {
	net.Conn
	SendCount  uint32
	RecvCount  uint32
	ReadCount  uint32
	WriteCount uint32
}

func (conn *CountConn) Read(p []byte) (n int, err error) {
	n, err = conn.Conn.Read(p)
	atomic.AddUint32(&conn.ReadCount, 1)
	return
}

func (conn *CountConn) Write(p []byte) (n int, err error) {
	n, err = conn.Conn.Write(p)
	atomic.AddUint32(&conn.WriteCount, 1)
	return
}

func client(initWait *sync.WaitGroup, conn *CountConn, startChan chan int, resultChan chan ClientResult, timeout time.Time, msg link.Message, pool *link.MemPool) {
	client, _ := link.NewSession(0, conn, fixhead.Uint16BE, pool, link.DefaultConfig)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		initWait.Done()
		<-startChan

		for {
			var err error
			if *asyncChan == 0 {
				err = client.Send(msg)
			} else {
				client.AsyncSend(msg)
			}
			if err != nil {
				if timeout.After(time.Now()) {
					println("send:", err.Error())
				}
				break
			}
			atomic.AddUint32(&conn.SendCount, 1)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		initWait.Done()
		<-startChan

		for {
			err := client.ProcessOnce(func(*link.Buffer) error {
				return nil
			})
			if err != nil {
				if timeout.After(time.Now()) {
					println("receive:", err.Error())
				}
				break
			}
			atomic.AddUint32(&conn.RecvCount, 1)
		}
	}()

	wg.Wait()
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

	var sum CountConn
	for i := 0; i < *proces; i++ {
		output := cmds[i].Stdout.String()
		fmt.Print(output)

		var conn CountConn
		fmt.Sscanf(output, OutputFormat,
			&conn.SendCount,
			&conn.RecvCount,
			&conn.ReadCount,
			&conn.WriteCount,
		)

		sum.SendCount += conn.SendCount
		sum.RecvCount += conn.RecvCount
		sum.ReadCount += conn.ReadCount
		sum.WriteCount += conn.WriteCount
	}

	fmt.Println("--------------------")
	fmt.Printf(OutputFormat,
		sum.SendCount, sum.RecvCount, sum.ReadCount, sum.WriteCount,
	)

	return true
}
