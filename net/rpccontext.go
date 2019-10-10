package net

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/vmihailenco/msgpack"
)

var (
	ErrInvalidBody = errors.New("invalid body")
)

// rpc context
type RpcContext struct {
	method  string
	client  *TcpClient
	message IMessage
}

// tcp client
func (ctx *RpcContext) Client() *TcpClient {
	return ctx.client
}

// cmd
func (ctx *RpcContext) Cmd() uint32 {
	return ctx.message.Cmd()
}

// rpc body
func (ctx *RpcContext) Body() []byte {
	return ctx.message.Body()
}

// rpc method
func (ctx *RpcContext) Method() string {
	return ctx.method
}

// write data
func (ctx *RpcContext) WriteData(data []byte) error {
	//case client.chSend <- asyncMessage{msg.Encrypt(client.SendSeq(), client.SendKey(), client.cipher), nil}:
	msg := NewRpcMessage(ctx.message.Cmd(), ctx.message.Ext(), data)
	data = msg.Encrypt(ctx.client.SendSeq(), ctx.client.SendKey(), ctx.client.cipher)
	return ctx.client.pushDataSync(data)
}

// write message
func (ctx *RpcContext) WriteMsg(msg IMessage) error {
	if ctx.message != msg {
		msg.SetExt(ctx.message.Ext())
	}
	data := msg.Encrypt(ctx.client.SendSeq(), ctx.client.SendKey(), ctx.client.cipher)
	return ctx.client.pushDataSync(data)
}

// bind data
func (ctx *RpcContext) Bind(v interface{}) error {
	return DefaultCodec.Unmarshal(ctx.Body(), v)
}

// write data marshal by default codec
func (ctx *RpcContext) Write(v interface{}) error {
	data, err := DefaultCodec.Marshal(v)
	if err != nil {
		return err
	}
	msg := NewRpcMessage(ctx.message.Cmd(), ctx.message.Ext(), data)
	data = msg.Encrypt(ctx.client.SendSeq(), ctx.client.SendKey(), ctx.client.cipher)
	return ctx.client.pushDataSync(data)
}

// bind json
func (ctx *RpcContext) BindJson(v interface{}) error {
	return json.Unmarshal(ctx.Body(), v)
}

// write json data
func (ctx *RpcContext) WriteJson(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	msg := NewRpcMessage(ctx.message.Cmd(), ctx.message.Ext(), data)
	data = msg.Encrypt(ctx.client.SendSeq(), ctx.client.SendKey(), ctx.client.cipher)
	return ctx.client.pushDataSync(data)
}

// bind gob data
func (ctx *RpcContext) BindGob(v interface{}) error {
	return gob.NewDecoder(bytes.NewBuffer(ctx.Body())).Decode(v)
}

// write gob data
func (ctx *RpcContext) WriteGob(v interface{}) error {
	buffer := &bytes.Buffer{}
	err := gob.NewEncoder(buffer).Encode(v)
	if err != nil {
		return err
	}
	msg := NewRpcMessage(ctx.message.Cmd(), ctx.message.Ext(), buffer.Bytes())
	data := msg.Encrypt(ctx.client.SendSeq(), ctx.client.SendKey(), ctx.client.cipher)
	return ctx.client.pushDataSync(data)
}

// bind msgpack data
func (ctx *RpcContext) BindMsgpack(v interface{}) error {
	return msgpack.Unmarshal(ctx.Body(), v)
}

// write msgpack data
func (ctx *RpcContext) WriteMsgpack(v interface{}) error {
	data, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	msg := NewRpcMessage(ctx.message.Cmd(), ctx.message.Ext(), data)
	data = msg.Encrypt(ctx.client.SendSeq(), ctx.client.SendKey(), ctx.client.cipher)
	return ctx.client.pushDataSync(data)
}

// bind protobuf data
func (ctx *RpcContext) BindProtobuf(v proto.Message) error {
	return proto.Unmarshal(ctx.Body(), v)
}

// write protobuf data
func (ctx *RpcContext) WriteProtobuf(v proto.Message) error {
	data, err := proto.Marshal(v)
	if err != nil {
		return err
	}
	msg := NewRpcMessage(ctx.message.Cmd(), ctx.message.Ext(), data)
	data = msg.Encrypt(ctx.client.SendSeq(), ctx.client.SendKey(), ctx.client.cipher)
	return ctx.client.pushDataSync(data)
}

// write error
func (ctx *RpcContext) Error(errText string) error {
	msg := NewRpcMessage(CmdRpcError, ctx.message.Ext(), []byte(errText))
	data := msg.Encrypt(ctx.client.SendSeq(), ctx.client.SendKey(), ctx.client.cipher)
	return ctx.client.pushDataSync(data)
}
