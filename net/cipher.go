package net

import (
	"encoding/binary"
	"github.com/nothollyhigh/kiss/util"
)

const (
	CipherGzipAll  = 0
	CipherGzipNone = -1

	DefaultThreshold = CipherGzipNone
)

// cipher interface
type ICipher interface {
	Init()
	Encrypt(seq int64, key uint32, data []byte) []byte
	Decrypt(seq int64, key uint32, data []byte) ([]byte, error)
}

// default cipher with gzip
type CipherGzip struct {
	threshold int
}

func (cipher *CipherGzip) Init() {

}

// encrypt message
func (cipher *CipherGzip) Encrypt(seq int64, key uint32, data []byte) []byte {
	if len(data) < DEFAULT_MESSAGE_HEAD_LEN || cipher.threshold < 0 || (len(data) <= cipher.threshold+DEFAULT_MESSAGE_HEAD_LEN) {
		return data
	}
	body := util.GZipCompress(data[DEFAULT_MESSAGE_HEAD_LEN:])
	newData := make([]byte, DEFAULT_MESSAGE_HEAD_LEN+len(body))
	copy(newData[:DEFAULT_MESSAGE_HEAD_LEN], data[:DEFAULT_MESSAGE_HEAD_LEN])
	copy(newData[DEFAULT_MESSAGE_HEAD_LEN:], body)
	cmd := binary.LittleEndian.Uint32(newData[DEFAULT_CMD_IDX_BEGIN:DEFAULT_CMD_IDX_END])
	binary.LittleEndian.PutUint32(newData[DEFAULT_CMD_IDX_BEGIN:DEFAULT_CMD_IDX_END], cmd|CmdFlagMaskGzip)
	binary.LittleEndian.PutUint32(newData[DEFAULT_BODY_LEN_IDX_BEGIN:DEFAULT_BODY_LEN_IDX_END], uint32(len(body)))
	return newData
}

// decrypt message
func (cipher *CipherGzip) Decrypt(seq int64, key uint32, data []byte) ([]byte, error) {
	if len(data) < DEFAULT_MESSAGE_HEAD_LEN {
		return nil, ErrorRpcInvalidMessageHeadLen
	}
	cmd := binary.LittleEndian.Uint32(data[DEFAULT_CMD_IDX_BEGIN:DEFAULT_CMD_IDX_END])
	if cmd&CmdFlagMaskGzip != CmdFlagMaskGzip {
		return data, nil
	}
	binary.LittleEndian.PutUint32(data[DEFAULT_CMD_IDX_BEGIN:DEFAULT_CMD_IDX_END], cmd&(^CmdFlagMaskGzip))
	body, err := util.GZipUnCompress(data[DEFAULT_MESSAGE_HEAD_LEN:])
	if err == nil {
		newData := make([]byte, DEFAULT_MESSAGE_HEAD_LEN+len(body))
		copy(newData, data[:DEFAULT_MESSAGE_HEAD_LEN])
		copy(newData[DEFAULT_MESSAGE_HEAD_LEN:], body)
		binary.LittleEndian.PutUint32(newData[DEFAULT_BODY_LEN_IDX_BEGIN:DEFAULT_BODY_LEN_IDX_END], uint32(len(body)))
		return newData, nil
	}

	return nil, err
}

func NewCipherGzip(threshold int) ICipher {
	return &CipherGzip{threshold}
}
