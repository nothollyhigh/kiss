package sync

import (
	"fmt"
	"sync"
	"time"
)

// wait session
type WaitSession struct {
	sync.Mutex
	sessions map[interface{}]chan interface{}
}

// add
func (ws *WaitSession) Add(sess interface{}) error {
	ws.Lock()
	if ws.sessions == nil {
		ws.sessions = map[interface{}]chan interface{}{}
	}
	if _, ok := ws.sessions[sess]; ok {
		ws.Unlock()
		return fmt.Errorf("session %v exist", sess)
	}
	ws.sessions[sess] = make(chan interface{}, 1)
	ws.Unlock()

	return nil
}

// done
func (ws *WaitSession) Done(sess interface{}, data interface{}) error {
	ws.Lock()
	if done, ok := ws.sessions[sess]; ok {
		ws.Unlock()
		done <- data
		return nil
	}
	ws.Unlock()

	return fmt.Errorf("session %v not exist", sess)
}

// func (ws *WaitSession) removeSession(sess interface{}) {
// 	ws.Lock()
// 	delete(ws.sessions, sess)
// 	ws.Unlock()
// }

// wait
func (ws *WaitSession) Wait(sess interface{}, timeout time.Duration) (interface{}, error) {
	ws.Lock()
	done, ok := ws.sessions[sess]
	if ok {
		ws.Unlock()

		var data interface{}
		if timeout > 0 {
			select {
			case data = <-done:
			case <-time.After(timeout):
				ws.Lock()
				delete(ws.sessions, sess)
				ws.Unlock()
				return nil, fmt.Errorf("wait session %v timeout", sess)
			}
		} else {
			data = <-done
		}

		ws.Lock()
		delete(ws.sessions, sess)
		ws.Unlock()

		return data, nil
	}

	ws.Unlock()

	return nil, fmt.Errorf("session %v not exist", sess)
}

// session length
func (ws *WaitSession) Len() int {
	ws.Lock()
	l := len(ws.sessions)
	ws.Unlock()
	return l
}
