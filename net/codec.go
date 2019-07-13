package net

import (
	"bytes"
	"encoding/gob"
	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/vmihailenco/msgpack"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	DefaultCodec        = json
	DefaultRpcCodecType = "json"
)

// codec interface
type ICodec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	// Type() string
}

// gob codec
type CodecGob struct{}

// marshal
func (c *CodecGob) Marshal(v interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	err := gob.NewEncoder(buffer).Encode(v)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// unmarshal
func (c *CodecGob) Unmarshal(data []byte, v interface{}) error {
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}

// json codec
type CodecJson struct{}

// marshal
func (c *CodecJson) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// unmarshal
func (c *CodecJson) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// msgpack codec
type CodecMsgpack struct{}

// marshal
func (c *CodecMsgpack) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// unmarshal
func (c *CodecMsgpack) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

// protobuf codec
type CodecProtobuf struct{}

// marshal
func (c *CodecProtobuf) Marshal(v interface{}) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if ok {
		return proto.Marshal(msg)
	}
	return nil, ErrorRpcInvalidPbMessage
}

// unmarshal
func (c *CodecProtobuf) Unmarshal(data []byte, v interface{}) error {
	msg, ok := v.(proto.Message)
	if ok {
		return proto.Unmarshal(data, msg)
	}
	return ErrorRpcInvalidPbMessage
}
