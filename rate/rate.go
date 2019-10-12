package rate

// deprecated

// import (
// 	"fmt"
// 	"github.com/nothollyhigh/kiss/log"
// 	"github.com/nothollyhigh/kiss/timer"
// 	"runtime"
// 	"sync"
// 	"time"
// )

// var (
// 	defaultLimiter = &Limiter{
// 		trigger:    timer.New("kiss-default-limiter"),
// 		bucketsMap: map[string]chan empty{},
// 	}

// 	ErrTimeout = fmt.Errorf("rate Wait timeout")
// )

// type empty struct{}

// type Limiter struct {
// 	sync.RWMutex
// 	trigger    *timer.Timer
// 	bucketsMap map[string]chan empty
// }

// func (l *Limiter) get(times int, interval time.Duration, constant bool, tags ...interface{}) chan empty {
// 	_, file, line, _ := runtime.Caller(2)
// 	key := fmt.Sprintf("%s_%d", file, line)
// 	for _, v := range tags {
// 		key += fmt.Sprintf("_%v", v)
// 	}

// 	l.Lock()
// 	defer l.Unlock()

// 	buckets, ok := l.bucketsMap[key]
// 	if ok {
// 		return buckets
// 	}

// 	if constant {
// 		buckets = make(chan empty, 1)
// 		interval = interval / time.Duration(times)
// 		l.trigger.Schedule(0, interval, 0, func() {
// 			select {
// 			case buckets <- empty{}:
// 			default:
// 			}

// 		})
// 	} else {
// 		buckets = make(chan empty, times)
// 		l.trigger.Schedule(0, interval, 0, func() {
// 			for i := 0; i < times; i++ {
// 				select {
// 				case buckets <- empty{}:
// 				default:
// 				}
// 			}
// 		})
// 	}

// 	l.bucketsMap[key] = buckets

// 	return buckets
// }

// func (l *Limiter) Wait(times int, interval time.Duration, constant bool, timeout time.Duration, tags ...interface{}) error {
// 	if times < 1 {
// 		log.Panic("rate Wait failed: [invalid times arg: %v]", times)
// 	}
// 	if interval < 1 {
// 		log.Panic("rate Wait failed: [invalid interval arg: %v]", interval)
// 	}

// 	if timeout > 0 {
// 		after := time.NewTimer(timeout)
// 		defer after.Stop()
// 		select {
// 		case <-l.get(times, interval, constant, tags...):
// 		case <-after.C:
// 			return ErrTimeout
// 		}
// 	} else {
// 		<-l.get(times, interval, constant, tags...)
// 	}

// 	return nil
// }

// func New(tag string) *Limiter {
// 	return &Limiter{
// 		trigger:    timer.New(fmt.Sprintf("kiss-limiter-%v", tag)),
// 		bucketsMap: map[string]chan empty{},
// 	}
// }

// func Wait(times int, interval time.Duration, constant bool, timeout time.Duration, tags ...interface{}) error {
// 	return defaultLimiter.Wait(times, interval, constant, timeout, tags)
// }
