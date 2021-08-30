package network

import (
	"net"
)

type IceSDP struct {
	frag       string
	pwd        string
	candidates []string
}

// VPNCommandType represents the command type
type VPNCommandType int32

type ClientRegReq struct {
	NodeName string
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
	targetMac string
	selfSdp   IceSDP
}

// ClientConnectToReply when rpc server reply this message, target vpn node is ready for idc.dail
type ClientConnectToReply struct {
	targetSdp IceSDP
	Ok        bool
}

type ClientOnConnectToReq struct {
	targetMac string
	selfSdp   IceSDP
}

type ClientOnConnectToReply struct {
	Ok bool
}

type ServerConnectToNotify struct {
	fromMac   string
	targetMac string
	fromSdp   IceSDP
	Ok        bool
}

type ServerBroadcastNotify struct {
	fromMac string
	payload []byte
}

type ClientBroadcastReq struct {
	fromMac string
	payload []byte
}

type ClientBroadcastReply struct {
	Ok bool
}
