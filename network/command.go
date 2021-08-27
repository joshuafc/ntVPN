package network

import (
	"net"
)

// VPNCommandType represents the command type
type VPNCommandType int32

type ClientRegReq struct {
	Ip       string
	Mac      string
	Mask     string
}

type ClientRegReply struct {
	Ok bool
}

type ClientUnRegReq struct {
	Ok bool
}

type ClientUnRegReply struct {
	Ok bool
}

type ClientQueryOthersReq struct {
}

type ClientInfo struct {
	Ip   string
	Mac  string
	Mask string
	Addr net.Addr
}

type ClientQueryOthersReply struct {
	Clients []ClientInfo
}

type ClientConnectToReq struct {
	Ok bool
}

type ClientConnectToReply struct {
	Ok bool
}

type ServerConnectToReq struct {
	Ok bool
}

type ServerConnectToReply struct {
	Ok bool
}
