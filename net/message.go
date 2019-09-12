package net

import (
	"encoding/binary"
)

// tcp async message for send queue
type asyncMessage struct {
	data []byte
	cb   func(*TcpClient, error)
}

// websocket async message for send queue
type wsAsyncMessage struct {
	data []byte
	cb   func(*WSClient, error)
}

const (
	// default message header length
	DEFAULT_MESSAGE_HEAD_LEN int = 16

	// default body length begin idx
	DEFAULT_BODY_LEN_IDX_BEGIN int = 0
	// default body length end idx
	DEFAULT_BODY_LEN_IDX_END int = 4

	// default cmd begin idx
	DEFAULT_CMD_IDX_BEGIN int = 4
	// default cmd end idx
	DEFAULT_CMD_IDX_END int = 8

	// default extension begin idx
	DEFAULT_EXT_IDX_BEGIN int = 8
	// default extension end idx
	DEFAULT_EXT_IDX_END int = 16

	// default gzip cipher flag mask
	CmdFlagMaskGzip = uint32(1) << 31

	// reserved cmd: ping
	CmdPing = uint32(0x1 << 24)
	// reserved cmd: ping2
	CmdPing2 = uint32(0x1<<24 + 1)
	// reserved cmd: set real ip
	CmdSetReaIp = uint32(0x1<<24 + 2)
	// reserved cmd: rpc method
	CmdRpcMethod = uint32(0x1<<24 + 3)
	// reserved cmd: rpc error
	CmdRpcError = uint32(0x1<<24 + 4)

	// max user space cmd
	CmdUserMax = uint32(0xFFFFFF)
)

// message interface
type IMessage interface {
	// message header length
	HeadLen() int
	// message body length
	BodyLen() int

	// message cmd
	Cmd() uint32
	// setting message cmd
	SetCmd(cmd uint32)

	// message extension
	Ext() uint64
	// setting message extension
	SetExt(ext uint64)

	// message rpc sequence
	RpcSeq() int64
	// setting message rpc sequence
	SetRpcSeq(seq int64)

	// all message data
	Data() []byte
	// setting all message data
	SetData(data []byte)

	// all message raw data
	RawData() []byte
	// setting all message raw data
	SetRawData(rawData []byte)

	// message body
	Body() []byte
	// setting message body
	SetBody(body []byte)

	// encrypt message
	Encrypt(seq int64, key uint32, cipher ICipher) []byte
	// decrypt message
	Decrypt(seq int64, key uint32, cipher ICipher) ([]byte, error)
}

// default message implementation
type Message struct {
	data    []byte
	rawData []byte
}

// header length
func (msg *Message) HeadLen() int {
	return DEFAULT_MESSAGE_HEAD_LEN
}

// body length
func (msg *Message) BodyLen() int {
	return int(binary.LittleEndian.Uint32(msg.data[DEFAULT_BODY_LEN_IDX_BEGIN:DEFAULT_BODY_LEN_IDX_END]))
}

// cmd
func (msg *Message) Cmd() uint32 {
	return binary.LittleEndian.Uint32(msg.data[DEFAULT_CMD_IDX_BEGIN:DEFAULT_CMD_IDX_END])
}

// setting cmd
func (msg *Message) SetCmd(cmd uint32) {
	binary.LittleEndian.PutUint32(msg.data[DEFAULT_CMD_IDX_BEGIN:DEFAULT_CMD_IDX_END], cmd)
}

// extension
func (msg *Message) Ext() uint64 {
	return binary.LittleEndian.Uint64(msg.data[DEFAULT_EXT_IDX_BEGIN:DEFAULT_EXT_IDX_END])
}

// setting extension
func (msg *Message) SetExt(ext uint64) {
	binary.LittleEndian.PutUint64(msg.data[DEFAULT_EXT_IDX_BEGIN:DEFAULT_EXT_IDX_END], ext)
}

// rpc sequence
func (msg *Message) RpcSeq() int64 {
	return int64(binary.LittleEndian.Uint64(msg.data[DEFAULT_EXT_IDX_BEGIN:DEFAULT_EXT_IDX_END]))
}

// setting rpc sequence
func (msg *Message) SetRpcSeq(seq int64) {
	binary.LittleEndian.PutUint64(msg.data[DEFAULT_EXT_IDX_BEGIN:DEFAULT_EXT_IDX_END], uint64(seq))
}

