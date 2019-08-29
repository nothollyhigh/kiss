package main

import (
	"fmt"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/net"
	"time"
)

var (
	addr = "127.0.0.1:8888"

	CMD_ECHO   = uint32(1)
	CMD_ECHO_2 = uint32(2)
)

func onConnected(c *net.TcpClient) {
	log.Info("TcpClient OnConnected")
}

func onEcho(client *net.TcpClient, msg net.IMessage) {
	log.Debug("tcp client onEcho from %v: %v", client.Conn.RemoteAddr().String(), string(msg.Body()))
}

func onEcho2(client *net.TcpClient, msg net.IMessage) {
	log.Debug("tcp client onEcho2 from %v: %v", client.Conn.RemoteAddr().String(), string(msg.Body()))
}

func main() {
	autoReconn := true
	netengine := net.NewTcpEngine()

	netengine.Handle(CMD_ECHO, onEcho)
	netengine.Handle(CMD_ECHO_2, onEcho2)

	client, err := net.NewTcpClient(addr, netengine, nil, autoReconn, onConnected)
	if err != nil {
		log.Panic("NewTcpClient failed: %v, %v", client, err)
	}

	for i := 0; true; i++ {
		cmd := uint32(i%2 + 1)
		err = client.SendMsg(net.NewMessage(cmd, []byte(fmt.Sprintf("hello %v", i))))
		if err != nil {
			log.Error("tcp client echo failed: %v", err)
		}
		time.Sleep(time.Second)
	}
}
