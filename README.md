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

link包核心由`Server`、`Session`、`CodecType`组成。`Server`和`Session`很容易理解，分别用于实现网络服务和连接管理。`CodecType`则提供具体的协议实现和io逻辑。

`Server`在使用的时候很简单，可以用 `link.Serve()`创建，也可以用`link.NewServer()`的方式创建。这样设计的目的是可以支持更多类型的`Listener`，不受限于net包。

`Session`在使用上分为两种情况，一种是由`Server.Accept()`产生的服务端会话，一种是由`link.Dial()`或`link.NewSession()`产生的客户端会话。

`CodecType`的设计目的是让每个`Session`都有各自的`Encoder`和`Decoder`用于消息的收发，这样才有机会实现有状态的通讯协议，比如有多阶段握手的通讯协议。

`Server`和`Session`上都有一个`interface{}`类型的`State`字段，可用于存储自定义状态。

`Session`上提供了关闭事件的监听机制，有一些应用场景需要在会话关闭时对一些资源做回收，就可以利用这个机制。

`Encoder`和`Decoder`都可以选择性的实现`Dispose()`方法，`Session`关闭时将会尝试调用这个方法，这可以可以做到`Encoder`和`Decoder`的资源回收利用，内置的`BufioCodecType`就利用这个机制引入了`sync.Pool`来提高对象的重用性。

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

示例，使用小端的`{packet, 4}`做分包协议，用Json作为消息格式：

```go
srv, err := link.Serve("tcp", "0.0.0.0:0", 
	link.Packet(4, 1024 * 1024, 4096, link.LittleEndian, link.Json()),
)
```

我是不会告诉你除了以上示例，阅读`all_test.go`和`example`目录下的代码也是很有帮助的！

内置类型
=======

link的核心部分代码是极少的，link另外提供了一些常用到的工具类型，下面一一对其进行介绍。

