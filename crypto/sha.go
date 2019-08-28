package crypto

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

func HmacSha1(message, secret string) string {
	hash := hmac.New(sha1.New, []byte(secret))
	hash.Write([]byte(message))
	return hex.EncodeToString(hash.Sum(nil))
}

func Sha1(message string) string {
	h := sha1.New()
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func HmacSha256(message, secret string) string {
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(message))
	return hex.EncodeToString(hash.Sum(nil))
}

func Sha256(message string) string {
	h := sha256.New()
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
