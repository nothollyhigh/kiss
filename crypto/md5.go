package crypto

import (
	"crypto/md5"
	"encoding/hex"
)

func MD5Encrypt(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
