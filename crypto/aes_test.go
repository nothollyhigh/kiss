package crypto

import (
	"bytes"
	"github.com/nothollyhigh/kiss/util"
	"testing"
)

func TestAes(t *testing.T) {
	key := util.RandBytes(32)
	iv := util.RandBytes(16)
	data := []byte("test data")
	encrypted, err := AesCBCEncrypt(key, iv, data)
	if err != nil {
		t.Fatalf("AesCBCEncrypt failed: %v", err)
	}

	decrypted, err := AesCBCDecrypt(key, iv, encrypted)
	if err != nil {
		t.Fatalf("AesCBCDecrypt failed: %v", err)
	}

	if !bytes.Equal(data, decrypted) {
		t.Fatalf("AesCBCDecrypt failed, not equal: %v != %v", string(decrypted), string(data))
	}

}
