package graceful

import (
	"fmt"
	"github.com/nothollyhigh/kiss/net"
	"github.com/nothollyhigh/kiss/timer"
	"github.com/nothollyhigh/kiss/util"
	"math"
	"sync"
	"time"
)

var TIME_FOREVER = time.Duration(math.MaxInt64)
var DEFAULT_Q_SIZE = 1024 * 4

type handler struct {
	cli *net.TcpClient
	msg net.IMessage
	cmd func(*net.TcpClient, net.IMessage)
}

type Module struct {
	sync.WaitGroup
	qsize  int
	chFunc chan func()
	chStop chan struct{}

	nextStateTimer *time.Timer
	nextStateFunc  func()

	ticker          *time.Ticker
	enableHeapTimer bool

	heepTimer *timer.Timer
	timers    map[interface{}]util.Empty
}

func (m *Module) Init() {
	m.timers = map[interface{}]util.Empty{}
	if m.enableHeapTimer {
		m.heepTimer = timer.New(fmt.Sprintf("[Module %p]", m))
	}
}

func (m *Module) Start() {
	m.Add(1)

	if m.qsize <= 0 {
		m.qsize = DEFAULT_Q_SIZE
	}

	m.chFunc = make(chan func(), m.qsize)
	m.chStop = make(chan struct{})

	if m.nextStateTimer == nil {
		m.nextStateTimer = time.NewTimer(TIME_FOREVER)
		m.nextStateFunc = nil
	}

	util.Go(func() {
		defer m.Done()
		for {
			select {
			case f := <-m.chFunc:
				util.Safe(f)
			case <-m.nextStateTimer.C:
				if m.nextStateFunc != nil {
					m.nextStateTimer.Reset(TIME_FOREVER)
					f := m.nextStateFunc
					m.nextStateFunc = nil
					util.Safe(f)
				}
			case <-m.chStop:
				return
			}
		}
	})
}

func (m *Module) SetQSize(size int) {
	m.qsize = size
}

func (m *Module) EnableHeapTimer(enable bool) {
	m.enableHeapTimer = enable
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

func (m *Module) Next(timeout time.Duration, f func()) {
	if m.nextStateTimer != nil {
		m.nextStateTimer.Stop()
	}
	m.nextStateTimer = time.NewTimer(timeout)
	m.nextStateFunc = f
}

func (m *Module) After(timeout time.Duration, f func()) interface{} {
	var timerId interface{}
	if !m.enableHeapTimer {
		timerId = time.AfterFunc(timeout, func() {
			m.Push(func() {
				if _, ok := m.timers[timerId]; ok {
					defer delete(m.timers, timerId)
					f()
				}
			})
		})
	} else {
		timerId = m.heepTimer.AfterFunc(timeout, func() {
			m.Push(func() {
				if _, ok := m.timers[timerId]; ok {
					defer delete(m.timers, timerId)
					f()
				}
			})
		})
	}
	m.timers[timerId] = util.Empty{}
	return timerId
}

func (m *Module) Cancel(timerId interface{}) {
	if _, ok := m.timers[timerId]; ok {
		defer delete(m.timers, timerId)

		if !m.enableHeapTimer {
			if t, ok := timerId.(*time.Timer); ok {
				t.Stop()
			}
		} else {
			if t, ok := timerId.(*timer.TimerItem); ok {
				t.Cancel()
			}
		}
	}
}

func (m *Module) Stop() {
	close(m.chStop)
	if m.enableHeapTimer {
		m.heepTimer.Stop()
	}
	m.Wait()
}

func (m *Module) Push(f func(), args ...interface{}) error {
	if len(args) > 0 {
		if timeout, ok := args[0].(time.Duration); ok {
			after := time.NewTimer(timeout)
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
