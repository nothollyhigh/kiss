package main

import (
	"github.com/nothollyhigh/kiss/graceful"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/net"
	"github.com/nothollyhigh/kiss/util"
	"os"
	"sync"
	"syscall"
	"time"
)

var ()

const (
	addr = ":8888"

	CMD_ECHO   = uint32(1)
	CMD_ECHO_2 = uint32(2)
)

type EchoModule struct {
	graceful.Module
}

func (m *EchoModule) onEcho(client *net.TcpClient, msg net.IMessage) {
	log.Info("tcp server EchoModule from %v: %v", client.Conn.RemoteAddr().String(), string(msg.Body()))
	client.SendMsg(msg)
}

type EchoModule2 struct {
	graceful.Module
	wgStop sync.WaitGroup
}

func (m *EchoModule2) onEcho(client *net.TcpClient, msg net.IMessage) {
	log.Info("tcp server EchoModule2 from %v: %v", client.Conn.RemoteAddr().String(), string(msg.Body()))
	client.SendMsg(msg)
}

// 异步推出
func (m *EchoModule2) Stop() {
	m.wgStop.Add(1)

	m.After(time.Second*2, func() {
		util.Go(func() {
			// 先处理模块内部优雅退出, 父类退出
			m.Module.Stop()

			// 模块stop结束
			m.wgStop.Done()
		})
	})

	// 等待异步优雅推出
	m.wgStop.Wait()
}

var (
	echoModule  = &EchoModule{}
	echoModule2 = &EchoModule2{}
)

func onEcho(client *net.TcpClient, msg net.IMessage) {
	echoModule.PushNet(echoModule.onEcho, client, msg)
}

func onEcho2(client *net.TcpClient, msg net.IMessage) {
	echoModule2.PushNet(echoModule2.onEcho, client, msg)
}

func main() {
	server := net.NewTcpServer("Echo")

	server.Handle(CMD_ECHO, onEcho)
	server.Handle(CMD_ECHO_2, onEcho2)

	for i := 0; i < 10; i++ {
		n := i
		echoModule.After(time.Second*time.Duration(n), func() {
			log.Info("time event: %v", n)
		})
	}

	graceful.Register(echoModule)
	graceful.Register(echoModule2)
	graceful.Start(1024)

	util.Go(func() { server.Start(addr) })

	util.HandleSignal(func(sig os.Signal) {
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			graceful.Stop()
			server.StopWithTimeout(time.Second*10, nil)
			os.Exit(0)
		}
	})
}