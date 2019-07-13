package sync

import (
	"testing"
	"time"
)

func TestMutex(t *testing.T) {
	SetDebug(true, time.Second/2)

	mtx := Mutex{}
	mtx.Lock()
	go func() {
		mtx.Lock()
	}()
	time.Sleep(time.Second)

	rwmtx := RWMutex{}
	rwmtx.Lock()
	go func() {
		rwmtx.Lock()
	}()

	time.Sleep(time.Second)
}
