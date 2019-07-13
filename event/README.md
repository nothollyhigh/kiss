#### 一、发布订阅

```golang
package main

import (
	"fmt"
	"github.com/nothollyhigh/kiss/event"
)

func main() {
	tag := "tag"
	evt := "evt"

	event.Subscribe(tag, evt, func(e interface{}, args ...interface{}) {
		fmt.Printf("receive event: %v, args: %v\n", e, args)
	})
	fmt.Println("publish event: evt, data: args1, args2, 1111")
	event.Publish(evt, "arg1", "arg2", 1111)
	event.Unsubscribe(tag)
	fmt.Println("publish event: evt, data: args1, args2, 2222")
	event.Publish(evt, "arg1", "arg2", 2222)
}
```