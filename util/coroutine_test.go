package util

import (
	"testing"
)

func TestGo(t *testing.T) {
	Go(func() {
		panic("err")
	})
}
