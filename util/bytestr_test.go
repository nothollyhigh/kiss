package util

import (
	"errors"
	"testing"
)

func TestStrToBytes(t *testing.T) {
	for i := 1; i <= 100; i++ {
		s := RandString(i)
		if news := string(StrToBytes(s)); news != s {
			t.Fatal(errors.New("StrToBytes failed"))
		}
	}
}

func TestRandString(t *testing.T) {
	for i := 1; i <= 100; i++ {
		if slen := len(RandString(i)); slen != i {
			t.Fatal(errors.New("invalid RandString len"))
		}
	}
}
