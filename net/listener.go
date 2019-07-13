package net

import (
	"net"
	"time"
)

// tcp socket option
type SocketOpt struct {
	NoDelay           bool
	Keepalive         bool
	KeepaliveInterval time.Duration
	ReadBufLen        int
	WriteBufLen       int
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	MaxHeaderBytes    int
}

// tcp listener
type Listener struct {
	*net.TCPListener
	opt *SocketOpt
}

// accept
func (ln Listener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return tc, err
	}

	if ln.opt == nil {
		//same as net.http.Server.ListenAndServe
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(3 * time.Minute)
	} else {
		tc.SetNoDelay(ln.opt.NoDelay)
		tc.SetKeepAlive(ln.opt.Keepalive)
		if ln.opt.Keepalive && ln.opt.KeepaliveInterval > 0 {
			tc.SetKeepAlivePeriod(ln.opt.KeepaliveInterval)
		}
		if ln.opt.ReadBufLen > 0 {
			tc.SetReadBuffer(ln.opt.ReadBufLen)
		}
		if ln.opt.WriteBufLen > 0 {
			tc.SetWriteBuffer(ln.opt.WriteBufLen)
		}
	}

	return tc, nil
}

// tcp listener factory
func NewListener(addr string, opt *SocketOpt) (net.Listener, error) {
	if addr == "" {
		addr = ":http"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		listener, err = net.Listen("tcp6", addr)
	}

	if err == nil {
		// if opt != nil {
		// 	if opt.Keepalive && opt.KeepaliveInterval < time.Minute {
		// 		opt.KeepaliveInterval = minKeepaliveInterval
		// 	}
		// 	if opt.ReadBufLen <= 0 {
		// 		opt.ReadBufLen = defautRecvBufLen
		// 	}
		// 	if opt.WriteBufLen <= 0 {
		// 		opt.WriteBufLen = defautSendBufLen
		// 	}
		// }
		return Listener{listener.(*net.TCPListener), opt}, err
	}
	return nil, err
}
