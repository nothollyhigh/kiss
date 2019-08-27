package net

import (
	"time"
)

var (
	// default tcp nodlay
	DefaultSockNodelay = true
	// default tcp keepalive
	DefaultSockKeepalive = false
	// default tcp bufio reader
	DefaultSockBufioReaderEnabled = false
	// default tcp client send queue size
	DefaultSendQSize = 512
	// default tcp client read buf length
	DefaultSockRecvBufLen = 1024
	// default tcp client write buf length
	DefaultSockSendBufLen = 1024
	// default max tcp client packet length
	DefaultSockPackMaxLen = 1024 * 1024
	// default tcp client linger time
	DefaultSockLingerSeconds = 0
	// default tcp client keepalive interval
	DefaultSockKeepaliveTime = time.Second * 60
	// default tcp client read block time
	DefaultSockRecvBlockTime = time.Second * 65
	// default tcp client write block time
	DefaultSockSendBlockTime = time.Second * 5

	// default rpc send queue size
	DefaultSockRpcSendQSize = 8192
	// default rpc read block time
	DefaultSockRpcRecvBlockTime = time.Second * 3600 * 24

	// default max concurrent
	DefaultMaxOnline = int64(40960)

	// default read block time
	DefaultReadTimeout = time.Second * 35

	// default write block time
	DefaultWriteTimeout = time.Second * 5

	// default shutdown timeout
	DefaultShutdownTimeout = time.Second * 5

	// default max websocket read length
	DefaultReadLimit int64 = 1024 * 1024
)
