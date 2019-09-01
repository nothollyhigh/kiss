package graceful

import (
	"fmt"
	"github.com/nothollyhigh/kiss/net"
	"github.com/nothollyhigh/kiss/util"
	"sync"
	"time"
)

// var TIME_FOREVER = time.Duration(math.MaxInt64)
var DEFAULT_Q_SIZE = 1024 * 8

type handler struct {
	cli *net.TcpClient
	msg net.IMessage
	cmd func(*net.TcpClient, net.IMessage)
}

type Module struct {
	sync.WaitGroup
	chFunc chan func()
	chStop chan struct{}

	ticker *time.Ticker
}

func (m *Module) Start(args ...interface{}) {
	m.Add(1)

	qsize := DEFAULT_Q_SIZE
	if len(args) > 0 {
		if size, ok := args[0].(int); ok && size > 0 {
			qsize = size
		}
	}

	m.chFunc = make(chan func(), qsize)
	m.chStop = make(chan struct{})

	util.Go(func() {
		defer m.Done()
		for {
			select {
			case f := <-m.chFunc:
				util.Safe(f)
			case <-m.chStop:
				return
			}
		}
	})
}

func (m *Module) EnableTick(interval time.Duration, onTick func()) error {
	if m.ticker != nil {
		return fmt.Errorf("ticker already started")
	}

	m.ticker = time.NewTicker(interval)

	util.Go(func() {
		defer func() {
			m.ticker.Stop()
			m.ticker = nil
		}()

		for {
			select {
			case <-m.ticker.C:
				m.Push(onTick)
			case <-m.chStop:
				return
			}
		}
	})

	return nil
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

func (m *Module) Push(f func(), args ...interface{}) error {
	if len(args) > 0 {
		if to, ok := args[0].(time.Duration); ok {
			after := time.NewTimer(to)
			defer after.Stop()
			select {
			case m.chFunc <- f:
				return nil
			case <-after.C:
				return fmt.Errorf("timeout")
			}
		}
	}
	m.chFunc <- f
	return nil
}
