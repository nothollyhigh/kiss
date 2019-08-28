package cron

import (
	"fmt"
	"testing"
	"time"
)

func TestCron(t *testing.T) {
	tag := "TestCron"
	async := true
	c := New(tag, "3,5,12,27,39,40,55-59 * * * * * *", async, func() {
		fmt.Println("cron task:", time.Now().Format("2006-01-02 15:04:05"))
	})

	c.Start()
	time.Sleep(time.Second * 60)
	c.Stop()
}

func TestCronMgr(t *testing.T) {
	tag := "TestCronMgr"
	async := true
	mgr := NewMgr()
	mgr.Add(tag, "3,5,12,27,39,40,55-59 * * * * * *", async, func() {
		fmt.Println("cron mgr task:", time.Now().Format("2006-01-02 15:04:05"))
	})

	time.Sleep(time.Second * 60)
	mgr.Clear()
}
