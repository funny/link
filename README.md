[![Go Foundation](https://img.shields.io/badge/go-foundation-green.svg)](http://golangfoundation.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/funny/link)](https://goreportcard.com/report/github.com/funny/link)
[![Build Status](https://travis-ci.org/funny/link.svg?branch=master)](https://travis-ci.org/funny/link)
[![codecov](https://codecov.io/gh/funny/link/branch/master/graph/badge.svg)](https://codecov.io/gh/funny/link)
[![GoDoc](https://img.shields.io/badge/api-reference-blue.svg)](https://godoc.org/github.com/funny/link)


介绍
====

最初开发这个包的目的是提炼一套可以在公司内多个项目间共用的网络层，因为在项目中我发现不同的网络应用一直重复一些相同或相类似的东西，比如最常用到的就是会话管理，不管是做游戏的前端连接层还是做服务器和服务器之间的RPC层或者是游戏的网关，虽然协议不一样但是它们都会需要会话的管理。会话管理看似简单，但是任何简单的问题一旦涉及到并发都会变得复杂起来，所以看似简单的会话管理每次实现起来都得再配套做单元测试甚至线上实际运行几个版本才能放心。所以我决定提取这些公共的部分，避免那些容易引入BUG的重复劳动。

但是在提取这些公共部分的时候并没有期初想象的那么容易，因为不同的应用场景有不同的需求，比如有的场景需要异步，有的场景需要同步，有的协议需要握手过程，有的则需要keepalive。从代码的提交历史里面可以看出这个包前后经过了很多次大的调整，因为要做一个能满足所有需求的通用网络层真的很难。

重构过程就像在简化公式，经过不断的提炼目前link变得十分的简单，同时它的定位也很清楚。link不是一个完整网络层也不是一个框架，它只是一个脚手架，它可以帮助你快速的实现出你所需要的网络层或者通讯框架，帮你约束网络层的实现方式，不至于用不合理的方式实现网络层，除此之外它不会管更多的事情。

link是协议无关的，使用link只需要理解少数几个概念就可以上手了。

核心
====

link包的核心是`Session`，`Session`的字面含义是`会话`，就是一次对话过程。每一个连接的生命周期被表达为一个会话过程，这个过程中通讯双方的消息有来有往。

会话过程所用的具体通讯协议通过`Codec`接口解耦。通过`Codec`接口可以自定义通讯的IO实现方式，如：TCP、UDP、UNIX套接字、共享内存等，也可以自定义流的实现方式，如：压缩、加密、校验等，也可以实现自定义的协议格式，如：JSON、Gob、XML、Protobuf等。

在实际项目中，通常不会只有一个会话，所以link提供了几种不同的`Session`管理方式。

`Manager`是最基础的`Session`管理方式，它负责创建和管理一组`Session`。`Manager`是不与通讯形式关联的，与通讯有关联的`Manager`叫`Server`，它的行为比`Manager`更具体，它负责从`net.Listener`上接收新连接并创建`Session`，然后在独立的goroutine中处理来自新连接的消息。

link还提供了`Channel`用于对`Session`进行按需分组，`Channel`用key-value的形式管理`Session`，`Channel`的key类型通过代码生成的形式来实现自定义。

示例
=======

link在`codec`目录下实现了集中简单的协议类型，可作为示例。

示例，创建一个使用Json作为消息格式的TCP服务端：

```go
package main

import (
	"log"

	"github.com/funny/link"
	"github.com/funny/link/codec"
)

type AddReq struct {
	A, B int
}

type AddRsp struct {
	C int
}

func main() {
	json := codec.Json()
	json.Register(AddReq{})
	json.Register(AddRsp{})

	server, err := link.Serve("tcp", "0.0.0.0:0", json, 0 /* sync send */)
	checkErr(err)
	addr := server.Listener().Addr().String()
	go server.Serve(link.HandlerFunc(serverSessionLoop))

	client, err := link.Connect("tcp", addr, json, 0)
	checkErr(err)
	clientSessionLoop(client)
}

func serverSessionLoop(session *link.Session) {
	for {
		req, err := session.Receive()
		checkErr(err)

		err = session.Send(&AddRsp{
			req.(*AddReq).A + req.(*AddReq).B,
		})
		checkErr(err)
	}
}

func clientSessionLoop(session *link.Session) {
	for i := 0; i < 10; i++ {
		err := session.Send(&AddReq{
			i, i,
		})
		checkErr(err)
		log.Printf("Send: %d + %d", i, i)

		rsp, err := session.Receive()
		checkErr(err)
		log.Printf("Receive: %d", rsp.(*AddRsp).C)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
```

补充说明
=======

[channel.go](https://github.com/funny/link/blob/master/channel.go)
--------------

这个文件里实现了`Channel`类型用于管理一组`Session`，通常用于发送广播和维护在线列表。

之前版本的`Channel`类型，用的是`Session.ID()`做key，这个设计会导致实际项目种出现类似这样的操作逻辑：

```
取用户ID -> 从自己维护的映射关系中取用户ID对应的Session ID -> 到Channel里取Session
```

而新的`Channel`类型可以自定义key类型，在上述场景中就可以直接用`用户ID`做key来索引`Session`：

```
取用户ID -> 到Channel里取Session
```

除了直观的可以看出少了一次map操作之外，其实额外维护一份`Session ID`映射关系也不是一件容易的事情，你需要重复`Channel`内部做的所有事情，而又不能重用`Channel`的代码。

使用`Channel.Fetch()`进行遍历发送广播的时候，请注意存在IO阻塞的可能，如果IO阻塞会影响业务处理，就需要使用异步发送。

相关项目
====

* [通讯协议代码生成](https://github.com/funny/fastbin)
* [用link开发的网关](https://github.com/funny/fastway)
* [slab算法的内存池](https://github.com/funny/slab)
* [支持断线重连和加密的流协议](https://github.com/funny/snet)

参与项目
=======

欢迎提交通过github的issues功能提交反馈或提问。

技术群：474995422