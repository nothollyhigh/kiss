package util

import (
	"testing"
)

func TestHash(t *testing.T) {
	count := make([]int, 128)
	for i := 0; i < 10000000; i++ {
		count[uint64(Hash(RandString(i%32+5)))%uint64(128)]++
	}

	t.Logf("count: %v", count)
}
