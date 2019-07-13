package net

import (
	"net"
	"time"
)

// ping tcp addr
func Ping(addr string, to time.Duration) error {
	conn, err := net.DialTimeout("tcp", addr, to)
	if err == nil {
		defer conn.Close()

	}
	return err
}