[channel.go](https://github.com/funny/link/blob/master/channel.go)
--------------

这个文件里实现了`Channel`类型的模板，`Channel`类型用于手工管理一组`Session`通常用于发送广播和维护在线列表。

这个文件是不参与编译的，所以link实际上不存在一个叫`Channel`的类型，这个文件只是提供类型的模板。

之前版本的通用`Channel.go`类型，用的是`Session.Id()`做key，这个设计会导致实际项目种出现类似这样的操作逻辑：

```
用户ID -> Session ID -> Session
```

而新的`Channel.go`类型可以自定义key类型，在上述场景中就可以直接用`用户ID`做key来索引`Session`：

```
用户ID -> Session
```

不同的应用场景会需要用不同的信息来索引`Session`，但是Go暂不支持泛型语法，所以我们通过`channel_gen.go`这个工具来生成具体的`Channel`类型的代码。

除了直观的可以看出少了一次map操作之外，其实额外维护一份`Session ID`映射关系也不是一件容易的事情，你需要重复`Channel.go`内部做的所有事情，而又不能重用`Channel.go`的代码。

所以自动生成代码的方式解决了以上所有问题，唯一需要做的就是手工执行一个命令。

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

需要在link包的根目录下执行`go generate channel.go`命令，来生成这些`Channel`类型，关于`go generate`的原理请参阅Go官方文档。

提示： 使用`Channel.Fetch()`进行遍历发送广播的时候，请注意存在io阻塞的可能，如果io阻塞会影响业务处理，可以通过异步发送的方式避免阻塞。

[codec_async.go](https://github.com/funny/link/blob/master/codec_async.go)
------------------

这个文件中实现了一个用于支持异步消息发送的`CodecType`。之前的版本中`Session`有一个`AsyncSend()`方法用于异步消息发送。我一直很不满意`AsyncSend()`的设计，从link包的历史版本中可以看到`AsyncSend()`经过了多次修改。

原因是不同的应用场景会有不同的异步消息发送需求，比如我们在游戏里很简单粗暴的把异步发送时出现chan阻塞的Session关闭掉，但是别的应用场景可能会需要等待一段时间后再重试，或者丢弃阻塞的消息，又或者阻塞允许一段时间，等到超时再做进一步处理。

所以`AsyncSend()`怎么改都不可能满足所有需求，最后我干脆删除它，由CodecType来决定消息是否异步发送，以及怎么进行异步发送。目前内置的`asyncCodecType`的逻辑是一旦遇到发送用的chan阻塞就立即返回`ErrBlocking`错误。如果这个设计不符合你的需求，你可以参考它实现出自己所需的异步发送逻辑。

需要注意，目前的`asyncCodecType`的设计会导致将`Session.Send()`的行为从同步变为异步，这样设计的目的是规避掉同时支持两种模式的复杂性，避免使用者误用。

对于高级用户如果需要同时支持同步和异步发送，可以自己实现一个`Encoder`在发送消息时通过判断消息类型来决定采用哪种发送方式，但是这样的设计需要周全的考虑各种并行执行的可能性。

[codec_bufio.go](https://github.com/funny/link/blob/master/codec_bufio.go)
------------------

这个文件中实现了带缓冲的IO以及缓冲对象重用，这是网络层很常用到的优化。缓冲读和缓冲写可以显著的降低实际的IO调用次数，在Go语言中一次实际的`net.Conn.Read()`调用开销并不低，它需要给文件句柄加锁然后放入事件循环里等待IO事件，这里面有一系列的系统调用。所以实际项目中，强烈建议使用bufio来降低IO调用次数。

有一个细节需要注意`sync.Pool`是跟着`BufioCodecType`实例的，所以在实际使用中，特别是创建客户端`Session`时，需要重用`BufioCodecType`而不是每次调用`link.Dial()`时都创建一个新的`BufioCodecType`实例。服务端不容易出现这个问题是因为`BufioCodecType`会被存在`Server`对象里，反复赋值给新建的服务端`Session`。

[codec_general.go](https://github.com/funny/link/blob/master/codec_general.go)
--------------------

里面实现了常见的Json、Gob、Xml格式的消息编解码，这三种消息格式都不需要分包协议就可以直接使用，但也可以跟分包协议配合使用。

[codec_packet.go](https://github.com/funny/link/blob/master/codec_packet.go)
--------------------

这是最实用的一个内置类型，里面实现了对应Erlang的`{packet, N}`格式的分包协议（`{N | 1,2,4,8}`）。

这种分包协议的包结构很简单，每个消息包由N个字节的固定长度的包头和不定长的包体组成，包头的数据是小端格式或者大段格式编码的包体长度值。分包的时候先读取包头，解码后获得包体长度，接着读取对应长度的数据即为包体。

这种分包协议简单易用，但是需要注意消息包的大小要控制好，否则容易成为漏洞被黑客利用，比如伪造一个长度超长的包头信息，让服务器一次申请一大块内存，或者频繁发送无效的大消息包，导致服务器内存耗尽。`Packet()`函数有一个`MaxPacketSize`参数用来限制最大包大小，当接收或发送的消息超过这个体积限制时，内部将抛出`ErrPacketTooLarge`错误，除了控制合理的`MaxPacketSize`之外，建议使用者在自己的网络层也要做好安全防范措施，比如控制每个用户的出错频率等。

优化提示1：在实践中，建议采用2字节包头结构，在需要发送大消息包的地方在协议上做消息分帧，而不是一次性发送一个大体积的消息包，这样除了起到安全防范作用，也可以获得较好的性能表现。

优化提示2：`PacketCodecType`消息解包时使用了`bufio.Reader`来减少的IO调用次数，通过`ReadBufferSize`参数调节`bufio.Reader`的缓存大小，细调这个参数可以获得较好的性价比。

优化提示3：在包体被完全缓存进`bufio.Reader`的情况下，可以在`Decoder`中直接取出`bufio.Reader`中的缓存进行反序列化操作，以减少数据拷贝。

示例，利用`github.com/funny/binary`包中提供的`BufioOptimizer`，直接拿`bufio.Reader`中的缓存进行消息反序列化：

```go
import (
	"io"
	"bufio"
	"github.com/funny/binary"
)

type MyCodecType struct {
}

func (ct MyCodecType) NewDecoder(r io.Reader) link.Decoder {
	// PacketCodecType传递过来的是一个io.LimitedReader
	lr := r.(*io.LimitedReader)
	return &MyDecoder {
		lr: lr, 
		// io.LimitedReader包裹着bufio.Reader
		bo: binary.BufioOptimizer{
			R: lr.R.(*bufio.Reader),
		},
	}
}

type MyDecoder struct {
	lr *io.LimitedReader
	bo binary.BufioOptimizer
}

func (d *MyDecoder) Decode(msg interface{}) error {
	var reader, err := d.bo.Next(d.lr.N)
	if err != nil {
		return err
	}
	// 接着就可以用reader进行消息反序列化了
}
```

优化提示4：在某些应用场景下，消息序列化之前就可以知道消息体积，这时候可以利用`PacketBuffer.Next(n)`取出buffer直接将数据序列化进去，这可以避免动态扩容并减少数据拷贝。

示例，使用`PacketBuffer.Next(n)`特性：


```go
import (
	"io"
	"bufio"
	"github.com/funny/binary"
)

type MyMessage interface {
	BinarySize() int
}

type MyCodecType struct {
}

func (ct MyCodecType) NewEncoder(w io.Writer) link.Encoder {
	return &MyEncoder {
		buf: w.(*link.PacketBuffer),
	}
}

type MyEncoder struct {
	buf *link.PacketBuffer	
}

func (e *MyEncoder) Encode(msg interface{}) error {
	n := msg.(MyMessage).BinarySize()
	buf := binary.Buffer{Data: e.buf.Next(n)}

	// 接着就可以用buf进行消息序列化了
}
```

[codec_safe.go](https://github.com/funny/link/blob/master/codec_safe.go)
-----------------

里面实现了线程安全的`CodecType`，旧版本的link里`Session`内置了收发锁让`Session.Receive()和`Session.Send()`可以被并发调用。但是实际项目中并发接收或者并发发送的场景很少，如果一开始就内置到`Session`里，这部分调用开销就多余了。

所以后来我删除了`Session`里面加锁的逻辑，引入了`ThreadSafe()`。在需要对收发过程进行加锁保护的时候可以用它。

消息分发
=======

实际项目中通常都会需要识别消息类型然后执行不同消息类型的解包（反序列化）接着调用不同的消息处理过程（业务逻辑），link包的测试代码和示例代码都没有直接体现出如何做消息分发，所以新手经常会卡在这一步。下面我就简单的演示怎样实现一个可以识别消息类型的CodecType，并演示如何进行消息分发。

首先我先假定我们需要实现这样一个协议格式：

```
4字节的包头 + 2字节的消息类型ID + 消息内容
```

通过消息类型ID的区分我们就可以支持一种以上的消息类型，消息内容格式这里不举例，我们只需要假设它们内容格式都不一样。

因为link已经内置了4字节包头的分包协议支持，所以我们不需要自己做分包，只需要实现一个识别消息类型的CodecType。

实现起来大概像这样子：

```go
package xxoo

import (
	"io"
	"github.com/funny/binary"
)

type MyCodecType struct {}

func (codecType MyCodecType) NewDecoder(r io.Reader) link.Decoder {
	return &MyDecoder{binary.Reader{R:r}}
}

func (codecType MyCodecType) NewEncoder(w io.Writer) link.Encoder {
	return &MyEncoder{binary.Writer{W:w}}
}

type MyDecoder struct {
	r binary.Reader
}

func (decoder *MyDecoder) Decode(msg interface{}) error {
	switch decoder.r.ReadUint16LE() {
	case 1:
		decoder.MessageType1(msg)
	case 2:
		decoder.MessageType2(msg)
	default:
		return errors.New("unknow message type")
	}
	if decoder.Error() != nil {
		// MyMessage接口说明在下面，请继续阅读文档
		*(msg.(*MyMessage)) = nil
		return decoder.Error()
	}
	return nil
}
```

上面示例只给出Decoder的结构，Encoder只是反过程，这里就不再重复。

把这个CodecType和link内置的分包协议结合起来创建一个TCP服务：

```
srv, err := link.Serve("tcp", "0.0.0.0:0", 
	link.Packet(4, 1024 * 1024, 4096, link.LittleEndian, MyCodecType{}),
)
```

现在消息可以按类型解析了，但是接收消息时link要求传入一个消息对象给`Session.Receive()`，这不就成了先有鸡还是先有蛋的问题了吗？在知道消息类型前我们怎么直到应该传入什么消息类型的对象呢？

这边需要脑筋急转弯一下，利用Go的interface机制可以解决这个问题并顺便帮我们把协议解析和消息分发解耦开。

我们知道所有的请求都需要被分发处理，那么我们可以定义这样一个接口：

```go
type MyMessage interface {
	Dispatch()
}
```

分发处理的入口要叫Dispatch还是Process还是Handle请自便，这里只是举例。

所有的上行消息（请求）都实现这个接口：

```go
type MessageType1 struct {
	Field1 int32
	Field2 int64
}

func (msg *MessageType1) Dispatch() {
	// 做爱做的事
}
```

这样我们就可以自然而然的这样调用link：

```go
var msg MyMessage

session.Receive(&msg)

msg.Dispatch()
```

注意传入`Receive()`的参数是`MyMessage`接口类型的指针，所以在赋值的时候需要这样写：

```go
func (decoder *MyDecoder) MessageType1(msg interface{}) {
	*(msg.(*Message)) = &MessageType1 {
		Field1: decoder.r.ReadInt32LE(),
		Field2: decoder.r.ReadInt64LE(),
	}
}
```

具体的Dispatch内是通过怎样的机制把消息分发给对应的业务接口的，这就八仙过海各显神通了，我在项目里用的是注册回调函数的方式，大家可以根据实际的项目情况设计。

包括上面示例中的DecodeType1和DecodeType2，实际项目中不一定是这样做的，接口比较多的项目里通常会需要把不同业务模块的消息类型分到不同的包里，示例只是提供思路，希望大家要灵活变通不要死记硬背。

总结
====

从link的核心代码和内置类型可以看出，核心其实很简单，IO调用方式和协议实现都靠`CodecType`解耦了。

建议在实际项目中根据项目需求，参考内置类型的设计实现针对项目的CodecType，这样可以得到最好的执行效率和使用体验。

如果有问题或者改进建议，欢迎加技术交(xian)流(liao)群一起讨论：188680931
