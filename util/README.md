#### 一、异常处理

```golang
package main

import (
	"fmt"
	"github.com/nothollyhigh/kiss/util"
	"sync"
	"time"
)

func test() {
	defer util.HandlePanic()
	panic("test panic")
}

func main() {
	test()

	util.Safe(func() {
		panic("closure panic")
	})

	wg := sync.WaitGroup{}
	wg.Add(1)
	util.Go(func() {
		defer wg.Done()
		panic("goroutine closure panic")
	})
	wg.Wait()
	time.Sleep(time.Second / 10)
	fmt.Println("over")
}
```

#### 二、任务池

```golang
package main

import (
	"fmt"
	"github.com/nothollyhigh/kiss/util"
	"math/rand"
	"sync"
	"time"
)

func main() {
	workers := util.NewWorkers("test", 10, 5)
	wg := &sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		idx := i
		wg.Add(1)
		workers.Go(func() {
			defer wg.Done()
			time.Sleep(time.Second / time.Duration(10*(1+rand.Intn(10))))
			fmt.Println(idx)
		}, 0)
	}
	wg.Wait()

	fmt.Println("-----")
	for i := 0; i < 20; i++ {
		idx := i
		go func() {
			wg.Add(1)
			defer wg.Done()
			err := workers.GoWait(func() {
				time.Sleep(time.Second / 2)
			})
			fmt.Println(idx, err)
		}()
	}
	wg.Wait()

	workers.Stop()
	fmt.Println("over")
}
```

#### 三、有序任务池

```golang
package main

import (
	"fmt"
	"github.com/nothollyhigh/kiss/util"
	"sync"
	"time"
)

func main() {
	wg := sync.WaitGroup{}
	cl := util.NewWorkersLink("test", 10, 20)
	for i := 1; i <= 50; i++ {
		idx := i
		wg.Add(1)
		cl.Go(func(task *util.LinkTask) {
			defer wg.Done()

			task.WaitPre()

			fmt.Println("---", idx)
			if idx%20 == 0 {
				fmt.Println("+++ sleep", idx)
				time.Sleep(time.Second)
			}

			task.Done(nil)
		})

	}
	wg.Wait()
	cl.Stop()
}
```

