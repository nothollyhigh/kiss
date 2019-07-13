package util

import (
	"fmt"
	"github.com/vmihailenco/msgpack"
	"math/rand"
	"time"
	"unsafe"
)

var (
	numsBytes   = "1234567890"
	letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func BytesToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func StrToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func RandBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}

func RandCodeString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = numsBytes[rand.Intn(len(numsBytes))]
	}
	return string(b)
}

func ShuffleString(s string) string {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]byte, len(s))
	perm := r.Perm(len(s))
	for i, randIndex := range perm {
		ret[i] = s[randIndex]
	}
	return string(ret)
}

func ShuffleBytes(b []byte) []byte {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]byte, len(b))
	perm := r.Perm(len(b))
	for i, randIndex := range perm {
		ret[i] = b[randIndex]
	}
	return ret
}

func Float2String(v interface{}) map[string]interface{} {
	data, _ := msgpack.Marshal(v)
	m := map[string]interface{}{}
	msgpack.Unmarshal(data, &m)
	for k, vv := range m {
		if fv, ok := vv.(float32); ok {
			m[k] = fmt.Sprintf("%.2f", fv)
			continue
		}
		if fv, ok := vv.(float64); ok {
			m[k] = fmt.Sprintf("%.2f", fv)
			continue
		}
		if mv, ok := vv.(map[string]interface{}); ok {
			m[k] = Float2String(mv)
		}
		if mv, ok := vv.(map[interface{}]interface{}); ok {
			m[k] = Float2String(mv)
		}
		if av, ok := vv.([]interface{}); ok {
			for i, avv := range av {
				av[i] = Float2String(avv)
			}
			m[k] = av
		}
	}
	return m
}
