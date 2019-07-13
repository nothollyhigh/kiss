package util

import (
	"fmt"
	"sync"
	"time"
)

var _timeIdMtx = &sync.Mutex{}

func TimeID(n int) (string, error) {
	_timeIdMtx.Lock()
	defer _timeIdMtx.Unlock()
	nano := time.Now().UnixNano()
	ids := fmt.Sprintf("%v", nano)
	if n > len(ids) {
		return "", fmt.Errorf("too long")
	}
	if n < len(ids) {
		sleep := int64(1)
		for i := 0; i < len(ids)-n; i++ {
			sleep *= 10
		}
		nano /= sleep
		time.Sleep(time.Duration(sleep + 9))
		ids = fmt.Sprintf("%v", nano)
	} else {
		time.Sleep(9)
	}
	return ids, nil
}

func TestTimeID() {
	wg := sync.WaitGroup{}
	emtx := &sync.Mutex{}
	exists := map[string]bool{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				id, _ := TimeID(19)
				emtx.Lock()
				if _, ok := exists[id]; ok {
					panic(id)
				}
				exists[id] = true
				emtx.Unlock()

				fmt.Println(len(id), id)
			}
		}()
	}
	wg.Wait()
}