// all data
func (msg *Message) Data() []byte {
	return msg.data
}

// setting all data
func (msg *Message) SetData(data []byte) {
	msg.data = data
}

// raw data
func (msg *Message) RawData() []byte {
	return msg.rawData
}

// setting raw data
func (msg *Message) SetRawData(rawData []byte) {
	msg.rawData = rawData
}

// body
func (msg *Message) Body() []byte {
	return msg.data[DEFAULT_MESSAGE_HEAD_LEN:]
}

// setting body
func (msg *Message) SetBody(data []byte) {
	needLen := DEFAULT_MESSAGE_HEAD_LEN + len(data) - len(msg.data)
	if needLen > 0 {
		msg.data = append(msg.data, make([]byte, needLen)...)
	} else if needLen < 0 {
		msg.data = msg.data[:DEFAULT_MESSAGE_HEAD_LEN+len(data)]
	}
	copy(msg.data[DEFAULT_MESSAGE_HEAD_LEN:], data)
	binary.LittleEndian.PutUint32(msg.data[DEFAULT_BODY_LEN_IDX_BEGIN:DEFAULT_BODY_LEN_IDX_END], uint32(len(data)))
}

// encrypt message
func (msg *Message) Encrypt(seq int64, key uint32, cipher ICipher) []byte {
	if cipher != nil {
		msg.rawData = cipher.Encrypt(seq, key, msg.data)
	} else {
		msg.rawData = msg.data
	}
	return msg.rawData
}

// decrypt message
func (msg *Message) Decrypt(seq int64, key uint32, cipher ICipher) ([]byte, error) {
	var err error
	if cipher != nil {
		msg.data, err = cipher.Decrypt(seq, key, msg.rawData)
	} else {
		msg.data = msg.rawData
	}
	return msg.data, err
}

// message factory
func NewMessage(cmd uint32, data []byte) *Message {
	msg := &Message{
		data: make([]byte, len(data)+DEFAULT_MESSAGE_HEAD_LEN),
	}
	binary.LittleEndian.PutUint32(msg.data[DEFAULT_CMD_IDX_BEGIN:DEFAULT_CMD_IDX_END], cmd)
	binary.LittleEndian.PutUint32(msg.data[DEFAULT_BODY_LEN_IDX_BEGIN:DEFAULT_BODY_LEN_IDX_END], uint32(len(data)))
	if len(data) > 0 {
		copy(msg.data[DEFAULT_MESSAGE_HEAD_LEN:], data)
	}
	return msg
}

// message factory by data
func RawMessage(data []byte) *Message {
	return &Message{
		data: data,
	}
}

// rpc message factory
func NewRpcMessage(cmd uint32, seq int64, data []byte) *Message {
	msg := &Message{
		data: make([]byte, len(data)+DEFAULT_MESSAGE_HEAD_LEN),
	}
	binary.LittleEndian.PutUint32(msg.data[DEFAULT_CMD_IDX_BEGIN:DEFAULT_CMD_IDX_END], cmd)
	binary.LittleEndian.PutUint64(msg.data[DEFAULT_EXT_IDX_BEGIN:DEFAULT_EXT_IDX_END], uint64(seq))
	binary.LittleEndian.PutUint32(msg.data[DEFAULT_BODY_LEN_IDX_BEGIN:DEFAULT_BODY_LEN_IDX_END], uint32(len(data)))
	if len(data) > 0 {
		copy(msg.data[DEFAULT_MESSAGE_HEAD_LEN:], data)
	}
	return msg
}

// real ip message
func RealIpMsg(ip string) *Message {
	// ret := strings.Split(ip, ".")
	// if len(ret) == 4 {
	// 	var err error
	// 	var ipv int
	// 	var data = make([]byte, 4)
	// 	for i, v := range ret {
	// 		ipv, err = strconv.Atoi(v)
	// 		if err != nil || ipv > 255 {
	// 			return nil
	// 		}
	// 		data[i] = byte(ipv)
	// 	}
	// 	return NewMessage(CmdSetReaIp, data)
	// }
	return NewMessage(CmdSetReaIp, []byte(ip))
}

// ping message
func PingMsg() *Message {
	return NewMessage(CmdPing, nil)
}
