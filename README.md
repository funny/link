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
	proto := packnet.NewFixProtocol(4, binary.BigEndian)
```

Setup a server on port `8080` and set protocol.

```go
	server, _ := ListenAndServe("tcp", "0.0.0.0:8080", proto)

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

Use the same protocol dial to the server.

```go
	proto := packnet.NewFixProtocol(4, binary.BigEndian)

	client, _ := packnet.Dial("tcp", "127.0.0.1:8080", proto)

	client.Start(nil)
```

Implement a message type.

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

Send a message to server.

```go
	message := TestMessage{ "Hello World!" }

	if err2 := client.Send(message); err2 != nil {
		panic(err2)
	}
```

Examples
=================

* [An echo server](//github.com/funny/examples/echo_server/)
* [An echo client](//github.com/funny/examples/echo_client/)
* [Broadcast server](//github.com/funny/examples/broadcast/)
