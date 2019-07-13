package util

import (
	"github.com/nothollyhigh/kiss/log"
	"sync/atomic"
	"time"
)

type Qps struct {
	Tag   string
	Born  time.Time
	Start time.Time
	Count int64
	Total int64

	ticker *time.Ticker
}

func (q *Qps) Add(n int64) {
	atomic.AddInt64(&q.Count, n)
	atomic.AddInt64(&q.Total, n)
}

func (q *Qps) Run(args ...interface{}) *Qps {
	interval := time.Second
	if len(args) > 0 {
		if i, ok := args[0].(int); ok {
			interval = interval * time.Duration(i)
		}
	}

	Go(func() {
		q.ticker = time.NewTicker(interval)
		for {
			q.Start = time.Now()
			if _, ok := <-q.ticker.C; !ok {
				log.Info("[qps %v] over", q.Tag)
				return
			}
			total := atomic.LoadInt64(&q.Total)
			log.Info("[qps %v]: %v / s | avg: %v / s | total: %v for %v s",
				q.Tag, atomic.SwapInt64(&q.Count, 0)/int64(interval/time.Second), int64(float64(total)/time.Since(q.Born).Seconds()), total, int64(time.Since(q.Born).Seconds()))
		}
	})

	return q
}

func (q *Qps) Stop() {
	q.ticker.Stop()
}

func NewQps(tag string) *Qps {
	return &Qps{
		Tag:   tag,
		Born:  time.Now(),
		Count: 0,
		Total: 0,
	}
}
