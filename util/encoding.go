package util

import (
	// "fmt"
	"github.com/json-iterator/go"
	"github.com/vmihailenco/msgpack"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

//json
func JsonAToB(a, b interface{}) error {
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, b)
}

//msgpack
func MsgpackAToB(a, b interface{}) error {
	data, err := msgpack.Marshal(a)
	if err != nil {
		return err
	}

	return msgpack.Unmarshal(data, b)
}
