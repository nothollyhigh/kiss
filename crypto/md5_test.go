package crypto

import (
	"testing"
)

func TestMd5(t *testing.T) {
	if string(MD5Encrypt("00000000")) != "dd4b21e9ef71e1291183a46b913ae6f2" {
		t.Fatal("md5 test failed")
	}
}
