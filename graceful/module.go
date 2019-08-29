package graceful

import (
	"github.com/nothollyhigh/kiss/net"
	"github.com/nothollyhigh/kiss/util"
	"math"
	"sync"
	"time"
)

var TIME_FOREVER = time.Duration(math.MaxInt64)
var DEFAULT_Q_SIZE = 1024 * 8

type handler struct {
	cli *net.TcpClient
	msg net.IMessage
	cmd func(*net.TcpClient, net.IMessage)
}

type Module struct {
	sync.WaitGroup
	chNet  chan handler
	chFunc chan func()
	chStop chan struct{}

	// timer     *time.Timer
	// timerFunc func()
}

func (m *Module) Start(args ...interface{}) {
	m.Add(1)

	qsize := DEFAULT_Q_SIZE
	if len(args) > 0 {
		if size, ok := args[0].(int); ok && size > 0 {
			qsize = size
		}
	}

	m.chNet = make(chan handler, qsize)
	m.chFunc = make(chan func(), qsize)
	m.chStop = make(chan struct{})

	// m.timer = time.NewTimer(TIME_FOREVER)

	util.Go(func() {
		defer m.Done()
		for {
			select {
			case h := <-m.chNet:
				util.Safe(func() { h.cmd(h.cli, h.msg) })
			case f := <-m.chFunc:
				util.Safe(f)
			case <-m.chStop:
				return
				// case <-m.timer.C:
				// 	if m.timerFunc != nil {
				// 		util.Safe(m.timerFunc)
				// 	}
				// 	m.timer.Reset(TIME_FOREVER)
				// }
			}
		}
	})
}

func (m *Module) After(to time.Duration, f func()) {
	time.AfterFunc(to, func() {
		m.Push(f)
	})
}

func (m *Module) Stop() {
	close(m.chStop)
	m.Wait()

}

func (m *Module) Push(f func()) {
	m.chFunc <- f
}

func (m *Module) PushNet(cmd func(*net.TcpClient, net.IMessage), cli *net.TcpClient, msg net.IMessage) {
	m.chNet <- handler{cli, msg, cmd}
}
