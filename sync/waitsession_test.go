package sync

import (
	"testing"
	"time"
)

func TestWaitSession(t *testing.T) {
	var ws = WaitSession{}
	var session interface{} = 1

	err := ws.Add(session)
	if err != nil {
		t.Fatalf("Add session 1 failed: %v", err)
	}

	go func() {
		time.Sleep(time.Second / 1000)
		data := 2
		ws.Done(session, data)
	}()

	data, err := ws.Wait(session, time.Second*3)
	if err != nil {
		t.Fatalf("session 1 err: %v", err)
	}
	if data == nil {
		t.Fatalf("session data should be 2 got nil")
	}

	err = ws.Add(session)
	if err != nil {
		t.Fatalf("Add session 1 failed: %v", err)
	}

	err = ws.Add(1)
	if err == nil {
		t.Fatalf("Add session 1 should failed for exist")
	}

	session = 2
	err = ws.Add(session)
	if err != nil {
		t.Fatalf("Add session 2 failed: %v", err)
	}

	session = 1
	data, err = ws.Wait(session, time.Second/10)
	if err == nil {
		t.Fatalf("session 1 should timeout got nil")
	}
	if data != nil {
		t.Fatalf("data should be nil got %v", data)
	}

	session = 2
	data, err = ws.Wait(2, time.Second/10)
	if err == nil {
		t.Fatalf("session 2 should timeout got nil")
	}
	if data != nil {
		t.Fatalf("data should be nil got %v", data)
	}
}
