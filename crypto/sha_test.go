package crypto

import (
	"testing"
)

func TestHmacSha1(t *testing.T) {
	if string(HmacSha1("00000000", "abcdefg")) != "ceae827404d938a73557c1bd5b496d4f3b1dbd79" {
		t.Fatal("md5 test failed")
	}
}
