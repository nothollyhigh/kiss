package crypto

import (
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"
)

func TripleDesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	origData = PKCS5Padding(origData, block.BlockSize())
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:8])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func TripleDesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, key[:8])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return origData, nil
}

func TripleDesB64Encrypt(data, key string) (result string, err error) {
	var encrypt []byte
	encrypt, err = TripleDesEncrypt([]byte(data), []byte(key))
	if err != nil {
		return
	}
	result = base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(encrypt)
	return
}

func TripleDesB64Decrypt(encrypt, key string) (result string, err error) {
	var data []byte
	data, err = base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(encrypt)
	if err != nil {
		return
	}
	var decrypt []byte
	decrypt, err = TripleDesDecrypt(data, []byte(key))
	if err != nil {
		return
	}
	result = string(decrypt)
	return
}
