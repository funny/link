package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/funny/link"
	_ "github.com/funny/utest"
)

var (
	serverAddr  = flag.String("addr", "127.0.0.1:10010", "echo server address")
	clientNum   = flag.Int("num", 1, "client number")
	messageSize = flag.Int("size", 64, "test message size")
	runTime     = flag.Int("time", 10, "benchmark run time in seconds")
	proces      = flag.Int("procs", 1, "how many benchmark process")
	randsize    = flag.Bool("rand", false, "random message size")
	waitMaster  = flag.Bool("wait", false, "DO NOT USE")
)

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

type TestCodec struct {
}

func (codec TestCodec) NewEncoder(w io.Writer) link.Encoder {
	return &TestEncoder{w}
}

func (codec TestCodec) NewDecoder(r io.Reader) link.Decoder {
	return &TestDecoder{r}
}

type TestEncoder struct {
	w io.Writer
}

func (encoder *TestEncoder) Encode(msg interface{}) error {
	_, err := encoder.w.Write(msg.([]byte))
	return err
}

type TestDecoder struct {
	r io.Reader
}

func (decoder *TestDecoder) Decode(msg interface{}) error {
	d := make([]byte, decoder.r.(*io.LimitedReader).N)
	_, err := io.ReadFull(decoder.r, d)
	if err != nil {
		return err
	}
	*(msg.(*[]byte)) = d
	return nil
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
		msg       = make([]byte, *messageSize)
		timeout   = time.Now().Add(time.Second * time.Duration(*runTime))
		initWait  = new(sync.WaitGroup)
		startChan = make(chan int)
		conns     = make([]*CountConn, 0, *clientNum)
	)

	for i := 0; i < *clientNum; i++ {
		conn, err := net.DialTimeout("tcp", *serverAddr, time.Second*3)
		if err != nil {
			panic(err)
		}

		countConn := &CountConn{Conn: conn}
		conns = append(conns, countConn)

		initWait.Add(2)
		go client(initWait, countConn, startChan, timeout, msg)
	}
	initWait.Wait()
	close(startChan)

	time.Sleep(time.Second * time.Duration(*runTime))
	var sum CountConn
	for i := 0; i < *clientNum; i++ {
		conn := conns[i]
		conn.Conn.Close()
		sum.SendCount += conn.SendCount
		sum.RecvCount += conn.RecvCount
		sum.ReadCount += conn.ReadCount
		sum.WriteCount += conn.WriteCount
	}
	fmt.Printf(OutputFormat, sum.SendCount, sum.RecvCount, sum.ReadCount, sum.WriteCount)
}

func client(initWait *sync.WaitGroup, conn *CountConn, startChan chan int, timeout time.Time, msg []byte) {
	client := link.NewSession(conn, link.Packet(2, *messageSize, 4096, binary.LittleEndian, TestCodec{}))

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		initWait.Done()
		<-startChan

		for {
			outMsg := msg
			if *randsize {
				outMsg = msg[:rand.Intn(*messageSize)]
			}
			if err := client.Send(outMsg); err != nil {
				if timeout.After(time.Now()) {
					println("send error:", err.Error())
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

		var inMsg []byte
		for {
			if err := client.Receive(&inMsg); err != nil {
				if timeout.After(time.Now()) {
					println("recv error:", err.Error())
				}
				break
			}
			atomic.AddUint32(&conn.RecvCount, 1)
		}
	}()

	wg.Wait()
}

type childProc struct {
	Cmd    *exec.Cmd
	Stdin  io.WriteCloser
	Stdout bytes.Buffer
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
			"go", "run", "echo_benchmark.go", "wait",
			"addr="+*serverAddr,
			"num="+strconv.Itoa(*clientNum / *proces),
			"size="+strconv.Itoa(*messageSize),
			"time="+strconv.Itoa(*runTime),
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

		var c CountConn
		fmt.Sscanf(output, OutputFormat,
			&c.SendCount,
			&c.RecvCount,
			&c.ReadCount,
			&c.WriteCount,
		)

		sum.SendCount += c.SendCount
		sum.RecvCount += c.RecvCount
		sum.ReadCount += c.ReadCount
		sum.WriteCount += c.WriteCount
	}

	fmt.Println("--------------------")
	fmt.Printf(OutputFormat, sum.SendCount, sum.RecvCount, sum.ReadCount, sum.WriteCount)
	return true
}
