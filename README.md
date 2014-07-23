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
	server.OnSessionStart(func(session *Session) {
		fmt.Println("new session in")

		session.OnMessage(func(session *Session, msg []byte) {
			fmt.Printf("new message: %s\n", msg)
		})
	})
```

On client side. Implement a message type for test.

```go

type TestMessage struct {
	Message string
}

func (msg TestMessage) RecommendPacketSize() uint {
	return uint(len(msg.Message))
}

func (msg TestMessage) AppendToPacket(packet []byte) []byte {
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
	message := TestMessage{ "Hello World!" }

	if err2 := client.Send(message); err2 != nil {
		panic(err2)
	}
```

Example 1 - echo
=================

The echo server.

```go
package main

import "encoding/binary"
import "github.com/funny/packnet"

// This is an echo server demo work with the echo_client.
// usage:
//     go run github.com/funny/examples/echo_server/main.go
func main() {
	protocol := packnet.NewFixProtocol(4, binary.BigEndian)

	server, err := packnet.ListenAndServe("tcp", "127.0.0.1:10010", protocol)
	if err != nil {
		panic(err)
	}

	server.OnSessionStart(func(session *packnet.Session) {
		println("client", session.RawConn().RemoteAddr().String(), "in")

		session.OnMessage(func(session *packnet.Session, message []byte) {
			println("client", session.RawConn().RemoteAddr().String(), "say:", string(message))

			session.Send(EchoMessage{message})
		})
	})

	server.OnSessionClose(func(session *packnet.Session) {
		println("client", session.RawConn().RemoteAddr().String(), "close")
	})

	server.Start()

	println("server start")

	<-make(chan int)
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

// This is an echo client demo work with the echo_server.
// usage:
//     go run github.com/funny/examples/echo_client/main.go
func main() {
	protocol := packnet.NewFixProtocol(4, binary.BigEndian)

	client, err := packnet.Dial("tcp", "127.0.0.1:10010", protocol, 1, 1024)
	if err != nil {
		panic(err)
	}

	client.OnMessage(func(session *packnet.Session, message []byte) {
		println("message:", string(message))
	})

	client.OnClose(func(session *packnet.Session) {
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

Example 2 - broadcast
=====================

The broadcast server. Use the echo client to receive broadcast.

```go
package main

import "time"
import "encoding/binary"
import "github.com/funny/packnet"

// This is broadcast server demo work with the echo_client.
// usage:
//     go run github.com/funny/examples/broadcast/main.go
func main() {
	protocol := packnet.NewFixProtocol(4, binary.BigEndian)

	server, err := packnet.ListenAndServe("tcp", "127.0.0.1:10010", protocol)
	if err != nil {
		panic(err)
	}

	channel := server.NewChannel()

	server.OnSessionStart(func(session *packnet.Session) {
		println("client", session.RawConn().RemoteAddr().String(), "in")

		channel.Join(session, nil)
	})

	server.OnSessionClose(func(session *packnet.Session) {
		println("client", session.RawConn().RemoteAddr().String(), "close")

		channel.Exit(session)
	})

	server.Start()

	go func() {
		for {
			time.Sleep(time.Second)

			channel.Broadcast(EchoMessage{time.Now().String()})
		}
	}()

	println("server start")

	<-make(chan int)
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
