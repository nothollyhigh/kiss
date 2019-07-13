#### 一、死锁告警

```golang
package main

import (
	"github.com/nothollyhigh/kiss/sync"
	"time"
)

func main() {
	// 设置死锁告警超时时间为3秒
	sync.SetDebug(true, time.Second/2)

	mtx := sync.Mutex{}
	mtx.Lock()
	go func() {
		mtx.Lock()
	}()
	time.Sleep(time.Second)

	rwmtx := sync.RWMutex{}
	rwmtx.Lock()
	go func() {
		rwmtx.Lock()
	}()

	time.Sleep(time.Second)
}
```
