package rate

import (
	"fmt"
	"github.com/nothollyhigh/kiss/timer"
	"runtime"
	"sync"
	"time"
)

var (
	mgr = limiterMgr{
		limters: map[string]*limiter{},
		trigger: timer.New("kiss-limiter"),
	}
)

type empty struct{}

type limiter struct {
	buckets chan empty
}

type limiterMgr struct {
	sync.RWMutex
	trigger *timer.Timer
	limters map[string]*limiter
}

func (m *limiterMgr) get(times int, interval time.Duration, constant bool) *limiter {
	_, file, line, _ := runtime.Caller(2)
	key := fmt.Sprintf("%s%d", file, line)

	m.Lock()
	defer m.Unlock()

	l, ok := m.limters[key]
	if ok {
		return l
	}

	l = &limiter{
		buckets: make(chan empty, times),
	}

	if constant {
		l.buckets = make(chan empty, 1)
		interval = interval / time.Duration(times)
		m.trigger.Schedule(0, interval, 0, func() {
			select {
			case l.buckets <- empty{}:
			default:
			}

		})
	} else {
		l.buckets = make(chan empty, times)
		m.trigger.Schedule(0, interval, 0, func() {
			for i := 0; i < times; i++ {
				select {
				case l.buckets <- empty{}:
				default:
				}
			}
		})
	}

	m.limters[key] = l

	return l
}

func Limit(times int, interval time.Duration, constant bool) {
	if times < 1 {
		panic(fmt.Errorf("rate.Limit failed: [invalid times arg: %v]", times))
	}
	if interval < 1 {
		panic(fmt.Errorf("rate.Limit failed: [invalid interval arg: %v]", interval))
	}
	<-mgr.get(times, interval, constant).buckets
}
