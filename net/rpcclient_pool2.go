package net

// import (
// 	"fmt"
// 	"sync/atomic"
// 	"time"
// )

// type RpcClientPool struct {
// 	idx     int64
// 	size    uint64
// 	clients []IRpcClient
// }

// func (pool *RpcClientPool) Client() IRpcClient {
// 	idx := atomic.AddInt64(&pool.idx, 1)
// 	return pool.clients[uint64(idx)%pool.size]
// }

// func (pool *RpcClientPool) Call(method string, req interface{}, rsp interface{}, timeout time.Duration) error {
// 	idx := uint64(atomic.AddInt64(&pool.idx, 1)) % pool.size
// 	return pool.clients[idx].Call(method, req, rsp, timeout)
// }

// func NewRpcClientPool(addr string, engine *TcpEngin, codec ICodec, size int, onConnected func(*TcpClient)) (*RpcClientPool, error) {
// 	if engine == nil {
// 		engine = NewTcpEngine()
// 		engine.SetSendQueueSize(DefaultSockRpcSendQSize)
// 		engine.SetSockRecvBlockTime(DefaultSockRpcRecvBlockTime)
// 	}
// 	//func NewClientPool(addr string, size int) *RpcClientPool {
// 	if size <= 0 {
// 		panic(fmt.Errorf("NewClientPool failed: invalid size %v", size))
// 	}
// 	clients := make([]IRpcClient, size)
// 	var err error
// 	for i := 0; i < size; i++ {
// 		clients[i], err = NewRpcClient(addr, engine, codec, onConnected)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return &RpcClientPool{
// 		idx:     0,
// 		size:    uint64(size),
// 		clients: clients,
// 	}, nil
// }
