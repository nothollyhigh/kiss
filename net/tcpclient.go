package net

import (
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// tcp client
type TcpClient struct {
	sync.RWMutex

	// tcp connection
	Conn *net.TCPConn

	// tcp engine parent
	parent *TcpEngin

	// recv packet sequence
	recvSeq int64

	// send packet sequence
	sendSeq int64

	// pre recv packet key
	recvKey uint32

	// pre send packet key
	sendKey uint32

	// ctypto cipher
	cipher ICipher

	// chan for message send queue
	chSend chan asyncMessage

	// client close callbacks
	onCloseMap map[interface{}]func(*TcpClient)

	// user data
	userdata interface{}

	// real ip
	realIp string

	// running flag
	running bool

	// shutdown flag
	shutdown bool
}

// ip
func (client *TcpClient) Ip() string {
	if client.realIp != "" {
		return client.realIp
	}
	if client.Conn != nil {
		addr := client.Conn.RemoteAddr().String()
		if pos := strings.LastIndex(addr, ":"); pos > 0 {
			return addr[:pos]
		}
	}
	return "0.0.0.0"
}

// port
func (client *TcpClient) Port() int {
	if client.Conn != nil {
		addr := client.Conn.RemoteAddr().String()
		if pos := strings.LastIndex(addr, ":"); pos > 0 {
			if port, err := strconv.Atoi(addr[pos+1:]); err == nil {
				return port
			}
		}
	}
	return 0
}

// set real ip
func (client *TcpClient) SetRealIp(ip string) {
	client.realIp = ip
}

// bind data
func (client *TcpClient) Bind(data []byte, v interface{}) error {
	if client.parent.Codec == nil {
		return ErrClientWithoutCodec
	}
	return client.parent.Codec.Unmarshal(data, v)
}

// setting close handler
func (client *TcpClient) OnClose(tag interface{}, cb func(client *TcpClient)) {
	client.Lock()
	if client.running {
		client.onCloseMap[tag] = cb
	}
	client.Unlock()
}

// unsetting close handler
func (client *TcpClient) CancelOnClose(tag interface{}) {
	client.Lock()
	if client.running {
		delete(client.onCloseMap, tag)
	}
	client.Unlock()
}

// send message
func (client *TcpClient) SendMsg(msg IMessage) error {
	var err error = nil
	client.Lock()
	if client.running {
		select {
		case client.chSend <- asyncMessage{msg.Encrypt(client.SendSeq(), client.SendKey(), client.cipher), nil}:
			client.Unlock()
		default:
			client.Unlock()
			client.parent.OnSendQueueFull(client, msg)
			err = ErrTcpClientSendQueueIsFull
		}
	} else {
		client.Unlock()
		err = ErrTcpClientIsStopped
	}
	if err != nil {
		log.Debug("SendMsg -> %v failed: %v", client.Ip(), err)
	}

	return err
}

// send message with callback
func (client *TcpClient) SendMsgWithCallback(msg IMessage, cb func(*TcpClient, error)) error {
	var err error = nil
	client.Lock()
	if client.running {
		select {
		case client.chSend <- asyncMessage{msg.Encrypt(client.SendSeq(), client.SendKey(), client.cipher), cb}:
			client.Unlock()
		default:
			client.Unlock()
			client.parent.OnSendQueueFull(client, msg)
			err = ErrTcpClientSendQueueIsFull
		}
	} else {
		client.Unlock()
		err = ErrTcpClientIsStopped
	}
	if err != nil {
		log.Debug("SendMsgWithCallback -> %v failed: %v", client.Ip(), err)
	}

	return err
}

// send data
func (client *TcpClient) SendData(data []byte) error {
	var err error = nil
	client.Lock()
	if client.running {
		select {
		case client.chSend <- asyncMessage{data, nil}:
			client.Unlock()
		default:
			client.Unlock()
			client.parent.OnSendQueueFull(client, data)
			err = ErrTcpClientSendQueueIsFull
		}
	} else {
		client.Unlock()
		err = ErrTcpClientIsStopped
	}
	if err != nil {
		log.Debug("SendData -> %v failed: %v", client.Ip(), err)
	}

	return err
}

// send data with callback
func (client *TcpClient) SendDataWithCallback(data []byte, cb func(*TcpClient, error)) error {
	var err error = nil
	client.Lock()
	if client.running {
		select {
		case client.chSend <- asyncMessage{data, cb}:
			client.Unlock()
		default:
			client.Unlock()
			client.parent.OnSendQueueFull(client, data)
			err = ErrTcpClientSendQueueIsFull
		}
	} else {
		client.Unlock()
		err = ErrTcpClientIsStopped
	}
	if err != nil {
		log.Debug("SendDataWithCallback -> %v failed: %v", client.Ip(), err)
	}

	return err
}

// push data sync, using for rpc
func (client *TcpClient) pushDataSync(data []byte) error {
	defer util.HandlePanic()
	var err error = nil
	client.Lock()
	if client.running {
		// client.Conn.SetWriteDeadline(time.Now().Add(client.parent.SockSendBlockTime()))
		// nwrite, err := client.Conn.Write(data)
		// if err != nil {
		// 	client.Unlock()
		// 	client.Stop()
		// 	return err
		// }
		// if nwrite != len(data) {
		// 	client.Unlock()
		// 	client.Stop()
		// 	return ErrTcpClientWriteHalf
		// }
		select {
		case client.chSend <- asyncMessage{data, nil}:
			client.Unlock()
		case <-time.After(client.parent.SockSendBlockTime()):
			client.Unlock()
			err = ErrRpcCallTimeout
		}
	} else {
		client.Unlock()
		err = ErrTcpClientIsStopped
	}
	if err != nil {
		log.Debug("pushDataSync -> %v failed: %v", client.Ip(), err)
	}

	return err
}

// receive sequence
func (client *TcpClient) RecvSeq() int64 {
	return atomic.LoadInt64(&client.recvSeq)
}

// send sequence
func (client *TcpClient) SendSeq() int64 {
	return atomic.LoadInt64(&client.sendSeq)
}

// receive key
func (client *TcpClient) RecvKey() uint32 {
	return client.recvKey
}

// send key
func (client *TcpClient) SendKey() uint32 {
	return client.sendKey
}

// cipher
func (client *TcpClient) Cipher() ICipher {
	return client.cipher
}

// setting cipher
func (client *TcpClient) SetCipher(cipher ICipher) {
	client.cipher = cipher
}

// user data
func (client *TcpClient) UserData() interface{} {
	return client.userdata
}

// setting user data
func (client *TcpClient) SetUserData(data interface{}) {
	client.userdata = data
}

// client start
func (client *TcpClient) start() {
	util.Go(client.readloop)
	util.Go(client.writeloop)
}

// client keepalive
func (client *TcpClient) Keepalive(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	msg := PingMsg()
	for {
		<-ticker.C

		client.Lock()
		shutdown := client.shutdown
		client.Unlock()

		if shutdown {
			return
		}
		client.SendMsg(msg)
	}
}

// restart for auto reconnect
func (client *TcpClient) restart(conn *net.TCPConn) {
	client.Lock()
	defer client.Unlock()
	if !client.running {
		client.running = true

		client.Conn = conn
		if client.cipher != nil {
			client.cipher.Init()
		}
		sendQsize := client.parent.SendQueueSize()
		if sendQsize <= 0 {
			sendQsize = DefaultSendQSize
		}
		client.chSend = make(chan asyncMessage, sendQsize)

		util.Go(client.writeloop)
		util.Go(client.readloop)
	}
}

// stop
func (client *TcpClient) stop() {
	defer util.HandlePanic()

	client.Lock()
	client.running = false
	client.Unlock()

	close(client.chSend)

	client.Conn.CloseRead()
	client.Conn.CloseWrite()
	client.Conn.Close()

	for _, cb := range client.onCloseMap {
		cb(client)
	}

	client.parent.OnDisconnected(client)
}

// Stop
func (client *TcpClient) Stop() error {
	defer util.HandlePanic()
	client.Lock()
	running := client.running
	client.running = false
	client.Unlock()
	if running {
		if client.Conn != nil {
			err := client.Conn.CloseRead()
			if err != nil {
				return err
			}
			return client.Conn.CloseWrite()
			//return client.Conn.Close()
		}
	}
	return ErrTcpClientIsStopped
}

// shutdown for auto reconnect client
func (client *TcpClient) Shutdown() error {
	client.Lock()
	client.shutdown = true
	client.Unlock()
	return client.Stop()
}

// send async message
// func (client *TcpClient) send(amsg *asyncMessage) error {
// 	client.RLock()
// 	running := client.running
// 	client.RUnlock()
// 	if !running {
// 		return ErrTcpClientIsStopped
// 	}
// 	defer util.HandlePanic()
// 	return client.parent.Send(client, amsg.data)
// }

func (client *TcpClient) writeloop() {
	defer client.Stop()

	var err error = nil
	for asyncMsg := range client.chSend {
		// err = client.send(&asyncMsg)
		err = client.parent.Send(client, asyncMsg.data)
		if asyncMsg.cb != nil {
			asyncMsg.cb(client, err)
		}
		if err != nil {
			break
		}
		atomic.AddInt64(&client.sendSeq, 1)
	}
}

// read loop
func (client *TcpClient) readloop() {
	defer client.stop()
	var imsg IMessage
	for {
		if imsg = client.parent.RecvMsg(client); imsg == nil {
			break
		}
		atomic.AddInt64(&client.recvSeq, 1)
		client.parent.OnMessage(client, imsg)
	}
}

// default create tcp client by tcp server
func createTcpClient(conn *net.TCPConn, parent *TcpEngin, cipher ICipher) *TcpClient {
	if parent == nil {
		parent = NewTcpEngine()
	}
	sendQsize := parent.SendQueueSize()
	if sendQsize <= 0 {
		sendQsize = DefaultSendQSize
	}

	conn.SetNoDelay(parent.SockNoDelay())
	conn.SetKeepAlive(parent.SockKeepAlive())
	if parent.SockKeepAlive() {
		conn.SetKeepAlivePeriod(parent.SockKeepaliveTime())
	}
	conn.SetReadBuffer(parent.SockRecvBufLen())
	conn.SetWriteBuffer(parent.SockSendBufLen())

	client := &TcpClient{
		Conn:       conn,
		parent:     parent,
		cipher:     cipher,
		chSend:     make(chan asyncMessage, sendQsize),
		onCloseMap: map[interface{}]func(*TcpClient){},
		running:    true,
	}

	addr := conn.RemoteAddr().String()
	if pos := strings.LastIndex(addr, ":"); pos > 0 {
		client.realIp = addr[:pos]
	}

	return client
}

// tcp client factory
func NewTcpClient(addr string, parent *TcpEngin, cipher ICipher, autoReconn bool, onConnected func(*TcpClient)) (*TcpClient, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Debug("NewTcpClient failed: ", err)
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Debug("NewTcpClient failed: %v", err)
		return nil, err
	}

	client := createTcpClient(conn, parent, cipher)
	client.start()

	if autoReconn {
		client.OnClose("reconn", func(*TcpClient) {
			util.Go(func() {
				times := 0
				tempDelay := time.Second / 10
				for !client.shutdown {
					times++
					time.Sleep(tempDelay)
					if conn, err := net.DialTCP("tcp", nil, tcpAddr); err == nil {
						client.Lock()
						defer client.Unlock()
						if !client.shutdown {
							log.Debug("TcpClient auto reconnect to %v %d success", addr, times)
							client.recvSeq = 0
							client.sendSeq = 0
							util.Go(func() {
								client.restart(conn)
								if onConnected != nil {
									onConnected(client)
								}
							})
						} else {
							conn.Close()
						}
						return
					} else {
						log.Debug("TcpClient auto reconnect to %v %d failed: %v", addr, times, err)
					}
					tempDelay *= 2
					if tempDelay > time.Second*2 {
						tempDelay = time.Second * 2
					}
				}
			})
		})
	}
	if onConnected != nil {
		util.Go(func() {
			onConnected(client)
		})
	}

	return client, nil
}
