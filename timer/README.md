#### 一、定时器

```golang
package main

import (
	"fmt"
	"github.com/nothollyhigh/kiss/timer"
	"time"
)

func main() {
	t1 := time.Now()
	<-timer.After(time.Second)
	fmt.Println("timer.After 111 sec:", time.Since(t1).Seconds())

	t2 := time.Now()
	timer.Once(time.Second, func() {
		fmt.Println("timer.Once after 222 sec:", time.Since(t2).Seconds())
	})
	time.Sleep(time.Second)

	t3 := time.Now()
	timer.AfterFunc(time.Second, func() {
		fmt.Println("timer.AfterFunc after 333 sec:", time.Since(t3).Seconds())
	})
	time.Sleep(time.Second)

	t4 := time.Now()
	timer.Schedule(time.Second, time.Second, 3, func() {
		fmt.Println("timer.Schedule after 444 sec:", time.Since(t4).Seconds())
		t4 = time.Now()
	})
	time.Sleep(time.Second * 4)

	t5 := time.Now()
	item := timer.Schedule(time.Second, time.Second, 3, func() {
		fmt.Println("timer.Schedule 555 cancel after 1.5 s, sec:", time.Since(t5).Seconds())
		t5 = time.Now()
	})
	time.Sleep(time.Second / 2 * 3)
	item.Cancel()

	time.Sleep(time.Second * 5)
	fmt.Println("over")
}
```