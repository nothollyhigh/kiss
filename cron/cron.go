package cron

import (
	"github.com/gorhill/cronexpr"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"sync"
	"time"
)

type Cron struct {
	sync.Mutex
	sync.WaitGroup

	Tag      string
	Task     func()
	Async    bool
	CronLine string

	timer  *time.Timer
	chStop chan struct{}
	expr   *cronexpr.Expression
}

func (cron *Cron) Start() {
	cron.Add(1)

	cron.Lock()

	if cron.timer != nil {
		cron.Done()
		return
	}

	expr, err := cronexpr.Parse(cron.CronLine)
	if err != nil {
		log.Panic("Cron Start failed, invalid cronLine: %v", err)
	}

	nextTime := expr.Next(time.Now())
	cron.timer = time.NewTimer(time.Until(nextTime))

	cron.Unlock()

	util.Go(func() {
		log.Info("cron %v start", cron.Tag)

		defer func() {
			cron.Lock()
			cron.Done()
			cron.timer.Stop()
			cron.timer = nil
			cron.Unlock()
		}()

		for {
			select {
			case <-cron.timer.C:
			case <-cron.chStop:
				return
			}
			if cron.Async {
				util.Go(cron.Task)
			} else {
				util.Safe(cron.Task)
			}

			nextTime = expr.Next(time.Now())
			cron.timer.Reset(time.Until(nextTime))
		}
	})
}

func (cron *Cron) Stop() {
	if cron == nil {
		return
	}

	defer log.Info("cron %v stop", cron.Tag)

	cron.Lock()
	chStop := cron.chStop
	cron.chStop = nil
	cron.Unlock()

	if chStop == nil {
		return
	}

	close(chStop)

	cron.Wait()
}

func (cron *Cron) Next(fromTime time.Time) time.Time {
	var next time.Time
	if cron.expr != nil {
		next = cron.expr.Next(fromTime)
	}
	return next

}

func New(tag string, cronLine string, async bool, task func()) *Cron {
	if cronLine == "" {
		log.Panic("NewCron failed: nil cornLine")
	}

	if task == nil {
		log.Panic("NewCron failed: nil task func")
	}

	return &Cron{
		Tag:      tag,
		Async:    async,
		CronLine: cronLine,
		Task:     task,

		chStop: make(chan struct{}),
	}
}

type CronMgr struct {
	sync.RWMutex
	crons map[string]*Cron
}

func (mgr *CronMgr) Add(tag string, cronLine string, async bool, task func()) *Cron {
	mgr.Lock()
	defer mgr.Unlock()

	if _, ok := mgr.crons[tag]; ok {
		log.Panic("CronMgr.Add failed: cron %v exists", tag)
	}
	cron := New(tag, cronLine, async, task)
	mgr.crons[tag] = cron
	cron.Start()

	return cron
}

func (mgr *CronMgr) Get(tag string) (*Cron, bool) {
	mgr.RLock()
	defer mgr.RUnlock()

	cron, ok := mgr.crons[tag]
	return cron, ok
}

func (mgr *CronMgr) Delete(tag string) {
	mgr.Lock()
	defer mgr.Unlock()

	if cron, ok := mgr.crons[tag]; ok {
		cron.Stop()
		delete(mgr.crons, tag)
	}
}

func (mgr *CronMgr) Clear() {
	mgr.Lock()
	defer mgr.Unlock()

	for _, cron := range mgr.crons {
		cron.Stop()
	}

	mgr.crons = map[string]*Cron{}
}

func NewMgr() *CronMgr {
	return &CronMgr{
		crons: map[string]*Cron{},
	}
}
