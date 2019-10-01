package net

import (
	"fmt"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"net"
	"os"
	"sync/atomic"
	"syscall"
	"time"
)

// tcp server
type TcpServer struct {
	*TcpEngin
	tag           string
	enableBroad   bool
	addr          string
	accepted      int64
	currLoad      int64
	maxLoad       int64
	listener      *net.TCPListener
	stopTimeout   time.Duration
	onStopTimeout func()
	onStopHandler func(server *TcpServer)
}

// add client
func (server *TcpServer) addClient(client *TcpClient) {
	if server.enableBroad {
		server.Lock()
		server.clients[client] = struct{}{}
		server.Unlock()

	}
	atomic.AddInt64(&server.currLoad, 1)
	server.OnNewClient(client)
}

// delete client
func (server *TcpServer) deleClient(client *TcpClient) {
	if server.enableBroad {
		server.Lock()
		delete(server.clients, client)
		server.Unlock()
	}
	atomic.AddInt64(&server.currLoad, -1)
}

// stop all clients
func (server *TcpServer) stopClients() {
	server.Lock()
	defer server.Unlock()

	for client, _ := range server.clients {
		client.Stop()
	}
}

// listener loop
func (server *TcpServer) listenerLoop() error {
	log.Debug("[TcpServer %s] Running on: \"%s\"", server.tag, server.addr)
	defer log.Debug("[TcpServer %s] Stopped.", server.tag)

	var (
		err       error
		conn      *net.TCPConn
		client    *TcpClient
		tempDelay time.Duration
	)
	for server.running {
		if conn, err = server.listener.AcceptTCP(); err == nil {
			if server.maxLoad == 0 || atomic.LoadInt64(&server.currLoad) < server.maxLoad {
				// if runtime.GOOS == "linux" {
				// conn.File() cause block mod and create new os thread for socket, then beyond max thread num
				// 	if file, err = conn.File(); err == nil {
				// 		idx = uint64(file.Fd())
				// 	}
				// } else {
				// 	idx = server.accepted
				// 	server.accepted++
				// }
				server.accepted++

				if err = server.OnNewConn(conn); err == nil {
					client = server.CreateClient(conn, server.TcpEngin, server.NewCipher())
					client.start()
					server.addClient(client)
				} else {
					log.Debug("[TcpServer %s] init conn error: %v\n", server.tag, err)
				}
			} else {
				conn.Close()
			}
		} else {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Debug("[TcpServer %s] Accept error: %v; retrying in %v", server.tag, err, tempDelay)
				time.Sleep(tempDelay)
			} else {
				log.Debug("[TcpServer %s] Accept error: %v", server.tag, err)
				if server.onStopHandler != nil {
					server.onStopHandler(server)
				}
				break
			}
		}
	}

	return err
}

// start
func (server *TcpServer) Start(addr string) error {
	server.Lock()
	running := server.running
	server.running = true
	server.Unlock()

	if !running {
		server.Add(1)

		tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			log.Fatal("[TcpServer %s] ResolveTCPAddr error: %v", server.tag, err)
			return err
		}

		server.listener, err = net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			log.Fatal("[TcpServer %s] Listening error: %v", server.tag, err)
			return err
		}

		server.addr = addr
		defer server.listener.Close()

		return server.listenerLoop()
	}
	return fmt.Errorf("server already started")
}

// stop
func (server *TcpServer) Stop() {
	server.Lock()
	running := server.running
	server.running = false
	server.Unlock()
	defer util.HandlePanic()

	if !running {
		return
	}

	server.listener.Close()
	server.Done()

	if server.stopTimeout > 0 {
		time.AfterFunc(server.stopTimeout, func() {
			log.Debug("[TcpServer %s] Stop Timeout.", server.tag)
			if server.onStopTimeout != nil {
				server.onStopTimeout()
			}
		})
	}

	log.Debug("[TcpServer %s] Stop Waiting...", server.tag)

	server.Wait()

	server.stopClients()

	if server.onStopHandler != nil {
		server.onStopHandler(server)
	}
	log.Debug("[TcpServer %s] Stop Done.", server.tag)
}

// stop with timeout
func (server *TcpServer) StopWithTimeout(stopTimeout time.Duration, onStopTimeout func()) {
	server.stopTimeout = stopTimeout
	server.onStopTimeout = onStopTimeout
	server.Stop()
}

// serve
func (server *TcpServer) Serve(addr string, stopTimeout time.Duration) {
	util.Go(func() {
		server.Start(addr)
	})

	server.stopTimeout = stopTimeout
	server.onStopTimeout = func() {
		os.Exit(0)
	}

	util.HandleSignal(func(sig os.Signal) {
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			server.Stop()
			os.Exit(0)
		}
	})
}

// currunt load
func (server *TcpServer) CurrLoad() int64 {
	return atomic.LoadInt64(&server.currLoad)
}

// max load
func (server *TcpServer) MaxLoad() int64 {
	return server.maxLoad
}

// set max concurrent
func (server *TcpServer) SetMaxConcurrent(maxLoad int64) {
	server.maxLoad = maxLoad
}

// total accept num
func (server *TcpServer) AcceptedNum() int64 {
	return server.accepted
}

// setting server stop handler
func (server *TcpServer) HandleServerStop(stopHandler func(server *TcpServer)) {
	server.onStopHandler = stopHandler
}

// enable broadcast
func (server *TcpServer) EnableBroadcast() {
	server.enableBroad = true
}

// broadcast
func (server *TcpServer) Broadcast(msg IMessage) {
	if !server.enableBroad {
		panic(ErrorBroadcastNotEnabled)
	}
	server.Lock()
	for c, _ := range server.clients {
		c.SendMsg(msg)
	}
	server.Unlock()
}

// broadcast with filter
func (server *TcpServer) BroadcastWithFilter(msg IMessage, filter func(*TcpClient) bool) {
	if !server.enableBroad {
		panic(ErrorBroadcastNotEnabled)
	}
	server.Lock()
	for c, _ := range server.clients {
		if filter(c) {
			c.SendMsg(msg)
		}
	}
	server.Unlock()
}

// tcp server factory
func NewTcpServer(tag string) *TcpServer {
	server := &TcpServer{
		TcpEngin: &TcpEngin{
			clients: map[*TcpClient]struct{}{},
			handlers: map[uint32]func(*TcpClient, IMessage){
				CmdSetReaIp: func(client *TcpClient, msg IMessage) {
					ip := msg.Body()
					client.SetRealIp(string(ip))
				},
			},

			sockNoDelay:            DefaultSockNodelay,
			sockKeepAlive:          DefaultSockKeepalive,
			sockBufioReaderEnabled: DefaultSockBufioReaderEnabled,
			sendQsize:              DefaultSendQSize,
			sockRecvBufLen:         DefaultSockRecvBufLen,
			sockSendBufLen:         DefaultSockSendBufLen,
			sockMaxPackLen:         DefaultSockPackMaxLen,
			sockRecvBlockTime:      DefaultSockRecvBlockTime,
			sockSendBlockTime:      DefaultSockSendBlockTime,
			sockKeepaliveTime:      DefaultSockKeepaliveTime,
		},
		maxLoad: DefaultMaxOnline,
		tag:     tag,
	}

	cipher := NewCipherGzip(DefaultThreshold)
	server.HandleNewCipher(func() ICipher {
		return cipher
	})

	server.HandleDisconnected(server.deleClient)

	return server
}

// rpc server factory
func NewRpcServer(tag string) *TcpServer {
	return NewTcpServer(tag)
}
