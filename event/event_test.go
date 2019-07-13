package event

import (
	"testing"
)

func TestEvent(t *testing.T) {
	event1 := "event1"
	event2 := 2222

	Subscribe("tag1", event1, func(e interface{}, args ...interface{}) {
		t.Logf("tag1, event: %v, args: %v", e, args)
	})

	SubscribeOnce("tag2", event1, func(e interface{}, args ...interface{}) {
		t.Logf("tag2, event: %v, args: %v", e, args)
	})

	Subscribe("tag3", event2, func(e interface{}, args ...interface{}) {
		t.Logf("tag3, event: %v, args: %v", e, args)
	})

	Subscribe("tag_all", EventAll, func(e interface{}, args ...interface{}) {
		t.Logf("tag_all, event: %v, args: %v", e, args)
	})

	t.Logf("--------------------------------------------\nevent1:\n")
	Publish(event1, "arg1", "arg2", 1111)
	t.Logf("--------------------------------------------\nevent1:\n")
	Publish(event1, "arg1", "arg2", 2222)
	t.Logf("--------------------------------------------\nevent1:\n")
	Unsubscribe("tag1")
	Publish(event1, "arg1", "arg2", 3333)
	t.Logf("--------------------------------------------\nevent2:\n")
	Publish(event2, "arg1", "arg2", 2222)
	t.Logf("--------------------------------------------")
}
