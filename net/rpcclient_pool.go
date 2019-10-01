package net

// import (
// 	"errors"
// 	"github.com/nothollyhigh/kiss/log"
// 	"github.com/nothollyhigh/kiss/util"
// 	"sync/atomic"
// 	"time"
// )

// // rpc client pool
// type RpcClientPool struct {
// 	idx     int64
// 	size    uint64
// 	codec   ICodec
// 	clients []*RpcClient
// }

// // codec
// func (pool *RpcClientPool) Codec() ICodec {
// 	return pool.codec
// }

// // rpc client
// func (pool *RpcClientPool) Client() *RpcClient {
// 	idx := atomic.AddInt64(&pool.idx, 1)
// 	return pool.clients[uint64(idx)%pool.size]
// }

// // call
// func (pool *RpcClientPool) Call(method string, req interface{}, rsp interface{}, timeout time.Duration) error {
// 	idx := atomic.AddInt64(&pool.idx, 1)
// 	client := pool.clients[uint64(idx)%pool.size]
// 	err := client.Call(method, req, rsp, timeout)
// 	return err
// }

// // rpc client pool factory
// func NewRpcClientPool(addr string, engine *TcpEngin, codec ICodec, poolSize int, onConnected func(*RpcClient)) (*RpcClientPool, error) {
// 	if codec == nil {
// 		codec = DefaultCodec
// 		log.Debug("use default rpc codec: %v", DefaultRpcCodecType)
// 	}

// 	pool := &RpcClientPool{
// 		size:    uint64(poolSize),
// 		codec:   codec,
// 		clients: make([]*RpcClient, poolSize),
// 	}

// 	if engine == nil {
// 		engine = NewTcpEngine()
// 		engine.SetSendQueueSize(DefaultSockRpcSendQSize)
// 		engine.SetSockRecvBlockTime(DefaultSockRpcRecvBlockTime)
// 	}

// 	for i := 0; i < poolSize; i++ {
// 		tmp := *engine
// 		engine = &tmp

// 		var err error
// 		rpcclient := &RpcClient{sessionMap: map[int64]*rpcsession{}, codec: pool.codec}
// 		rpcclient.TcpClient, err = newTcpClient(addr, engine, NewCipherGzip(DefaultThreshold), true, func(c *TcpClient) {
// 			if onConnected != nil {
// 				onConnected(engine.rpcclients[c])
// 			}
// 		})
// 		if err != nil {
// 			return nil, err
// 		}

// 		if onConnected != nil {
// 			onConnected(rpcclient)
// 		}

// 		// clients := map[*TcpClient]*RpcClient{}
// 		engine.HandleMessage(func(c *TcpClient, msg IMessage) {
// 			switch msg.Cmd() {
// 			case CmdPing2:
// 			case CmdRpcMethod:
// 				rpcclient.Lock()
// 				session, ok := rpcclient.sessionMap[msg.RpcSeq()]
// 				rpcclient.Unlock()
// 				if ok {
// 					session.done <- &RpcMessage{msg, nil}
// 				} else {
// 					log.Debug("no rpcsession waiting for rpc response, cmd %X, ip: %v", msg.Cmd(), c.Ip())
// 				}
// 			case CmdRpcError:
// 				rpcclient.Lock()
// 				session, ok := rpcclient.sessionMap[msg.RpcSeq()]
// 				rpcclient.Unlock()
// 				if ok {
// 					session.done <- &RpcMessage{msg, errors.New(string(msg.Body()))}
// 				} else {
// 					log.Debug("no rpcsession waiting for rpc response, cmd %X, ip: %v", msg.Cmd(), c.Ip())
// 				}
// 			default:
// 				engine.RUnlock()
// 				handler, ok := engine.handlers[msg.Cmd()]
// 				engine.RUnlock()
// 				if ok {
// 					engine.Add(1)
// 					defer engine.Done()
// 					defer util.HandlePanic()
// 					handler(c, msg)
// 				} else {
// 					log.Debug("no handler for cmd %v", msg.Cmd())
// 				}
// 			}
// 		})

// 		pool.clients[i] = rpcclient

// 		util.Go(func() {
// 			rpcclient.TcpClient.Keepalive(engine.SockKeepaliveTime())
// 		})

// 		rpcclient.OnClose("-", func(*TcpClient) {
// 			rpcclient.Lock()
// 			defer rpcclient.Unlock()
// 			for _, session := range rpcclient.sessionMap {
// 				close(session.done)
// 			}
// 			rpcclient.sessionMap = map[int64]*rpcsession{}
// 		})
// 	}

// 	return pool, nil
// }
