package util

import (
	"testing"
	"time"
)

func TestTimeCount(t *testing.T) {
	tc := NewTimeCount("TestTimeCount")
	defer tc.Warn(time.Second)
	time.Sleep(time.Second * 2)
}
