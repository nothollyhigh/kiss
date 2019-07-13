package timer

import (
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"
)

//go test -v
//go test -bench="."

func TestTimer(t *testing.T) {
	var deviation int64
	var num int64 = 10000
	sleepTime := time.Millisecond * 2000

	wg := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(i int) {
			start := time.Now().UnixNano()
			<-After(sleepTime)
			end := time.Now().UnixNano()
			d := (end - start) - int64(sleepTime)
			if d >= 0 {
				deviation += d
			} else {
				deviation -= d
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	t.Logf("avg deviation: %v ms", math.Round(float64(deviation)/float64(num)/1000000))
}

func BenchmarkTimer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		After(time.Millisecond * time.Duration(rand.Intn(10000)+100))
	}
}

func BenchmarkTimerStd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		time.After(time.Millisecond * 100)
	}
}
