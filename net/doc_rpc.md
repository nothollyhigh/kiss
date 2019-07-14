## 网络通信的两种模式

1. 请求-应答式：客户端请求服务器返回结果，client<->server

2. 通知式：无需应答，方向上可以分为 client->server 和 server->client


## RPC（Remote Procedure Call）的本质

- 通常的rpc是单向的（client->server）、阻塞方式调用

- 通常的rpc是绑定了序列化方案的，比如grpc用protobuf

- 远程过程调用，从本质上讲，http也算rpc，只是传统rpc都是长连接，相比之下http太浪费

- 服务端方法要按格式写代码，像grpc这种，代码可读性交叉，使用也比较拘束

- 通常的rpc服务端实现，方法返回即调用结束，不支持异步返回


## KISS 的 RPC

- 支持client单向阻塞方式调用，也支持server向client发送通知

- 用户可以自主选择序列化方案，默认支持json

- 使用tcp协议，并且与kiss/net的tcp协议兼容

- 写法像http handler一样简单，不需要像grpc一样按格式写一大堆丑陋的代码

- 服务端方法支持异步回包


## KISS RPC 的缺点

- 作者精力有限，目前只支持golang


## 示例代码

- rpc server

```golang
package main

import (
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/net"
	"time"
)

var (
	addr = "0.0.0.0:8888"
)

type HelloRequest struct {
	Message string
}

type HelloResponse struct {
	Message string
}

// Hello方法
func onHello(ctx *net.RpcContext) {
	req := &HelloRequest{}

	err := ctx.Bind(req)
	if err != nil {
		log.Error("onHello failed: %v", err)
		return
	}

	err = ctx.Write(&HelloResponse{Message: req.Message})
	if err != nil {
		log.Error("onHello failed: %v", err)
		return
	}

	log.Info("HelloRequest: %v", req.Message)
}

func main() {
	server := net.NewRpcServer("Rpc")

	// 初始化方法，类似http初始化路由
	server.HandleRpcMethod("Hello", onHello, true)

	// 启动服务
	server.Serve(addr, time.Second*5)
}
```

- rpc client

```golang
package main

import (
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/net"
	"time"
)

var (
	addr = "0.0.0.0:8888"
)

type HelloRequest struct {
	Message string
}

type HelloResponse struct {
	Message string
}

func onConnected(c *net.TcpClient) {
	log.Info("RpcClient OnConnected")
}

func main() {
	engine := net.NewTcpEngine()
	client, err := net.NewRpcClient(addr, engine, nil, onConnected)
	if err != nil {
		log.Panic("NewReqClient Error: %v", err)
	}

	for {
		req := &HelloRequest{Message: "kiss"}
		rsp := &HelloResponse{}

		// 调用Hello方法
		err = client.Call("Hello", req, rsp, time.Second*3)
		if err != nil {
			log.Error("Hello failed: %v", err)
		} else {
			log.Info("HelloResponse: %v", rsp.Message)
		}

		time.Sleep(time.Second)
	}
}
```