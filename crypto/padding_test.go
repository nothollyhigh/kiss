package crypto

import (
	"github.com/nothollyhigh/kiss/util"
	"testing"
)

func TestPKCS5(t *testing.T) {
	unpaddingData := []byte(util.RandString(32))
	paddingData := PKCS5Padding(unpaddingData, 32)
	if string(PKCS5UnPadding(paddingData)) != string(unpaddingData) {
		t.Fatal("padding test failed")
	}
}

func TestZero(t *testing.T) {
	unpaddingData := []byte(util.RandString(32))
	paddingData := ZeroPadding(unpaddingData, 32)
	if string(ZeroUnPadding(paddingData)) != string(unpaddingData) {
		t.Fatal("padding test failed")
	}
}

func TestPKCS7(t *testing.T) {
	unpaddingData := []byte(util.RandString(32))
	paddingData := PKCS7Padding(unpaddingData, 32)
	if string(PKCS7UnPadding(paddingData)) != string(unpaddingData) {
		t.Fatal("padding test failed")
	}
}
