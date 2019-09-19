package util

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrWorkerPoolTimeout = errors.New("timeout")
	ErrWorkerPoolStopped = errors.New("worker pool has stopped")
)

func Go(cb func()) {
	go func() {
		defer HandlePanic()
		cb()
	}()
}

// func Go(cb interface{}, args ...interface{}) {
// 	f := reflect.ValueOf(cb)
// 	go func() {
// 		defer HandlePanic(true)
// 		n := len(args)
// 		if n > 0 {
// 			refargs := make([]reflect.Value, n)
// 			for i := 0; i < n; i++ {
// 				refargs[i] = reflect.ValueOf(args[i])
// 			}
// 			f.Call(refargs)

// 		} else {
// 			f.Call(nil)
// 		}
// 	}()
// }

type WorkerPool struct {
	sync.RWMutex
	sync.WaitGroup
	tag     string
	chTask  chan func()
	running bool
}

func (c *WorkerPool) runLoop(partition int) {
	Go(func() {
		defer c.Done()

		for h := range c.chTask {
			Safe(func() {
				defer c.Done()
				h()
			})

		}
	})
}

func (c *WorkerPool) Go(h func(), to time.Duration) error {
	c.RLock()
	defer c.RUnlock()

	if !c.running {
		return ErrWorkerPoolStopped
	}

	defer HandlePanic()

	c.Add(1)

	if to > 0 {
		after := time.NewTimer(to)
		defer after.Stop()
		select {
		case c.chTask <- h:
		case <-after.C:
			c.Done()
			return ErrWorkerPoolTimeout
		}
	} else {
		c.chTask <- h
	}
	return nil
}

func (c *WorkerPool) Stop() {
	c.Add(1)
	Go(func() {
		c.Lock()
		defer c.Unlock()
		defer c.Done()
		c.running = false
		close(c.chTask)
	})
	c.Wait()
}

func NewWorkerPool(tag string, qCap int, corNum int) *WorkerPool {
	c := &WorkerPool{
		tag:     tag,
		chTask:  make(chan func(), qCap),
		running: true,
	}
	for i := 0; i < corNum; i++ {
		c.Add(1)
		c.runLoop(i)
	}
	return c
}

type LinkTask struct {
	caller func(*LinkTask)
	pre    chan interface{}
	ch     chan interface{}
}

func (task *LinkTask) Done(data interface{}) {
	task.ch <- data
}

func (task *LinkTask) WaitPre() interface{} {
	if task.pre != nil {
		return <-task.pre
	}
	return nil
}

type WorkerPoolLink struct {
	sync.RWMutex
	sync.WaitGroup
	tag     string
	chTask  chan *LinkTask
	pre     *LinkTask
	running bool
}

func (c *WorkerPoolLink) runLoop(partition int) {
	Go(func() {
		defer c.Done()
		for task := range c.chTask {
			Safe(func() {
				defer c.Done()
				task.caller(task)
			})
		}
	})
}

func (c *WorkerPoolLink) Go(h func(task *LinkTask)) error {
	c.RLock()
	defer c.RUnlock()

	if !c.running {
		return ErrWorkerPoolStopped
	}

	var task *LinkTask
	if c.pre != nil {
		task = &LinkTask{
			caller: h,
			pre:    c.pre.ch,
			ch:     make(chan interface{}, 1),
		}
	} else {
		task = &LinkTask{
			caller: h,
			ch:     make(chan interface{}, 1),
		}
	}
	c.pre = task

	c.Add(1)
	c.chTask <- task

	return nil
}

func (c *WorkerPoolLink) Stop() {
	c.Add(1)
	Go(func() {
		c.Lock()
		defer c.Unlock()
		defer c.Done()
		c.running = false
		close(c.chTask)
	})
	c.Wait()
}

func NewWorkerPoolLink(tag string, qCap int, corNum int) *WorkerPoolLink {
	c := &WorkerPoolLink{
		tag:     tag,
		chTask:  make(chan *LinkTask, qCap),
		running: true,
	}
	for i := 0; i < corNum; i++ {
		c.Add(1)
		c.runLoop(i)
	}

	return c
}
