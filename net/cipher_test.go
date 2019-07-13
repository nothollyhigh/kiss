package net

import (
	"testing"
)

func TestCipherGzip(t *testing.T) {
	cipher := NewCipherGzip(DefaultThreshold)
	str := ""
	for i := 0; i < 256; i++ {
		str += "abcdefghij"
	}
	msg := NewMessage(1, []byte(str))
	data := cipher.Encrypt(0, 0, msg.data)
	data2, err := cipher.Decrypt(0, 0, data)
	if err != nil {
		t.Logf("TestCipherGzip failed: %v", err)
	}
	if string(msg.data) != string(data2) {
		t.Logf("TestCipherGzip failed: not equal")
	}
	data = cipher.Encrypt(0, 0, msg.data)
	data2, err = cipher.Decrypt(0, 0, data)
	if err != nil {
		t.Logf("TestCipherGzip failed: %v", err)
	}
	if string(msg.data) != string(data2) {
		t.Logf("TestCipherGzip failed: not equal")
	}
}
