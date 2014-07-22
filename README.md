Introduction
============

This is a network library for Go.

It focus on packet based persistent connection communication.

It provide a packet splitting protocol like Erlang's `{packet: N}`. And supported custom packet splitting protocol.

But it didn't limit the encode or decode format of the request and response.

Also this library provide session management and broadcast features to make your life easy.

!!! THIS PROJECT IS STILL IN EARLY STAGE !!!

!!! API MAY RADICALLY CHANGE IN THE NEAR FUTURE !!!

How to install
==============

```
go get github.com/funny/packnet
```

How to use
===========

Choose a protocol for your project.

```go
import (
	"fmt"
	"encoding/binary"
	"github.com/funny/packnet"
)

...

	proto := packnet.NewFixProtocol(4, binary.BigEndian)
```

Setup a server on port `8080` and set protocol.

```go
	server, err := ListenAndServe("tcp", "0.0.0.0:8080", proto)
	if err != nil {
		panic(err)
	}
	server.Start()
```

Hook the server's session start event to handle incoming connections.

And setup a request handler on the new session to handle incoming messages.

```go
	server.SetSessionStartHook(func(session *Session) {
		fmt.Println("new session in")

		session.SetRequestHandlerFunc(func(session *Session, msg []byte) {
			fmt.Printf("new message: %s\n", msg)
		})
	})
```

On client side. Implement a message type for test.

```go

type TestMessage struct {
	Message string
}

func (msg *TestMessage) RecommendPacketSize() uint {
	return uint(len(msg.Message))
}

func (msg *TestMessage) AppendToPacket(packet []byte) []byte {
	return append(packet, msg.Message...)
}
```

Then use same protocol dial to the server.

```go
	proto := packnet.NewFixProtocol(4, binary.BigEndian)

	client, err := Dial("tcp", "127.0.0.1:8080", proto, 1, 1024)
	if err != nil {
		panic(err)
	}
	client.Start()
```

Send a message to server.

```go
	message := &TestMessage{ "Hello World!" }

	if err2 := client.Send(message); err2 != nil {
		panic(err2)
	}
```

Example 1 - echo
=================

The echo server.

```go
package main

import "sync"
import "encoding/binary"
import "github.com/funny/packnet"

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(2)

	protocol := packnet.NewFixProtocol(4, binary.BigEndian)

	server, err := packnet.ListenAndServe("tcp", "127.0.0.1:10010", protocol)
	if err != nil {
		panic(err)
	}

	server.SetSessionStartHook(func(session *packnet.Session) {
		println("client from: ", session.RawConn().RemoteAddr().String())

		session.SetMessageHandlerFunc(func(session *packnet.Session, message []byte) {
			println("message:", string(message))

			session.Send(EchoMessage{message})
		})

		wg.Done()
	})

	server.SetSessionCloseHook(func(session *packnet.Session) {
		wg.Done()
	})

	server.Start()

	println("server start")

	wg.Wait()

	println("bye")
}

type EchoMessage struct {
	Content []byte
}

func (msg EchoMessage) RecommendPacketSize() uint {
	return uint(len(msg.Content))
}

func (msg EchoMessage) AppendToPacket(packet []byte) []byte {
	return append(packet, msg.Content...)
}
```

The echo client.

```go
package main

import "fmt"
import "encoding/binary"
import "github.com/funny/packnet"

func main() {
	protocol := packnet.NewFixProtocol(4, binary.BigEndian)

	client, err := packnet.Dial("tcp", "127.0.0.1:10010", protocol, 1, 1024)
	if err != nil {
		panic(err)
	}

	client.SetMessageHandlerFunc(func(session *packnet.Session, message []byte) {
		println("message:", string(message))
	})

	client.SetCloseCallback(func(session *packnet.Session) {
		println("closed")
	})

	client.Start()

	for {
		var input string
		if _, err := fmt.Scanf("%s\n", &input); err != nil {
			break
		}
		client.Send(EchoMessage{input})
	}

	client.Close()

	println("bye")
}

type EchoMessage struct {
	Content string
}

func (msg EchoMessage) RecommendPacketSize() uint {
	return uint(len(msg.Content))
}

func (msg EchoMessage) AppendToPacket(packet []byte) []byte {
	return append(packet, msg.Content...)
}
```