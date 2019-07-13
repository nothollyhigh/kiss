## 目录

- [协议格式](#协议格式)
- [Tcp Echo](#tcp-echo)
	- [tcp echo server](#tcp-echo-server)
	- [tcp echo client](#tcp-echo-client)
- [Websocket Echo](#websocket-echo)
	- [raw ws echo server](#tcp-echo-server)
	- [raw ws echo client](#tcp-echo-client)
	- [kiss ws echo server](#tcp-echo-server)
	- [kiss ws echo client](#tcp-echo-client)
- [Rpc Echo](#rpc-echo)
	- [rpc echo server](#tcp-echo-server)
	- [rpc echo client](#tcp-echo-client)
- [Http Echo](#http-echo)
	- [http server](#tcp-echo-server)

## 协议格式

- 协议格式，小端字节序

包体长度 | 命令号 | 扩展字段
---- | ---- | --------
4字节 |  4字节  | 8字节

1. 包体长度，4字节，有最大长度限制，TcpServer/TcpEngine可以通过SetSockMaxPackLen设置最大包长
2. 命令号，4字节，用户层协议最大命令号为0xFFFFFF，大于0xFFFFFF为net包保留协议号
   心跳包协议号： CmdPing = uint32(0x1 << 24)，无需包体
3. 扩展字段，在rpc时为rpc call的序号标识

## Tcp Echo

### tcp echo server

```golang
package main

import (
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/net"
	"time"
)

var ()

const (
	addr = ":8888"

	CMD_ECHO = uint32(1)
)

func onEcho(client *net.TcpClient, msg net.IMessage) {
	log.Info("tcp server onEcho from %v: %v", client.Conn.RemoteAddr().String(), string(msg.Body()))
	client.SendMsg(msg)
}

func main() {
	server := net.NewTcpServer("Echo")

	// 初始化协议号
	server.Handle(CMD_ECHO, onEcho)

	server.Serve(addr, time.Second*5)
}
```

### tcp echo client

```golang
package main

import (
	"fmt"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/net"
	"time"
)

var (
	addr = "127.0.0.1:8888"

	CMD_ECHO = uint32(1)
)

func onConnected(c *net.TcpClient) {
	log.Info("TcpClient OnConnected")
}

func onEcho(client *net.TcpClient, msg net.IMessage) {
	log.Debug("tcp client onEcho from %v: %v", client.Conn.RemoteAddr().String(), string(msg.Body()))
}

func main() {
	autoReconn := true
	netengine := net.NewTcpEngine()

	// 初始化协议号
	netengine.Handle(CMD_ECHO, onEcho)

	client, err := net.NewTcpClient(addr, netengine, nil, autoReconn, onConnected)
	if err != nil {
		log.Panic("NewTcpClient failed: %v, %v", client, err)
	}

	for i := 0; true; i++ {
		err = client.SendMsg(net.NewMessage(CMD_ECHO, []byte(fmt.Sprintf("hello %v", i))))
		if err != nil {
			log.Error("tcp client echo failed: %v", err)
		}
		time.Sleep(time.Second)
	}
}
```



## Websocket Echo

### 用户自定义消息处理

### raw ws echo server

```golang
package main

import (
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/net"
)

var (
	addr = ":8888"

	CMD_ECHO = uint32(1)
)

func onEcho(client *net.WSClient, msg net.IMessage) {
	log.Info("ws server onEcho from %v: %v", client.Conn.RemoteAddr().String(), string(msg.Data()))
	client.SendMsg(msg)
}

func main() {
	server, err := net.NewWebsocketServer("echo", addr)
	if err != nil {
		log.Panic("NewWebsocketServer failed: %v", err)
	}

	// 初始化http ws路由
	server.HandleWs("/ws/echo")

	// 设置消息处理接口
	server.HandleMessage(onEcho)

	server.Serve()
}
```

### raw ws echo client

```golang
package main

import (
	"fmt"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/net"
	"time"
)

var (
	addr = "ws://localhost:8888/ws/echo"
)

func onMessage(client *net.WSClient, msg net.IMessage) {
	log.Debug("ws client onEcho from %v: %v", client.Conn.RemoteAddr().String(), string(msg.Data()))
}

func main() {
	client, err := net.NewWebsocketClient(addr)
	if err != nil {
		log.Panic("NewWebsocketClient failed: %v, %v", err, time.Now())
	}

	// 设置消息处理接口
	client.HandleMessage(onMessage)

	for i := 0; true; i++ {
		err = client.SendMsg(net.RawMessage([]byte(fmt.Sprintf("hello %v", i))))
		if err != nil {
			log.Error("ws client echo failed: %v", err)
			break
		}
		time.Sleep(time.Second)
	}
}
```


### 按KISS的协议格式处理消息

### kiss ws echo server

```golang
package main

import (
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/net"
)

var (
	addr = ":8888"

	CMD_ECHO = uint32(1)
)

func onEcho(client *net.WSClient, msg net.IMessage) {
	log.Info("ws server onEcho from %v: %v", client.Conn.RemoteAddr().String(), string(msg.Body()))
	client.SendMsg(msg)
}

func main() {
	server, err := net.NewWebsocketServer("echo", addr)
	if err != nil {
		log.Panic("NewWebsocketServer failed: %v", err)
	}

	// 初始化http ws路由
	server.HandleWs("/ws/echo")

	// 初始化协议号
	server.Handle(CMD_ECHO, onEcho)

	server.Serve()
}
```

### kiss ws echo client

```golang
package main

import (
	"fmt"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/net"
	"time"
)

var (
	addr = "ws://localhost:8888/ws/echo"

	CMD_ECHO = uint32(1)
)

func onEcho(client *net.WSClient, msg net.IMessage) {
	log.Debug("ws client onEcho from %v: %v", client.Conn.RemoteAddr().String(), string(msg.Body()))
}

func main() {
	client, err := net.NewWebsocketClient(addr)
	if err != nil {
		log.Panic("NewWebsocketClient failed: %v, %v", err, time.Now())
	}

	// 初始化协议号
	client.Handle(CMD_ECHO, onEcho)

	for i := 0; true; i++ {
		err = client.SendMsg(net.NewMessage(CMD_ECHO, []byte(fmt.Sprintf("hello %v", i))))
		if err != nil {
			log.Error("ws client echo failed: %v", err)
			break
		}
		time.Sleep(time.Second)
	}
}

```



## Rpc Echo

### rpc server

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

### rpc client

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



## Http Echo

### http server

```golang
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/net"
	"net/http"
	"time"
)

// http://localhost:8080/hello
func main() {
	addr := ":8080"

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.GET("/hello", func(c *gin.Context) {
		log.Info("onHello")
		c.String(http.StatusOK, "hello")
	})

	net.ServeHttp("Hello", addr, router, time.Second*5, nil)
}
```