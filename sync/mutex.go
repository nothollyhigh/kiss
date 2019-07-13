package sync

import (
	"fmt"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"sync"
	"time"
)

var (
	debug      = false
	mtxlock    = &sync.Mutex{}
	timeout    = (time.Second * 3)
	mtxmutexes = map[interface{}]*time.Timer{}

	separator = "---------------------------------------\n"
)

// get lock timer
func getLockTimer(key interface{}) *time.Timer {
	mtxlock.Lock()
	defer mtxlock.Unlock()
	if tmr, ok := mtxmutexes[key]; ok {
		return tmr
	}
	return nil
}

// add lock timer
func addLockTimer(key interface{}, tmr *time.Timer) {
	mtxlock.Lock()
	defer mtxlock.Unlock()
	mtxmutexes[key] = tmr
}

// remove lock timer
func removeLockTimer(key interface{}, expired bool) {
	mtxlock.Lock()
	defer mtxlock.Unlock()
	if tmr, ok := mtxmutexes[key]; ok {
		if !expired {
			tmr.Stop()
		}
		delete(mtxmutexes, key)
	}
}

// set debug flag and timeout for print deadlock stacks
func SetDebug(flag bool, args ...interface{}) {
	debug = flag
	if debug {
		if mtxmutexes == nil {
			mtxmutexes = make(map[interface{}]*time.Timer)
		}
	}
	if len(args) == 1 {
		t, ok := args[0].(time.Duration)
		if ok {
			timeout = t
		}
	}
}

// mutex
type Mutex struct {
	sync.Mutex
	unlockkey string
	lastCall  string
}

// lock
func (mt *Mutex) Lock() {
	if debug {
		t1 := time.Now()
		stack := util.GetStacks()
		tmr := time.AfterFunc(timeout, func() {
			str := "\n" + separator + fmt.Sprintf("\nMutex Lock() Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds()) + fmt.Sprintf("  now: %v\n  this Call :", t1) + stack + "\n  last Call :\n" + mt.lastCall + "\n" + separator
			log.Debug(str)
		})

		mt.Mutex.Lock()

		tmr.Stop()
		mt.lastCall = stack

		{
			if mt.unlockkey == "" {
				mt.unlockkey = fmt.Sprintf("%pul", mt)
			}
			if tmr := getLockTimer(mt.unlockkey); tmr == nil {
				t1 := time.Now()
				tmr = time.AfterFunc(timeout, func() {
					str := "\n" + separator + fmt.Sprintf("\nMutex Unlock() Wait Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds()) + fmt.Sprintf("  now: %v\n", t1) + "  last Call :\n" + mt.lastCall + "\n" + separator
					log.Debug(str)
					removeLockTimer(mt.unlockkey, true)
				})
				addLockTimer(mt.unlockkey, tmr)
			}
		}
	} else {
		mt.Mutex.Lock()
	}
}

// unlock
func (mt *Mutex) Unlock() {
	mt.Mutex.Unlock()
	if debug {
		removeLockTimer(mt.unlockkey, false)
	}
}

// rwmutex
type RWMutex struct {
	sync.RWMutex
	unlockkey string
	lastCall  string
}

// lock
func (rwmt *RWMutex) Lock() {
	if debug {
		t1 := time.Now()
		stack := util.GetStacks()
		tmr := time.AfterFunc(timeout, func() {
			str := "\n" + separator + fmt.Sprintf("\nRWMutex Lock() Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds()) + fmt.Sprintf("  now: %v\n", t1) + "  this Call :\n" + stack + "  last Call :\n" + rwmt.lastCall + "\n" + separator
			log.Debug(str)
		})

		rwmt.RWMutex.Lock()

		tmr.Stop()
		rwmt.lastCall = stack

		{
			if rwmt.unlockkey == "" {
				rwmt.unlockkey = fmt.Sprintf("%pul", rwmt)
			}
			if tmr := getLockTimer(rwmt.unlockkey); tmr == nil {
				t1 := time.Now()
				tmr = time.AfterFunc(timeout, func() {
					str := "\n" + separator + fmt.Sprintf("\nRWMutex Unlock() Wait Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds()) + fmt.Sprintf("  now: %v\n", t1) + "  last Call :\n" + rwmt.lastCall + "\n" + separator
					log.Debug(str)
					removeLockTimer(rwmt.unlockkey, true)
				})
				addLockTimer(rwmt.unlockkey, tmr)
			}
		}
	} else {
		rwmt.RWMutex.Lock()
	}
}

// unlock
func (rwmt *RWMutex) Unlock() {
	rwmt.RWMutex.Unlock()
	if debug {
		removeLockTimer(rwmt.unlockkey, false)
	}
}

// rlock
func (rwmt *RWMutex) RLock() {
	if debug {
		t1 := time.Now()
		stack := util.GetStacks()
		tmr := time.AfterFunc(timeout, func() {
			str := "\n" + separator + fmt.Sprintf("\nRWMutex RLock() Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds()) + fmt.Sprintf("  now: %v\n", t1) + "  this Call :\n" + stack + "  last Call :\n" + rwmt.lastCall + separator
			log.Debug(str)
		})

		rwmt.RWMutex.RLock()

		tmr.Stop()
		rwmt.lastCall = stack

		{
			if rwmt.unlockkey == "" {
				rwmt.unlockkey = fmt.Sprintf("%pul", rwmt)
			}
			if tmr := getLockTimer(rwmt.unlockkey); tmr == nil {
				t1 := time.Now()
				tmr = time.AfterFunc(timeout, func() {
					str := "\n" + separator + fmt.Sprintf("\nRWMutex RUnlock() Wait Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds()) + fmt.Sprintf("  now: %v\n", t1) + "  last Call :\n" + rwmt.lastCall + "\n" + separator
					log.Debug(str)
					removeLockTimer(rwmt.unlockkey, true)
				})
				addLockTimer(rwmt.unlockkey, tmr)
			}
		}
	} else {
		rwmt.RWMutex.RLock()
	}
}

// rnnlock
func (rwmt *RWMutex) RUnlock() {
	rwmt.RWMutex.RUnlock()
	if debug {
		removeLockTimer(rwmt.unlockkey, false)
	}
}
