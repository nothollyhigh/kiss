package timer

import (
	"container/heap"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"math"
	"sync"
	"time"
)

const (
	TIME_FOREVER = time.Duration(math.MaxInt64)
)

var (
	DefaultTimer = New("default")
)

// timer item
type TimerItem struct {
	Index    int
	Expire   time.Time
	Callback func()
	Parent   *Timer
}

// cancel timer
func (it *TimerItem) Cancel() {
	it.Parent.Cancel(it)
}

// reset timer
func (it *TimerItem) Reset(delay time.Duration) {
	it.Parent.Reset(it, delay)
}

// timer array for heap sort
type Timers []*TimerItem

// length
func (tm Timers) Len() int { return len(tm) }

// less
func (tm Timers) Less(i, j int) bool {
	return tm[i].Expire.Before(tm[j].Expire)
}

// swap
func (tm Timers) Swap(i, j int) {
	tm[i], tm[j] = tm[j], tm[i]
	tm[i].Index, tm[j].Index = i, j
}

// push
func (tm *Timers) Push(x interface{}) {
	n := len(*tm)
	item := x.(*TimerItem)
	item.Index = n
	*tm = append(*tm, item)
}

// remove
func (tm *Timers) Remove(i int) {
	n := tm.Len() - 1
	if n != i {
		(*tm).Swap(i, n)
		(*tm)[n] = nil
		*tm = (*tm)[:n]
		heap.Fix(tm, i)
	} else {
		(*tm)[n] = nil
		*tm = (*tm)[:n]
	}

}

// pop
func (tm *Timers) Pop() interface{} {
	old := *tm
	n := len(old)
	if n > 0 {
		tm.Swap(0, n-1)
		item := old[n-1]
		item.Index = -1
		*tm = old[:n-1]
		heap.Fix(tm, 0)
		return item
	} else {
		return nil
	}
}

// heap head
func (tm *Timers) Head() *TimerItem {
	t := *tm
	n := t.Len()
	if n > 0 {
		return t[0]
	}
	return nil
}

// timer
type Timer struct {
	sync.Mutex
	tag     string
	timers  Timers
	trigger *time.Timer
	chStop  chan struct{}
}

// block timeout
func (tm *Timer) After(timeout time.Duration) <-chan struct{} {
	tm.Lock()
	defer tm.Unlock()

	ch := make(chan struct{}, 1)
	item := &TimerItem{
		Index:  len(tm.timers),
		Expire: time.Now().Add(timeout),
		Callback: func() {
			ch <- struct{}{}
		},
	}
	tm.timers = append(tm.timers, item)
	heap.Fix(&(tm.timers), item.Index)
	if head := tm.timers.Head(); head == item {
		tm.trigger.Reset(head.Expire.Sub(time.Now()))
	}

	return ch
}

// aysnc callback after timeout
func (tm *Timer) AfterFunc(timeout time.Duration, cb func()) *TimerItem {
	return tm.Once(timeout, cb)
}

// async callback for once after timeout
func (tm *Timer) Once(timeout time.Duration, cb func()) *TimerItem {
	tm.Lock()
	defer tm.Unlock()

	now := time.Now()
	item := &TimerItem{
		Index:    len(tm.timers),
		Expire:   now.Add(timeout),
		Callback: cb,
		Parent:   tm,
	}
	tm.timers = append(tm.timers, item)
	heap.Fix(&(tm.timers), item.Index)
	if head := tm.timers.Head(); head == item {
		tm.trigger.Reset(head.Expire.Sub(now))
	}

	return item
}

// async callback loop task
func (tm *Timer) Schedule(delay time.Duration, interval time.Duration, repeat int64, cb func()) *TimerItem {
	tm.Lock()
	defer tm.Unlock()

	var (
		item *TimerItem
		now  = time.Now()
	)

	looptime := repeat

	item = &TimerItem{
		Index:  len(tm.timers),
		Expire: now.Add(delay),
		Callback: func() {
			now = time.Now()
			looptime--
			if repeat <= 0 || looptime > 0 {
				tm.Lock()

				item.Index = len(tm.timers)
				item.Expire = now.Add(interval)
				tm.timers = append(tm.timers, item)
				heap.Fix(&(tm.timers), item.Index)

				tm.Unlock()
			}
			cb()

			if head := tm.timers.Head(); head == item {
				tm.trigger.Reset(head.Expire.Sub(now))
			}
		},
		Parent: tm,
	}

	tm.timers = append(tm.timers, item)
	heap.Fix(&(tm.timers), item.Index)
	if head := tm.timers.Head(); head == item {
		tm.trigger.Reset(head.Expire.Sub(now))
	}

	return item
}

