package net

import (
	"errors"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"sync/atomic"
	"time"
)

// rpc message
type RpcMessage struct {
	msg IMessage
	err error
}

// rpc session
type rpcsession struct {
	seq  int64
	done chan *RpcMessage
}

// rpc client
type RpcClient struct {
	*TcpClient
	sessionMap map[int64]*rpcsession
	codec      ICodec
}

// remove rpc session
func (client *RpcClient) removeSession(seq int64) {
	client.Lock()
	delete(client.sessionMap, seq)
	if len(client.sessionMap) == 0 {
		client.sessionMap = map[int64]*rpcsession{}
	}
	client.Unlock()
}

// call cmd
func (client *RpcClient) callCmd(cmd uint32, data []byte) ([]byte, error) {
	var session *rpcsession
	client.Lock()
	if client.running {
		session = &rpcsession{
			seq:  atomic.AddInt64(&client.sendSeq, 1),
			done: make(chan *RpcMessage, 1),
		}
		msg := NewRpcMessage(cmd, session.seq, data)
		client.chSend <- asyncMessage{msg.data, nil}
		client.sessionMap[session.seq] = session
	} else {
		client.Unlock()
		return nil, ErrRpcClientIsDisconnected
	}
	client.Unlock()
	defer client.removeSession(session.seq)
	msg, ok := <-session.done
	if !ok {
		return nil, ErrRpcClientIsDisconnected
	}
	return msg.msg.Body(), msg.err
}

// call cmd with timeout
func (client *RpcClient) callCmdWithTimeout(cmd uint32, data []byte, timeout time.Duration) ([]byte, error) {
	var session *rpcsession
	client.Lock()
	if !client.running {
		client.Unlock()
		return nil, ErrRpcClientIsDisconnected
	}

	// after := time.After(timeout)
	after := time.NewTimer(timeout)
	defer after.Stop()

	session = &rpcsession{
		seq:  atomic.AddInt64(&client.sendSeq, 1),
		done: make(chan *RpcMessage, 1),
	}
	msg := NewRpcMessage(cmd, session.seq, data)
	select {
	case client.chSend <- asyncMessage{msg.data, nil}:
		client.sessionMap[session.seq] = session
	case <-after.C:
		client.Unlock()
		return nil, ErrRpcCallTimeout
	}

	client.Unlock()
	defer client.removeSession(session.seq)
	select {
	case msg, ok := <-session.done:
		if !ok {
			return nil, ErrRpcClientIsDisconnected
		}
		return msg.msg.Body(), msg.err
	case <-after.C:
		return nil, ErrRpcCallTimeout
	}
	return nil, ErrRpcCallClientError
}

func (client *RpcClient) callCmdWithTimer(cmd uint32, data []byte, after *time.Timer) ([]byte, error) {
	var session *rpcsession
	client.Lock()
	if !client.running {
		client.Unlock()
		return nil, ErrRpcClientIsDisconnected
	}

	session = &rpcsession{
		seq:  atomic.AddInt64(&client.sendSeq, 1),
		done: make(chan *RpcMessage, 1),
	}
	msg := NewRpcMessage(cmd, session.seq, data)
	select {
	case client.chSend <- asyncMessage{msg.data, nil}:
		client.sessionMap[session.seq] = session
	case <-after.C:
		client.Unlock()
		return nil, ErrRpcCallTimeout
	}

	client.Unlock()
	defer client.removeSession(session.seq)
	select {
	case msg, ok := <-session.done:
		if !ok {
			return nil, ErrRpcClientIsDisconnected
		}
		return msg.msg.Body(), msg.err
	case <-after.C:
		return nil, ErrRpcCallTimeout
	}
	return nil, ErrRpcCallClientError
}

// codec
func (client *RpcClient) Codec() ICodec {
	return client.codec
}

// call cmd
func (client *RpcClient) CallCmd(cmd uint32, req interface{}, rsp interface{}) error {
	data, err := client.codec.Marshal(req)
	if err != nil {
		return err
	}
	rspdata, err := client.callCmd(cmd, data)
	if err != nil {
		return err
	}
	if rsp != nil {
		err = client.codec.Unmarshal(rspdata, rsp)
	}
	return err
}

// call cmd with timeout
func (client *RpcClient) CallCmdWithTimeout(cmd uint32, req interface{}, rsp interface{}, timeout time.Duration) error {
	data, err := client.codec.Marshal(req)
	if err != nil {
		return err
	}
	rspdata, err := client.callCmdWithTimeout(cmd, data, timeout)
	if err != nil {
		return err
	}
	if rsp != nil {
		err = client.codec.Unmarshal(rspdata, rsp)
	}
	return err
}

// rpc call
func (client *RpcClient) Call(method string, req interface{}, rsp interface{}, timeout time.Duration) error {
	data, err := client.codec.Marshal(req)
	if err != nil {
		return err
	}
	data = append(data, make([]byte, len(method)+1)...)
	copy(data[len(data)-len(method)-1:], method)
	data[len(data)-1] = byte(len(method))
	rspdata, err := client.callCmdWithTimeout(CmdRpcMethod, data, timeout)
	if err != nil {
		return err
	}
	if rsp != nil {
		err = client.codec.Unmarshal(rspdata, rsp)
	}
	return err
}

