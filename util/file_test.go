package util

import (
	"testing"
)

func TestPathExist(t *testing.T) {
	if !PathExist("./file.go") {
		t.Fatal("PahtExist error")
	} else {
		t.Log("./file.go exist")
	}
}
