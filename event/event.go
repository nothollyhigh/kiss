package event

import (
	"fmt"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"sync"
)

var (
	EventAll = "*"

	DefaultEvent *Event = New(nil)
)

// event handler
type eventHandler struct {
	once    bool
	handler func(evt interface{}, args ...interface{})
}

// event management
type Event struct {
	sync.Mutex
	listenerMap map[interface{}]interface{}
	listeners   map[interface{}]map[interface{}]eventHandler
}

// safe call event handler
func handle(handler func(evt interface{}, args ...interface{}), evt interface{}, args ...interface{}) {
	defer util.HandlePanic()
	handler(evt, args...)
}

// subscribe
func (eventMgr *Event) Subscribe(tag interface{}, evt interface{}, handler func(evt interface{}, args ...interface{})) error {
	eventMgr.Lock()
	defer eventMgr.Unlock()

	if _, ok := eventMgr.listenerMap[tag]; ok {
		log.Debug("Subscribe failed: tag %v exist!", tag)
		return fmt.Errorf("[Event %v] exist", tag)
	}

	eventMgr.listenerMap[tag] = evt
	if eventMgr.listeners[evt] == nil {
		eventMgr.listeners[evt] = make(map[interface{}]eventHandler)
	}
	eventMgr.listeners[evt][tag] = eventHandler{false, handler}

	return nil
}

// subscribe once
func (eventMgr *Event) SubscribeOnce(tag interface{}, evt interface{}, handler func(evt interface{}, args ...interface{})) error {
	eventMgr.Lock()
	defer eventMgr.Unlock()

	if _, ok := eventMgr.listenerMap[tag]; ok {
		log.Debug("SubscribeOnce failed: tag %v exist!", tag)
		return fmt.Errorf("[Event %v] exist", tag)
	}

	eventMgr.listenerMap[tag] = evt
	if eventMgr.listeners[evt] == nil {
		eventMgr.listeners[evt] = make(map[interface{}]eventHandler)
	}
	eventMgr.listeners[evt][tag] = eventHandler{true, handler}

	return nil
}

// unsubscribe without lock
func (eventMgr *Event) UnsubscribeWithoutLock(tag interface{}) {
	if evt, ok := eventMgr.listenerMap[tag]; ok {
		delete(eventMgr.listenerMap, tag)
		delete(eventMgr.listeners[evt], tag)
		if len(eventMgr.listeners[evt]) == 0 {
			delete(eventMgr.listeners, evt)
		}
	}
}

// unsubscribe
func (eventMgr *Event) Unsubscribe(tag interface{}) {
	eventMgr.Lock()
	defer eventMgr.Unlock()

	if evt, ok := eventMgr.listenerMap[tag]; ok {
		delete(eventMgr.listenerMap, tag)
		delete(eventMgr.listeners[evt], tag)
		if len(eventMgr.listeners[evt]) == 0 {
			delete(eventMgr.listeners, evt)
		}
	}
}

// publish
func (eventMgr *Event) Publish(evt interface{}, args ...interface{}) {
	eventMgr.Lock()
	defer eventMgr.Unlock()

	/*
		els := eventMgr.listeners[evt]
		aels := eventMgr.listeners[EventAll]
		all := make([]eventHandler, len(els)+len(aels))
		i := 0
		for _, l := range els {
			all[i] = l
			i++
		}
		for _, l := range aels {
			all[i] = l
			i++
		}

		eventMgr.Unlock()

		for _, l := range all {
			eventHandler(l, evt, args)
		}
	*/
	if listeners, ok := eventMgr.listeners[evt]; ok {
		for tag, listener := range listeners {
			handle(listener.handler, evt, args...)
			if listener.once {
				delete(eventMgr.listenerMap, tag)
				delete(listeners, tag)
				if len(listeners) == 0 {
					delete(eventMgr.listeners, evt)
				}
			}
		}
	}
	if listeners, ok := eventMgr.listeners[EventAll]; ok {
		for tag, listener := range listeners {
			handle(listener.handler, evt, args...)
			if listener.once {
				delete(listeners, tag)
			}
		}
	}
}

// subscribe
func Subscribe(tag interface{}, evt interface{}, handler func(evt interface{}, args ...interface{})) error {
	return DefaultEvent.Subscribe(tag, evt, handler)
}

// subscribe once
func SubscribeOnce(tag interface{}, evt interface{}, handler func(evt interface{}, args ...interface{})) error {
	return DefaultEvent.SubscribeOnce(tag, evt, handler)
}

// unsubscribe without lock
func UnsubscribeWithoutLock(tag interface{}) {
	DefaultEvent.UnsubscribeWithoutLock(tag)
}

// unsubscribe
func Unsubscribe(tag interface{}) {
	DefaultEvent.Unsubscribe(tag)
}

// publish
func Publish(evt interface{}, args ...interface{}) {
	DefaultEvent.Publish(evt, args...)
}

// event factory
func New(tag interface{}) *Event {
	return &Event{
		listenerMap: make(map[interface{}]interface{}),
		listeners:   make(map[interface{}]map[interface{}]eventHandler),
	}
}
