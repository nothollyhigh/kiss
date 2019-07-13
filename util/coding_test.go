package util

import (
	"testing"
)

func TestUTF8GBK(t *testing.T) {
	utf8Str := "测试"
	gbkStr := UTF8ToGBK(utf8Str)
	utf8Str2 := GBKToUTF8(gbkStr)
	if utf8Str != utf8Str2 {
		t.Fatal(utf8Str2)
	}
}
