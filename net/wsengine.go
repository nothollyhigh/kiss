package net

import (
	// "encoding/binary"
	"github.com/gorilla/websocket"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"sync"
	"time"
)

// websocket engine
type WSEngine struct {
	sync.Mutex
	sync.WaitGroup

	// 序列化
	Codec ICodec

	// 读超时时间
	ReadTimeout time.Duration

	// 写超时时间
	WriteTimeout time.Duration

	// 读最大包长限制
	ReadLimit int64

	// 发送队列容量
	SendQSize int

	// websocket消息类型
	MessageType int

	// shutdown flag
	shutdown bool

	//ctypto cipher
	cipher ICipher

	// message handlers
	handlers map[uint32]func(cli *WSClient, msg IMessage)

	// user defined message handler
	messageHandler func(cli *WSClient, msg IMessage)

	// receive message handler
	recvHandler func(cli *WSClient) IMessage

	// send message handler
	sendHandler func(cli *WSClient, data []byte) error

	// send queue full handler
	sendQueueFullHandler func(cli *WSClient, msg interface{})

	// new cipher handler
	newCipherHandler func() ICipher
}

// receive message
func (engine *WSEngine) RecvMsg(cli *WSClient) IMessage {
	if engine.recvHandler != nil {
		return engine.recvHandler(cli)
	}

	var err error
	var data []byte

	if cli.ReadTimeout > 0 {
		err = cli.Conn.SetReadDeadline(time.Now().Add(cli.ReadTimeout))
		if err != nil {
			log.Debug("Websocket SetReadDeadline failed: %v", err)
			return nil
		}
	}

	_, data, err = cli.Conn.ReadMessage()
	if err != nil {
		log.Debug("Websocket ReadMessage failed: %v", err)
		return nil
	}

	msg := &Message{
		rawData: data,
		data:    nil,
	}

	if _, err = msg.Decrypt(cli.RecvSeq(), cli.RecvKey(), cli.Cipher()); err != nil {
		log.Debug("%s RecvMsg Decrypt Err: %v", cli.Conn.RemoteAddr().String(), err)
		return nil
	}

	return msg
}

// send websocket data
func (engine *WSEngine) Send(cli *WSClient, data []byte) error {
	defer util.HandlePanic()

	if engine.sendHandler != nil {
		return engine.sendHandler(cli, data)
	}

	var err error

	if cli.WriteTimeout > 0 {
		err = cli.Conn.SetWriteDeadline(time.Now().Add(engine.WriteTimeout))
		if err != nil {
			log.Debug("%s Send SetReadDeadline Err: %v", cli.Conn.RemoteAddr().String(), err)
			cli.Stop()
			return err
		}
	}

	err = cli.Conn.WriteMessage(engine.MessageType, data)
	if err != nil {
		log.Debug("%s Send Write Err: %v", cli.Conn.RemoteAddr().String(), err)
		cli.Stop()
	}

	return err
}

// setting user defined message handler
func (engine *WSEngine) HandleMessage(h func(cli *WSClient, msg IMessage)) {
	engine.messageHandler = h
}

// setting message handler by cmd
func (engine *WSEngine) Handle(cmd uint32, h func(cli *WSClient, msg IMessage)) {
	if cmd == CmdPing {
		log.Panic(ErrorReservedCmdPing.Error())
	}
	if cmd == CmdSetReaIp {
		log.Panic(ErrorReservedCmdSetRealip.Error())
	}
	if cmd == CmdRpcMethod {
		log.Panic(ErrorReservedCmdRpcMethod.Error())
	}
	if cmd == CmdRpcError {
		log.Panic(ErrorReservedCmdRpcError.Error())
	}
	if cmd > CmdUserMax {
		log.Panic(ErrorReservedCmdInternal.Error())
	}
	if _, ok := engine.handlers[cmd]; ok {
		log.Panic("Websocket Handle failed, cmd %v already exist", cmd)
	}
	engine.handlers[cmd] = h
}

// setting receive message handler
func (engine *WSEngine) HandleRecv(recver func(cli *WSClient) IMessage) {
	engine.recvHandler = recver
}

// setting send message handler
func (engine *WSEngine) HandleSend(sender func(cli *WSClient, data []byte) error) {
	engine.sendHandler = sender
}

// handle send queue full
func (engine *WSEngine) OnSendQueueFull(cli *WSClient, msg interface{}) {
	if engine.sendQueueFullHandler != nil {
		engine.sendQueueFullHandler(cli, msg)
	}
}

// setting send queue full handler
func (engine *WSEngine) HandleSendQueueFull(h func(cli *WSClient, msg interface{})) {
	engine.sendQueueFullHandler = h
}

// new cipher
func (engine *WSEngine) NewCipher() ICipher {
	if engine.newCipherHandler != nil {
		return engine.newCipherHandler()
	}
	return nil
}

// setting new cipher handler
func (engine *WSEngine) HandleNewCipher(newCipher func() ICipher) {
	engine.newCipherHandler = newCipher
}

// handle message
func (engine *WSEngine) onMessage(cli *WSClient, msg IMessage) {
	if engine.shutdown {
		// switch msg.Cmd() {
		// case CmdPing:
		// case CmdSetReaIp:
		// case CmdRpcMethod:
		// case CmdRpcError:
		// default:
		// 	log.Debug("engine is not running, ignore cmd %X, ip: %v", msg.Cmd(), cli.Ip())
		// 	return
		// }
		return
	}

	if engine.messageHandler != nil {
		engine.messageHandler(cli, msg)
		return
	}

	cmd := msg.Cmd()
	if cmd == CmdPing {
		cli.SendMsg(ping2Msg)
		return
	}
	if cmd == CmdPing2 {
		return
	}

	if h, ok := engine.handlers[cmd]; ok {
		engine.Add(1)
		defer engine.Done()
		defer util.HandlePanic()
		h(cli, msg)
	} else {
		log.Debug("Websocket no handler for cmd: %v", cmd)
	}
}

// websocket engine factory
func NewWebsocketEngine() *WSEngine {
	engine := &WSEngine{
		Codec:        DefaultCodec,
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
		ReadLimit:    DefaultReadLimit,
		SendQSize:    DefaultSendQSize,
		MessageType:  websocket.TextMessage,
		shutdown:     false,
		handlers: map[uint32]func(*WSClient, IMessage){
			CmdSetReaIp: func(cli *WSClient, msg IMessage) {
				ip := msg.Body()
				cli.SetRealIp(string(ip))
			},
		},
	}

	cipher := NewCipherGzip(DefaultThreshold)
	engine.HandleNewCipher(func() ICipher {
		return cipher
	})

	return engine
}
