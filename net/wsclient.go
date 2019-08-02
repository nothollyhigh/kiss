package net

import (
	"crypto/tls"
	"github.com/gorilla/websocket"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// websocket client
type WSClient struct {
	*WSEngine
	sync.RWMutex

	// websocket connection
	Conn *websocket.Conn

	// real ip
	realIp string

	// send queue
	chSend chan wsAsyncMessage

	// running flag
	running bool

	// cipher
	cipher ICipher

	// recv packet sequence
	recvSeq int64

	// send packet sequence
	sendSeq int64

	// pre recv packet key
	recvKey uint32

	// pre send packet key
	sendKey uint32

	// user data
	userdata interface{}

	// client close callbacks
	onCloseMap map[interface{}]func(*WSClient)
}

// read loop
func (cli *WSClient) readloop() {
	defer util.HandlePanic()
	defer cli.Stop()

	var imsg IMessage
	for {
		if imsg = cli.WSEngine.RecvMsg(cli); imsg == nil {
			break
		}
		atomic.AddInt64(&cli.recvSeq, 1)
		cli.WSEngine.onMessage(cli, imsg)
	}
}

// write loop
func (cli *WSClient) writeloop() {
	defer cli.Stop()
	defer util.HandlePanic()

	var err error
	for msg := range cli.chSend {
		err = cli.WSEngine.Send(cli, msg.data)
		if msg.cb != nil {
			msg.cb(cli, err)
		}
		if err != nil {
			break
		}

		atomic.AddInt64(&cli.sendSeq, 1)
	}
}

// keepalive
func (cli *WSClient) Keepalive(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	msg := PingMsg()
	for {
		<-ticker.C

		cli.Lock()
		running := cli.running
		cli.Unlock()

		if !running {
			return
		}

		cli.SendMsg(msg)
	}
}

// receive sequence
func (cli *WSClient) RecvSeq() int64 {
	return atomic.LoadInt64(&cli.recvSeq)
}

// send sequence
func (cli *WSClient) SendSeq() int64 {
	return atomic.LoadInt64(&cli.sendSeq)
}

// receive key
func (cli *WSClient) RecvKey() uint32 {
	return cli.recvKey
}

// send key
func (cli *WSClient) SendKey() uint32 {
	return cli.sendKey
}

// cipher
func (cli *WSClient) Cipher() ICipher {
	return cli.cipher
}

// setting cipher
func (cli *WSClient) SetCipher(cipher ICipher) {
	cli.cipher = cipher
}

// user data
func (cli *WSClient) UserData() interface{} {
	return cli.userdata
}

// setting user data
func (cli *WSClient) SetUserData(data interface{}) {
	cli.userdata = data
}

// ip
func (cli *WSClient) Ip() string {
	if cli.realIp != "" {
		return cli.realIp
	}
	if cli.Conn != nil {
		addr := cli.Conn.RemoteAddr().String()
		if pos := strings.LastIndex(addr, ":"); pos > 0 {
			return addr[:pos]
		}
	}
	return "0.0.0.0"
}

// port
func (cli *WSClient) Port() int {
	if cli.Conn != nil {
		addr := cli.Conn.RemoteAddr().String()
		if pos := strings.LastIndex(addr, ":"); pos > 0 {
			if port, err := strconv.Atoi(addr[pos+1:]); err == nil {
				return port
			}
		}
	}
	return 0
}

// setting real ip
func (cli *WSClient) SetRealIp(ip string) {
	cli.realIp = ip
}

// bind data
func (cli *WSClient) Bind(data []byte, v interface{}) error {
	if cli.Codec == nil {
		return ErrClientWithoutCodec
	}
	return cli.Codec.Unmarshal(data, v)
}

// send message
func (cli *WSClient) SendMsg(msg IMessage) error {
	var err error = nil
	cli.Lock()
	if cli.running {
		select {
		case cli.chSend <- wsAsyncMessage{msg.Encrypt(cli.SendSeq(), cli.SendKey(), cli.cipher), nil}:
			cli.Unlock()
		default:
			cli.Unlock()
			cli.OnSendQueueFull(cli, msg)
			err = ErrWSClientSendQueueIsFull
		}
	} else {
		cli.Unlock()
		err = ErrWSClientIsStopped
	}
	if err != nil {
		log.Debug("[Websocket] SendMsg -> %v failed: %v", cli.Ip(), err)
	}

	return err
}

// send message with callback
func (cli *WSClient) SendMsgWithCallback(msg IMessage, cb func(*WSClient, error)) error {
	var err error = nil
	cli.Lock()
	if cli.running {
		select {
		case cli.chSend <- wsAsyncMessage{msg.Encrypt(cli.SendSeq(), cli.SendKey(), cli.cipher), cb}:
			cli.Unlock()
		default:
			cli.Unlock()
			cli.OnSendQueueFull(cli, msg)
			err = ErrTcpClientSendQueueIsFull
		}
	} else {
		cli.Unlock()
		err = ErrWSClientIsStopped
	}
	if err != nil {
		log.Debug("SendMsgWithCallback -> %v failed: %v", cli.Ip(), err)
	}

	return err
}

// send data
func (cli *WSClient) SendData(data []byte) error {
	var err error = nil
	cli.Lock()
	if cli.running {
		select {
		case cli.chSend <- wsAsyncMessage{data, nil}:
			cli.Unlock()
		default:
			cli.Unlock()
			cli.OnSendQueueFull(cli, data)
			err = ErrTcpClientSendQueueIsFull
		}
	} else {
		cli.Unlock()
		err = ErrWSClientIsStopped
	}
	if err != nil {
		log.Debug("SendData -> %v failed: %v", cli.Ip(), err)
	}

	return err
}

// send data with callback
func (cli *WSClient) SendDataWithCallback(data []byte, cb func(*WSClient, error)) error {
	var err error = nil
	cli.Lock()
	if cli.running {
		select {
		case cli.chSend <- wsAsyncMessage{data, cb}:
			cli.Unlock()
		default:
			cli.Unlock()
			cli.OnSendQueueFull(cli, data)
			err = ErrWSClientSendQueueIsFull
		}
	} else {
		cli.Unlock()
		err = ErrWSClientIsStopped
	}
	if err != nil {
		log.Debug("SendDataWithCallback -> %v failed: %v", cli.Ip(), err)
	}

	return err
}

// Stop
func (cli *WSClient) Stop() {
	cli.Lock()
	running := cli.running
	if running {
		cli.running = false
		cli.Conn.Close()
		close(cli.chSend)
	}
	cli.Unlock()
	if running {
		cli.RLock()
		for _, cb := range cli.onCloseMap {
			cb(cli)
		}
		cli.RUnlock()
	}
}

// setting close handler
func (cli *WSClient) OnClose(tag interface{}, cb func(client *WSClient)) {
	cli.Lock()
	cli.onCloseMap[tag] = cb
	cli.Unlock()
}

// unsetting close handler
func (cli *WSClient) CancelOnClose(tag interface{}) {
	cli.Lock()
	delete(cli.onCloseMap, tag)
	cli.Unlock()
}

// default create websocket client by websocket server
func newClient(conn *websocket.Conn, engine *WSEngine) *WSClient {
	sendQSize := DefaultSendQSize
	if engine != nil && engine.SendQSize > 0 {
		sendQSize = engine.SendQSize
	}

	cipher := engine.NewCipher()
	cli := &WSClient{
		WSEngine:   engine,
		Conn:       conn,
		chSend:     make(chan wsAsyncMessage, sendQSize),
		running:    true,
		cipher:     cipher,
		onCloseMap: map[interface{}]func(*WSClient){},
	}

	addr := conn.RemoteAddr().String()
	if pos := strings.LastIndex(addr, ":"); pos > 0 {
		cli.realIp = addr[:pos]
	}

	return cli
}

// websocket client factory
func NewWebsocketClient(addr string) (*WSClient, error) {
	dialer := &websocket.Dialer{}
	dialer.TLSClientConfig = &tls.Config{}
	conn, _, err := dialer.Dial(addr, nil)

	if err != nil {
		return nil, err
	}

	cli := newClient(conn, NewWebsocketEngine())

	util.Go(cli.readloop)
	util.Go(cli.writeloop)

	return cli, nil
}

func NewWebsocketTLSClient(addr string) (*WSClient, error) {
	dialer := &websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//dialer.TLSClientConfig.InsecureSkipVerify = true
	conn, _, err := dialer.Dial(addr, nil)
	if err != nil {
		return nil, err
	}

	cli := newClient(conn, NewWebsocketEngine())

	util.Go(cli.readloop)
	util.Go(cli.writeloop)

	return cli, nil
}
