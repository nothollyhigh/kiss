package util

import (
	"testing"
)

func TestHandlePanic(t *testing.T) {
	func() {
		defer HandlePanic()
		panic("throw err for TestHandlePanic")
	}()
}

func TestSafe(t *testing.T) {
	Safe(func() {
		defer HandlePanic()
		panic("throw err for TestSafe")
	})
	t.Logf("GetProcName(): %s", GetProcName())
}