// rpc call
func (client *RpcClient) CallWithTimer(method string, req interface{}, rsp interface{}, after *time.Timer) error {
	data, err := client.codec.Marshal(req)
	if err != nil {
		return err
	}
	data = append(data, make([]byte, len(method)+1)...)
	copy(data[len(data)-len(method)-1:], method)
	data[len(data)-1] = byte(len(method))
	rspdata, err := client.callCmdWithTimer(CmdRpcMethod, data, after)
	if err != nil {
		return err
	}
	if rsp != nil {
		err = client.codec.Unmarshal(rspdata, rsp)
	}
	return err
}

// func Upgrade(client *TcpClient, codec ICodec) *RpcClient {
// 	if codec == nil {
// 		codec = DefaultCodec
// 		log.Debug("use default rpc codec: %v", DefaultRpcCodecType)
// 	}

// 	rpcclient := &RpcClient{TcpClient: client, sessionMap: map[int64]*rpcsession{}, codec: codec}

// 	rpcclient.OnClose("-", func(*TcpClient) {
// 		rpcclient.Lock()
// 		defer rpcclient.Unlock()
// 		for _, session := range rpcclient.sessionMap {
// 			close(session.done)
// 		}
// 		rpcclient.sessionMap = map[int64]*rpcsession{}
// 	})

// 	engine := client.parent
// 	engine.rpcclients[client] = rpcclient
// 	engine.HandleMessage(func(c *TcpClient, msg IMessage) {
// 		switch msg.Cmd() {
// 		case CmdPing2:
// 		case CmdRpcMethod:
// 			rc := engine.rpcclients[c]
// 			rc.Lock()
// 			session, ok := rc.sessionMap[msg.Ext()]
// 			rc.Unlock()
// 			if ok {
// 				session.done <- &RpcMessage{msg, nil}
// 			} else {
// 				log.Debug("no rpcsession waiting for rpc response seq: %v", msg.Ext())
// 			}
// 		case CmdRpcError:
// 			rc := engine.rpcclients[c]
// 			rc.Lock()
// 			session, ok := rc.sessionMap[msg.Ext()]
// 			rc.Unlock()
// 			if ok {
// 				session.done <- &RpcMessage{msg, errors.New(string(msg.Body()))}
// 			} else {
// 				log.Debug("no rpcsession waiting for rpc response, cmd %v, ip: %v", msg.Cmd(), c.Ip())
// 			}
// 		default:
// 			if handler, ok := engine.handlers[msg.Cmd()]; ok {
// 				engine.Add(1)
// 				defer engine.Done()
// 				defer util.HandlePanic()
// 				handler(c, msg)
// 			} else {
// 				log.Debug("no handler for cmd %v", msg.Cmd())
// 			}
// 		}
// 	})

// 	return rpcclient
// }

// rpc client factory
func NewRpcClient(addr string, engine *TcpEngin, codec ICodec, onConnected func(*RpcClient)) (*RpcClient, error) {
	if engine == nil {
		engine = NewTcpEngine()
		engine.SetSendQueueSize(DefaultSockRpcSendQSize)
		engine.SetSockRecvBlockTime(DefaultSockRpcRecvBlockTime)
	}

	tmp := *engine
	engine = &tmp

	if codec == nil {
		codec = DefaultCodec
		log.Debug("use default rpc codec: %v", DefaultRpcCodecType)
	}

	var err error
	rpcclient := &RpcClient{sessionMap: map[int64]*rpcsession{}, codec: codec}

	rpcclient.TcpClient, err = newTcpClient(addr, engine, NewCipherGzip(DefaultThreshold), true, func(c *TcpClient) {
		if onConnected != nil {
			onConnected(rpcclient)
		}
	})

	if err != nil {
		return nil, err
	}

	// engine.rpcclients[rpcclient.TcpClient] = rpcclient

	if onConnected != nil {
		onConnected(rpcclient)
	}

	util.Go(func() {
		rpcclient.TcpClient.Keepalive(engine.SockKeepaliveTime())
	})

	rpcclient.OnClose("-", func(*TcpClient) {
		rpcclient.Lock()
		defer rpcclient.Unlock()
		for _, session := range rpcclient.sessionMap {
			close(session.done)
		}
		rpcclient.sessionMap = map[int64]*rpcsession{}
	})

	engine.HandleMessage(func(c *TcpClient, msg IMessage) {
		switch msg.Cmd() {
		case CmdPing2:
		case CmdRpcMethod:
			rpcclient.Lock()
			session, ok := rpcclient.sessionMap[msg.Ext()]
			rpcclient.Unlock()
			if ok {
				session.done <- &RpcMessage{msg, nil}
			} else {
				log.Debug("no rpcsession waiting for rpc response seq: %v", msg.Ext())
			}
		case CmdRpcError:
			rpcclient.Lock()
			session, ok := rpcclient.sessionMap[msg.Ext()]
			rpcclient.Unlock()
			if ok {
				session.done <- &RpcMessage{msg, errors.New(string(msg.Body()))}
			} else {
				log.Debug("no rpcsession waiting for rpc response, cmd %v, ip: %v", msg.Cmd(), c.Ip())
			}
		default:
			if handler, ok := engine.handlers[msg.Cmd()]; ok {
				engine.Add(1)
				defer engine.Done()
				defer util.HandlePanic()
				handler(c, msg)
			} else {
				log.Debug("no handler for cmd %v", msg.Cmd())
			}
		}
	})

	return rpcclient, nil
}
