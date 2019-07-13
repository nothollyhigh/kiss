package net

import (
	"errors"
	"fmt"
)

// errors
var (
	ErrTcpClientIsStopped       = errors.New("tcp client is stopped")
	ErrTcpClientWriteHalf       = errors.New("tcp client write half")
	ErrTcpClientSendQueueIsFull = errors.New("tcp client's send queue is full")

	ErrRpcClientIsDisconnected  = errors.New("rpc client disconnected")
	ErrRpcClientSendQueueIsFull = errors.New("rpc client's send queue is full")
	ErrRpcCallTimeout           = errors.New("rpc call timeout")
	ErrRpcCallClientError       = errors.New("rpc client error")

	ErrorRpcInvalidPbMessage = errors.New("invalid pb Message")

	ErrorBroadcastNotEnabled = errors.New("broadcast not enabled")

	ErrorReservedCmdInternal  = fmt.Errorf("cmd > %d/0x%X is reserved for internal, plz use other number", CmdUserMax, CmdUserMax)
	ErrorReservedCmdPing      = fmt.Errorf("cmd %d/0x%X is reserved for ping, plz use other number", CmdPing, CmdPing)
	ErrorReservedCmdSetRealip = fmt.Errorf("cmd %d/0x%X is reserved for set client's real ip, plz use other number", CmdSetReaIp, CmdSetReaIp)
	ErrorReservedCmdRpcMethod = fmt.Errorf("cmd %d/0x%X is reserved for rpc method, plz use other number", CmdRpcMethod, CmdRpcMethod)
	ErrorReservedCmdRpcError  = fmt.Errorf("cmd %d/0x%X is reserved for rpc method error, plz use other number", CmdRpcError, CmdRpcError)

	ErrWSClientIsStopped       = errors.New("websocket client is stopped")
	ErrWSClientSendQueueIsFull = errors.New("websocket client's send queue is full")
	ErrClientWithoutCodec      = errors.New("websocket client has no codec")
	ErrWSEngineShutdownTimeout = errors.New("shutdown timeout")
)
