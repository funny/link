2015-08-24

* 将会话管理从`Server`中剥离到`Manager`
* 将`Encoder`和`Decoder`合并为`Codec`
* 将`CodecType`改为`Protocol`
* `Session`通过`SendChanSize`决定发送是异步还是同步

2015-11-23

* 修复了一些BUG，更新了示例代码
* BufioCodecType进一步优，把Encoder和Decoder也纳入对象重用。
* AsyncCodecType去掉了AsyncMsg{}的消息包裹需求，避免用户用错。但是这样的设计会导致一旦加了AsyncCodecType，Session.Send就会变成异步发送消息。如果用户需要实现可选的异步消息发送，可以参考原来的AsyncMsg{}设计，但是需要注意线程安全
* 加入了基础的分包协议，尽量做到不重复申请内存不做多余的数据拷贝

2015-11-22

* 把BufioCodecType从示例项目中移回link包，并添加sync.Pool优化
* 添加SafeCodecType用来进行可选的收发过程加锁
* 删掉了Session.AsyncSend()，改用AsyncCodecType来提供默认的异步发送，用户可以通过CodecType实现自己的异步发送逻辑
* 为Encoder和Decoder引入可选的Dispose()方法，可以在Session关闭时做资源回收
* 补充了缺失已久的文档

2015-11-21

* 去掉了Session上的收发加锁代码，如果需要对收发过程做并发保护，可以在Encoder和Decoder中自行加锁
* 完善Server.Stop()的逻辑，当真正被执行到的Server.Stop()过程还没结束前，其余的Server.Stop()调用都会等待而不是直接返回，否则可能出现Stop过程还没真正结束后续逻辑就被继续执行的问题。

2015-06-23

* 把分包协议和流协议都放进link包
* 加上StreamCodec和PacketCodec的设计，适应不同协议类型和不同消息格式的需求

2015-06-09

* 把Gob消息移到stream包，因为Gob消息不需要外部做消息分包
* 把广播协议拆分到独立的接口，去掉Server上的广播方法，让设计更清晰

2015-06-03

* 重构完成，加入网关协议并单元测试网关主要功能

2015-05-29

* 将消息收发的具体实现剥离到独立的包，link核心只保留Server、Session、Channel
* 进一步抽象接口，由外部来自定义具体的消息序列化和反序列化接口
* 实现流协议和分包协议两种协议类型模块

2015-05-27

* Session的异步发送去掉了超时设置，在异步发送管道阻塞的时候直接当成故障连接关闭
* Session的异步发送用Goroutine改为按需创建
* 压力测试工具加入随机大小的设置项，测试Go是否会持续增长内存

2015-05-21

* 将旧版的Buffer和新版的Conn中关于协议解析的方法统一成binary.Reader和binary.Writer
* 将旧版Buffer实现io相关接口的部分提取到binary.Buffer
* 将InMessage和OutMessage的参数改为binary.Reader和binary.Writer，使用者既可以使用基于bufio的通讯层，也可以实现基于Buffer的网络层

2015-05-19

* 简化设计，明确库的定位
* 去掉Buffer，改为统一使用bufio
* 完善压测程序的逻辑

2015-05-14

* 加入消息帧的支持
* 加入行分割协议的支持

2015-05-13

* 重新实现内存池，将内存池中的内存块按大小分成几类，并实现池中小对象多于大对象的算法
* 合并和InBuffer和OutBuffer，解决Buffer在写入时不能从内存池分配新内存块的问题
* 重构了协议实现时所需的几个接口，让协议实现可以支持握手过程，解耦请求解析和请求处理这两个过程
* 将定长包头的协议分离到独立的包，并另外实现了变长包头协议，用于测试新版接口灵活性
* 解决了示例中压测工具在测试结束时可能永久等待消息接受的并发BUG
* 压力测试工具加入多进程模式，避免压测时受到单进程线程数上限限制