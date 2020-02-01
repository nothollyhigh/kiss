package graceful

import (
	"fmt"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/net"
	"github.com/nothollyhigh/kiss/timer"
	"github.com/nothollyhigh/kiss/util"
	"math"
	"sync"
	"time"
)

var TIME_FOREVER = time.Duration(math.MaxInt64)
var TIME_EXEC_BLOCK = time.Second * 60
var DEFAULT_Q_SIZE = 1024 * 4
var ERR_MODULE_PUSH_TIMEOUT = fmt.Errorf("push timeout")
var ERR_MODULE_STOPPED = fmt.Errorf("stopped")

type handler struct {
	cli *net.TcpClient
	msg net.IMessage
	cmd func(*net.TcpClient, net.IMessage)
}

type Module struct {
	sync.RWMutex
	sync.WaitGroup

	running bool

	qsize  int
	chFunc chan func()
	chStop chan struct{}

	nextStateTimer *time.Timer
	nextStateFunc  func()

	// ticker          *time.Ticker

	timers map[interface{}]util.Empty

	enableHeapTimer bool
	heepTimer       *timer.Timer
}

func (m *Module) Init() {
	m.timers = map[interface{}]util.Empty{}
	if m.enableHeapTimer {
		m.heepTimer = timer.New(fmt.Sprintf("[Module %p]", m))
	}
}

func (m *Module) Start() {
	m.Lock()
	defer m.Unlock()

	if m.running {
		return
	}

	m.Add(1)

	if m.qsize <= 0 {
		m.qsize = DEFAULT_Q_SIZE
	}

	m.running = true

	m.chFunc = make(chan func(), m.qsize)
	m.chStop = make(chan struct{}, 1)

	if m.nextStateTimer == nil {
		m.nextStateTimer = time.NewTimer(TIME_FOREVER)
		m.nextStateFunc = nil
	}

	util.Go(func() {
		defer m.Done()

		running := m.running

		for {
			select {
			case f := <-m.chFunc:
				m.RLock()
				running = m.running
				m.RUnlock()
				if running {
					util.Safe(f)
				}
			case <-m.nextStateTimer.C:
				m.RLock()
				running = m.running
				m.RUnlock()
				if running {
					if m.nextStateFunc != nil {
						m.nextStateTimer.Reset(TIME_FOREVER)
						f := m.nextStateFunc
						m.nextStateFunc = nil
						util.Safe(f)
					}
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

func (m *Module) Next(timeout time.Duration, f func()) {
	m.Lock()
	defer m.Unlock()
	if m.running {
		m.nextStateTimer.Reset(timeout)
		m.nextStateFunc = f
	}
}

func (m *Module) After(timeout time.Duration, f func()) interface{} {
	m.Lock()
	running := m.running
	m.Unlock()

	var timerId interface{}

	if running {
		if !m.enableHeapTimer {
			timerId = time.AfterFunc(timeout, func() {
				err := m.push(func() {
					m.Lock()
					if _, ok := m.timers[timerId]; ok {
						delete(m.timers, timerId)
						m.Unlock()
						f()
						return
					}
					m.Unlock()
				})
				if err != nil {
					log.Error("[Module %p] After failed: %v", m, err)
				}
			})
		} else {
			timerId = m.heepTimer.AfterFunc(timeout, func() {
				err := m.push(func() {
					m.Lock()
					if _, ok := m.timers[timerId]; ok {
						delete(m.timers, timerId)
						m.Unlock()
						f()
						return
					}
					m.Unlock()
				})
				if err != nil {
					log.Error("[Module %p] After failed: %v", m, err)
				}
			})
		}

		m.Lock()
		m.timers[timerId] = util.Empty{}
		m.Unlock()
	}

	return timerId
}

func (m *Module) Cancel(timerId interface{}) {
	m.Lock()
	_, ok := m.timers[timerId]
	if ok {
		defer delete(m.timers, timerId)
	}
	m.Unlock()

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

func (m *Module) Stop() {
	util.Go(func() {
		m.Lock()
		defer m.Unlock()

		if m.running {
			m.running = false

			if m.enableHeapTimer {
				m.heepTimer.Stop()
			} else {
				for t, _ := range m.timers {
					if tm, ok := t.(*time.Timer); ok {
						tm.Stop()
					}
				}
			}

			close(m.chStop)
		}
	})

	m.Wait()
}

func (m *Module) push(f func(), args ...interface{}) error {
	if len(args) > 0 {
		timeout, ok := args[0].(time.Duration)
		if !ok {
			timeout = TIME_EXEC_BLOCK
		}

		after := time.NewTimer(timeout)
		defer after.Stop()
		select {
		case m.chFunc <- f:
			return nil
		case <-after.C:
			return ERR_MODULE_PUSH_TIMEOUT
		}

	} else {
		timeout := TIME_EXEC_BLOCK

		after := time.NewTimer(timeout)
		defer after.Stop()
		select {
		case m.chFunc <- f:
			return nil
		case <-after.C:
			return ERR_MODULE_PUSH_TIMEOUT
		}
	}

	return nil
}

func (m *Module) Exec(f func(), args ...interface{}) error {
	m.Lock()
	running := m.running
	m.Unlock()

	if running {
		return m.push(f, args...)
	}

	return ERR_MODULE_STOPPED
}
