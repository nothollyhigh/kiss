package util

import (
	"fmt"
	"testing"
)

type X struct {
	A int
	B string
}

func TestEncoding(t *testing.T) {
	var (
		a    = map[string]interface{}{"A": 3, "B": "foo"}
		b    = &X{}
		loop = 1000000
	)

	tc := NewTimeCount("msgpack")
	for i := 0; i < loop; i++ {
		MsgpackAToB(a, b)
	}
	tc.Output()
	fmt.Println("msgpack:", b)

	tc = NewTimeCount("json")
	for i := 0; i < loop; i++ {
		JsonAToB(a, b)
	}
	tc.Output()
	fmt.Println("json:", b)
}
