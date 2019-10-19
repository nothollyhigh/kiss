# KISS - Keep It Simple & Stupid

[![MIT licensed][1]][2]
[![Go Report Card][3]][4]

[1]: https://img.shields.io/badge/license-MIT-blue.svg
[2]: LICENSE.md
[3]: https://goreportcard.com/badge/github.com/nothollyhigh/kiss
[4]: https://goreportcard.com/report/github.com/nothollyhigh/kiss


- [KISS原则](https://zh.wikipedia.org/wiki/KISS%E5%8E%9F%E5%88%99) 是指在设计当中应当注重简约的原则

- 作者水平有限，欢迎交流和指点，qq群： 817937655


## 安装
* go get github.com/nothollyhigh/kiss/...


## KISS可以用来做什么？

- 有的人喜欢"框架"这个词，KISS的定位是提供一些基础组件方便搭积木实现架构方案，组件不限于项目类型

- 作者主要从事游戏和web服务器开发，常用来构建游戏服务器，一些示例：

> 1. [单进程服务器示例](https://github.com/nothollyhigh/hellokiss)

> 2. [服务器集群示例](https://github.com/nothollyhigh/kisscluster)，另：集群是不同功能服务的拆分和实现，每个游戏的需求都可能不一样，请根据实际需求自行设计和实现

> 3. [kissgate网关](https://github.com/nothollyhigh/kissgate)，支持kiss格式的tcp/websocket连接反向代理到tcp服务，支持线路检测、负载均衡、realip等，常用来做游戏集群的网关，kiss协议格式详见 [net包](https://github.com/nothollyhigh/kiss/blob/master/net/README.md)


## KISS组件包简介

### 一、[net，网络包](https://github.com/nothollyhigh/kiss/blob/master/net/README.md)

1. Tcp
   可以用做游戏服务器，支持自定义协议格式、压缩、加密等

2. Websocket
   可以用做游戏服务器，支持自定义协议格式、压缩、加密等

3. Rpc
   可以灵活使用任意序列化、反序列化，给用户更多自由，如protobuf、json、msgpack、gob等
   支持服务端异步处理，服务端不必须在方法中处理完调用结果，可以异步处理结束后再发送结果
   不像GRPC等需要生成协议、按格式写那么多额外的代码，用法上像写 net/http 包的路由一样简单

4. Http
   支持优雅退出、pprof等

- 详见 [net](https://github.com/nothollyhigh/kiss/blob/master/net/README.md)

### 二、[log，日志包](https://github.com/nothollyhigh/kiss/blob/master/log/README.md)

- 自己实现 log 包之前，我简单尝试过标准库的 log 和一些三方的日志包，但是对日志文件落地不太友好，
  比如日志文件按目录拆分、文件按时间和size切分、日志位置信息等

- KISS 的 log 包日志支持：

1. 日志位置信息，包括文件、行数
2. 支持文件日志，支持bufio
3. 文件日志按时间拆分目录
4. 文件日志按时间格式切分
5. 文件日志按size切分
6. 支持钩子对日志做结构化或其他自定义处理

- 详见 [log](https://github.com/nothollyhigh/kiss/blob/master/log/README.md)

### 三、[sync包](https://github.com/nothollyhigh/kiss/blob/master/sync/README.md)

- 开启debug支持死锁告警
  web相关无状态服务通常不需要锁，游戏逻辑多耦合，不小心可能导致死锁，可以用这个包来排查

- WaitSession用法类似标准库的WaitGroup，但是可以指定session进行等待，支持超时

- 详见 [sync](https://github.com/nothollyhigh/kiss/blob/master/sync/README.md)

### 四、[timer，定时器](https://github.com/nothollyhigh/kiss/blob/master/timer/README.md)

- 标准库的 time.AfterFunc 触发时会创建一个协程来调用回调函数，大量定时器短时间内集中触发时开销较大

- KISS 的 timer 是小堆实现的，一个实例只需要一个协程就可以管理大量定时器，支持同步异步接口
  主要用于优化标准库 time.AfterFunc 的协程开销，但要注意线头阻塞的问题
  耗时较长的定时器回调建议仍使用 time.AfterFunc

- 详见 [timer](https://github.com/nothollyhigh/kiss/blob/master/timer/README.md)

### 五、[event，事件包](https://github.com/nothollyhigh/kiss/blob/master/event/README.md)

- 进程内的发布订阅组件，观察者模式，可以用于模块间解耦

- 详见 [event](https://github.com/nothollyhigh/kiss/blob/master/event/README.md)

### 六、[graceful](https://github.com/nothollyhigh/kiss/blob/master/graceful/README.md)

- 优雅管理子模块的包，方便封装不同功能模块，每个模块一个逻辑协程，调用模块的Exec来串行化逻辑操作以避免加锁

- graceful.Register可注册多个模块，实现相应接口，并由graceful自动管理启动和优雅退出等

- 详见 [graceful](https://github.com/nothollyhigh/kiss/blob/master/graceful/README.md)

### 七、[util，杂货铺](https://github.com/nothollyhigh/kiss/blob/master/util/README.md)

- 最常用的 Go，HandlePanic，Safe，处理panic，打印异常调用栈信息

- Workers，任务池，用于控制一定数量的协程异步处理任务
  WorkersLink，也是任务池，但在Workers基础上做了一点扩展，支持异步处理的顺序组装

- Qps，方便统计、打印一些qps功能

- 详见 [util](https://github.com/nothollyhigh/kiss/blob/master/util/README.md)
