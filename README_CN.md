简介
====

这是一个简易的Go语言网络库，它专注于解决基于消息包的长连接通讯需求。

它内置了类似于Erlang的`{packet: N}`那样的分包协议，同时也支持自定义的分包协议。

但是它并不限制请求和响应消息中的编码和解码格式。

这个库中还提供了Session管理和广播等功能，可加速您的开发效率。

安装
====

```
go get github.com/funny/link
```

使用
====

首先，为您的项目选择一个分包协议，这里我们使用大端格式的2个字节长度包头：

```go
proto := link.PacketN(2, link.BigEndian, link.DefaultBuffer)
```

在指定的端口启动一个服务器：

```go
server, _ := link.Listen("tcp", "0.0.0.0:8080", proto)
```

处理新进连接，并为新的Session设置消息处理器：

```go
server.Handle(func(session *link.Session) {
	fmt.Println("session start")

	session.Handle(func(session *link.Session, msg link.Buffer) {
		fmt.Printf("new message: %s\n", msg.Data())
	})

	fmt.Println("session closed")
})
```

在客户端，使用同样的分包协议连接到服务器：

```go
proto := link.PacketN(2, link.BigEndian, link.DefaultBuffer)

client, _ := link.Dial("tcp", "127.0.0.1:8080", proto)
```

发送一个消息给服务端：

```go
client.Send(link.Binary("Hello World!"))
```

示例
====

* [Echo server](https://github.com/funny/link/blob/master/examples/echo_server/main.go)
* [Echo client](https://github.com/funny/link/blob/master/examples/echo_client/main.go)
* [Broadcast server](https://github.com/funny/link/blob/master/examples/broadcast/main.go)
* [Benchmark tool](https://github.com/funny/link/blob/master/examples/benchmark/main.go)

文档
====

[Let's Go!](http://godoc.org/github.com/funny/link)