// cancel timer
func (tm *Timer) Cancel(item *TimerItem) {
	tm.Lock()
	defer tm.Unlock()
	n := tm.timers.Len()
	if n == 0 {
		log.Debug("Timer(%v) Cancel Error: Timer Size Is 0!", tm.tag)
		return
	}
	if item.Index > 0 && item.Index < n {
		if item != tm.timers[item.Index] {
			log.Debug("Timer(%v) Cancel Error: Invalid Item!", tm.tag)
			return
		}
		tm.timers.Remove(item.Index)
	} else if item.Index == 0 {
		if item != tm.timers[item.Index] {
			log.Debug("Timer(%v) Cancel Error: Invalid Item!", tm.tag)
			return
		}
		tm.timers.Remove(item.Index)
		if head := tm.timers.Head(); head != nil && head != item {
			tm.trigger.Reset(head.Expire.Sub(time.Now()))
		}
	} else {
		log.Debug("Timer(%v) Cancel Error: Invalid Index: %d", tm.tag, item.Index)
	}
}

// reset
func (tm *Timer) Reset(item *TimerItem, delay time.Duration) {
	tm.Lock()
	defer tm.Unlock()

	n := tm.timers.Len()
	if n == 0 {
		log.Debug("Timer(%s) Reset Error: Timer Size Is 0!", tm.tag)
		return
	}
	if item.Index < n {
		if item != tm.timers[item.Index] {
			log.Debug("Timer(%s) Reset Error: Invalid Item!", tm.tag)
			return
		}
		item.Expire = time.Now().Add(delay)
		heap.Fix(&(tm.timers), item.Index)
	} else {
		log.Debug("Timer(%s) Reset Error: Invalid Item!", tm.tag)
	}
}

// size
func (tm *Timer) Size() int {
	tm.Lock()
	defer tm.Unlock()
	return len(tm.timers)
}

// stop
func (tm *Timer) Stop() {
	tm.Lock()
	defer tm.Unlock()
	close(tm.chStop)
	tm.trigger.Stop()
}

// once
func (tm *Timer) once() {
	defer util.HandlePanic()
	tm.Lock()
	item := tm.timers.Pop()
	if item != nil {
		if head := tm.timers.Head(); head != nil {
			tm.trigger.Reset(head.Expire.Sub(time.Now()))
		}
	} else {
		tm.trigger.Reset(TIME_FOREVER)
	}
	tm.Unlock()

	if item != nil {
		item.(*TimerItem).Callback()
	}
}

// block timeout
func After(timeout time.Duration) <-chan struct{} {
	return DefaultTimer.After(timeout)
}

// async callback once after timeout
func AfterFunc(timeout time.Duration, cb func()) *TimerItem {
	return DefaultTimer.AfterFunc(timeout, cb)
}

// async callback once after timeout
func Once(timeout time.Duration, cb func()) *TimerItem {
	return DefaultTimer.Once(timeout, cb)
}

// async callback loop task
func Schedule(delay time.Duration, interval time.Duration, repeat int64, cb func()) *TimerItem {
	return DefaultTimer.Schedule(delay, interval, repeat, cb)
}

// cancel timer
func Cancel(item *TimerItem) {
	DefaultTimer.Cancel(item)
}

// timer factory
func New(tag string) *Timer {
	tm := &Timer{
		tag:     tag,
		trigger: time.NewTimer(TIME_FOREVER),
		chStop:  make(chan struct{}),
		timers:  []*TimerItem{},
	}

	go func() {
		defer log.Debug("timer(%v) stopped", tag)
		for {
			select {
			case <-tm.trigger.C:
				tm.once()
			case <-tm.chStop:
				return
			}
		}
	}()

	return tm
}
