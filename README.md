[![Build Status](https://travis-ci.org/funny/link.svg?branch=master)](https://travis-ci.org/funny/link)
[![Coverage Status](https://coveralls.io/repos/funny/link/badge.svg?branch=master&service=github)](https://coveralls.io/github/funny/link?branch=master)

介绍
====

最初开发这个包的目的是提炼一套可以在公司内多个项目间共用的网络层，因为在项目中我发现不同的网络应用一直重复一些相同或相类似的东西，比如最常用到的就是会话管理，不管是做游戏的前端连接层还是做服务器和服务器之间的RPC层或者是游戏的网关，虽然协议不一样但是它们都会需要会话的管理。会话管理看似简单，但是涉及到并发简单的需求就变得复杂起来，所以看似简单的会话管理每次实现起来都得再配套做单元测试甚至线上实际运行几个版本才能放心。所以我决定提取这些公共的部分，避免那些容易引入BUG的重复劳动。

但是在提取这些公共部分的时候并没有期初想象的那么容易，因为不同的应用场景有不同的需求，比如有的场景需要异步，有的场景需要同步，有的协议需要握手过程，有的则需要keepalive。从代码的提交历史里面可以看出这个包前后经过了很多次大的调整，因为要做一个能满足所有需求的通用底层真的很难。

经过不断的提炼，就像在简化公式一样的，link变得十分的简单，同时它的定位也很清楚。link不是一个网络层也不是一个框架，它只是一个脚手架，但它可以帮助你快速的实现出你所需要的网络层或者框架，帮你约束网络层的实现方式，不至于用不合理的方式实现网络层，除此之外它不会管更多的事情。

link是协议无关的，使用link只需要理解`CodecType`的概念，你可以加入任何你需要的通讯协议实现。

基础
====

link包核心由`Server`、`Session`、`CodecType`组成。`Server`和`Session`很容易理解，分别用于实现网络服务和连接管理，`CodecType`则提供具体的协议实现和IO处理。

`Server`在使用的时候很简单，可以用 `link.Serve()`创建，也可以用`link.NewServer()`的方式创建，这样设计的目的是可以支持更多类型的`Listener`，不受限于net包。

`Session`在使用上分为两种情况，一种是由`Server.Accept()`产生的服务端会话，一种是由`link.Dial()`或`link.NewSession()`产生的客户端会话。

`CodecType`的设计目的是让每个`Session`都有各自的`Encoder`和`Decoder`用于消息的收发，这样才有机会实现有状态的通讯协议或者有状态的IO，比如有多阶段握手的通讯协议和使用`bufio.Reader`。

`Server`和`Session`上都有一个`interface{}`类型的`State`字段，可用于存储自定义状态。

`Session`上提供了关闭事件的监听机制，有一些应用场景需要在会话关闭时对一些资源做回收就可以利用这个机制。

`Encoder`和`Decoder`都可以选择性的实现`Dispose()`方法，`Session`关闭时将会尝试调用这个方法，可以通过这个方法来实现`Encoder`和`Decoder`的内部资源回收利用，内置的`BufioCodecType`就利用这个机制引入了`sync.Pool`来重用`bufio.Reader`。

一些示例
=======

示例，创建一个使用Json作为消息格式的TCP服务端：

```go
srv, err := link.Serve("tcp", "0.0.0.0:0", link.Json())
```

示例，使用Bufio优化IO：

```go
srv, err := link.Serve("tcp", "0.0.0.0:0", link.Bufio(link.Json()))
```

示例，加入线程安全：

```go
srv, err := link.Serve("tcp", "0.0.0.0:0", link.ThreadSafe(link.Json()))
```

示例，把发送方式改为异步：

```go
srv, err := link.Serve("tcp", "0.0.0.0:0", link.Async(link.Json()))
```

除了以上示例，阅读`all_test.go`和`example`目录下的代码也是很重要的示例。

内置类型
=======

link的核心部分代码极少，基于这个核心link提供了一些常用到的工具类型来辅助项目开发，下面一一对这些工具类型进行介绍。

[channel.go](https://github.com/funny/link/blob/master/channel.go)
--------------

这个文件里实现了`Channel`类型的模板，`Channel`类型用于管理一组`Session`，通常用于发送广播和维护在线列表。

这个文件是不参与编译的，link实际上不存在一个叫`Channel`的类型，这个文件只是`Channel`类型的模板。

之前版本的通用`Channel.go`类型，用的是`Session.Id()`做key，这个设计会导致实际项目种出现类似这样的操作逻辑：

```
取用户ID -> 从自己维护的映射关系中取用户ID对应的Session ID -> 到Channel里取Session
```

而新的`Channel.go`类型可以自定义key类型，在上述场景中就可以直接用`用户ID`做key来索引`Session`：

```
取用户ID -> 到Channel里取Session
```

除了直观的可以看出少了一次map操作之外，其实额外维护一份`Session ID`映射关系也不是一件容易的事情，你需要重复`Channel.go`内部做的所有事情，而又不能重用`Channel.go`的代码。

不同的应用场景会需要用不同的信息来索引`Session`，但是Go暂不支持泛型语法，所以我们通过`channel_gen.go`这个工具来生成具体的`Channel`类型的代码。

举例，生成一个用`uint64`类型作为key的`Channel`：

```
go run channel_gen.go Uint64Channel uint64 channel_uint64.go
```

`channel_gen.go`的参数列表如下：

* 第一个参数为类型名，不一定非得叫Channel，可以根据实际使用场景来命名
* 第二个参数是key的类型，通常是int之类的简单类型，但如果需要同时根据多个条件索引Session，可以使用结构体做key
* 第三个参数是输出的代码文件名
* 第四个参数是可选的包名称，没有指定此参数时生成的代码归属于link包，你可以通过这个参数生成归属于自己包的代码

此外，link借助`go generate`命令内置了一组常用到的`Channel`类型的代码生成。因为这些代码是工具自动生成的，所以不纳入版本管理，在刚拿到link包的代码时是找不到这些代码的。

需要在link包的根目录下执行`go generate channel.go`命令来生成这些类型。

关于`go generate`的原理请参阅Go官方文档。

提示： 使用`Channel.Fetch()`进行遍历发送广播的时候，请注意存在IO阻塞的可能，如果IO阻塞会影响业务处理，就需要使用异步发送，关于异步发送请参考`codec_async.go`的说明。

[codec_async.go](https://github.com/funny/link/blob/master/codec_async.go)
------------------

这份代码中实现了一个用于支持异步消息发送的`CodecType`。之前的版本中`Session`有一个`AsyncSend()`方法用于异步消息发送。我一直很不满意`AsyncSend()`的设计，从link包的历史版本中可以看到`AsyncSend()`经过了多次修改。

原因是不同的应用场景会有不同的异步消息发送需求，比如我们在游戏里很简单粗暴的把异步发送时出现chan阻塞的Session关闭掉，但是别的应用场景可能会需要等待一段时间后再重试，或者丢弃阻塞的消息，又或者阻塞允许一段时间等到超时再做进一步处理。

需求多种多样，所以`AsyncSend()`怎么改都不可能满足所有需求，最后我干脆删除它，由CodecType来决定消息是否异步发送，以及怎么进行异步发送。

目前内置的`asyncCodecType`的逻辑是一旦遇到发送用的chan阻塞就立即关闭`Writer`并返回`ErrBlocking`错误。如果这个设计不符合你的需求，你可以参考它实现出自己所需的异步发送逻辑。

需要注意，目前的`asyncCodecType`的设计会将`Session.Send()`的行为从同步变为异步，这样设计的目的是规避掉同时支持两种模式的复杂性，避免使用者误用。

对于高级用户如果需要同时支持同步和异步发送，可以自己实现一个`Encoder`在发送消息时通过判断消息类型来决定采用哪种发送方式，但是这样的设计需要周全的考虑各种并行执行的可能性。

[codec_bufio.go](https://github.com/funny/link/blob/master/codec_bufio.go)
------------------

这份代码中实现了带缓冲的IO以及`bufio.Reader`重用。缓冲读和缓冲写可以显著的降低实际的IO调用次数，在Go语言中一次实际的`net.Conn.Read()`调用开销并不低，它需要完成给文件句柄加锁然后放入事件循环里等待IO事件等一系列动作。所以实际项目中，强烈建议使用`bufio.Reader`来降低IO调用次数。

有一个细节需要注意`sync.Pool`是跟着`BufioCodecType`实例的，所以在实际使用中，特别是创建客户端`Session`时，需要重用`BufioCodecType`而不是每次调用`link.Dial()`时都创建一个新的`BufioCodecType`实例。服务端不容易出现这个问题是因为`BufioCodecType`会被存在`Server`对象里反复赋值给新建的`Session`。

[codec_general.go](https://github.com/funny/link/blob/master/codec_general.go)
--------------------

这份代码中实现了常见的Json、Gob、Xml格式的消息编解码，这三种消息格式都不需要额外分包协议就可以直接使用，但也可以跟分包协议配合使用。

[codec_fastbin.go](https://github.com/funny/link/blob/master/codec_fastbin.go)
--------------------

这个文件中实现了link和[`fastbin`](https://github.com/funny/fastbin)的配套接口，用fastbin生成的消息接口和消息编解码可以通过此CodecType跟link配套使用。

fastbin用的是每个消息4个字节的包头来进行分包和消息识别于派发，包头的前2个字节是小端格式编码的包体长度值，第3个字节为服务类型ID，第4个字节为消息类型ID。

所以fastbin支持256个服务类型，每个服务类型中可处理256种消息。

由于包头固定是2个字节，所以最大的消息长度是64K，实际应用场景中如果有可能出现超过此大小的消息，需要自己再封装一层消息分帧。

在使用次类型时，`Session`接收的消息必须是`*FbRequest`，发送的消息必须实现`FbMessage`接口。

接收消息后，通过调用`*FbRequest`的`Process`方法来进行请求处理：

```go
var req FbRequest

session.Receive(&req)

// 用默认的Session实现
req.Process(FbSessionWrapper{session})
```

Process方法之所以用`link.FbSession`接口类型做参数而不用`*link.Session`，目的是要跟`*link.Session`解耦，在项目中才有机会做自定义的Session管理。

fastbin不一定符合你的项目需求，如果要自己实现分包、消息识别和消息分发可以把这份代码当成示例来用。

[codec_safe.go](https://github.com/funny/link/blob/master/codec_safe.go)
-----------------

这份代码实现了线程安全的`CodecType`，旧版本的link里`Session`内置了收发锁让`Session.Receive()和`Session.Send()`可以被并发调用。但是实际项目中并发接收或者并发发送的场景很少，如果一开始就内置到`Session`里，这部分调用开销就多余了。

所以后来我删除了`Session`里面加锁的逻辑，引入了`ThreadSafe()`。在需要对收发过程进行加锁保护的时候可以用它。

总结
====

link的核心其实很简单，IO调用方式和协议实现都靠`CodecType`解耦，理解了`CodecType`就能熟练的用link搭建针对各种场景的网络层。

建议在实际项目中根据项目需求，参考内置类型的设计实现针对项目的`CodecType`，这样可以得到最好的执行效率和使用体验。

如果有问题或者改进建议，欢迎加技术交(xian)流(liao)群一起讨论：188680931

附录
====

* 也许用得上的免配置通用网关 - [https://github.com/funny/gateway](https://github.com/funny/gateway)
* 配套`codec_fastbin.go`使用的fastbin项目 - [https://github.com/funny/fastbin](https://github.com/funny/fastbin)
* `codec_fastbin.go`中可以用到的内存池 - [https://github.com/funny/slab](https://github.com/funny/slab)