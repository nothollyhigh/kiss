#### 一、模块

```golang
package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/nothollyhigh/kiss/graceful"
)

var (
	module = &MyModule{}
)

type MyModule struct {
	graceful.Module
}

func (m *MyModule) Start() {
	m.Module.Start()
	fmt.Println("MyModule Start()")
}

func (m *MyModule) Stop() {
	fmt.Println("MyModule Stop()")
}

func (m *MyModule) Print_1() {
	fmt.Println("m.Print_1")
}

func (m *MyModule) Print_2() {
	fmt.Println("m.Print_2")
}

func (m *MyModule) Print_3() {
	fmt.Println("m.Print_3")
}

func main() {
	// 每个module.Start()启动一个逻辑协程
	graceful.Register(module)
	graceful.Start()

	go func() {
		time.Sleep(time.Second / 10)
		for {
			time.Sleep(time.Second)
			// 多个协程调用module.Exec都会发送到module的逻辑协程chan中串行执行以避免锁操作
			module.Exec(module.Print_1)
		}
	}()

	go func() {
		time.Sleep(time.Second / 10 * 2)
		for {
			time.Sleep(time.Second)
			module.Exec(module.Print_2)
		}
	}()

	go func() {
		time.Sleep(time.Second / 10 * 3)
		for {
			time.Sleep(time.Second)
			module.Exec(module.Print_3)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan)
	<-sigChan

	graceful.Stop()
}
```
