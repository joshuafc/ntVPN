package network

import (
	"bytes"
	"encoding/gob"
	"net"
)

// VPNCommandType represents the command type
type VPNCommandType int32

// VPNCommandType enum
const (
	ClientRegisterRequest VPNCommandType = iota
	ClientRegisterResponse
	ClientUnRegisterRequest
	ClientUnRegisterResponse
	ClientQueryOthersRequest
	ClientQueryOthersResponse
	ClientConnectToRequest
	ClientConnectToResponse
	ServerConnectToRequest
	ServerConnectToResponse
)

//func ReadMsg(context context.Context, conn net.Conn) (VPNCommandMsg,error) {
//	decoder := gob.NewDecoder(conn)
//	for {
//		if context.Err() != nil {
//			break
//		}
//		msg := VPNCommandMsg{}
//		decoder.Decode(&msg)
//	}
//
//}

type VPNCommandMsg struct {
	CommandType VPNCommandType
	Payload     []byte
	Padding     []byte // add padding to make command large than 32 byte to satisfy kcp protocol
}

func FromDetailMsg(msgType VPNCommandType, msgData interface{}) (msg *VPNCommandMsg) {
	msg = new(VPNCommandMsg)
	msg.CommandType = msgType
	var buffer bytes.Buffer
	err := gob.NewEncoder(&buffer).Encode(msgData)
	if err != nil {
		panic(err)
	}
	msg.Payload = buffer.Bytes()
	payloadLen := len(msg.Payload)
	if payloadLen < 32 {
		msg.Padding = make([]byte, 32-payloadLen)
	}
	return
}

func (v *VPNCommandMsg) ToDetailMsg(msg interface{}) error {
	payloadBuf := bytes.NewBuffer(v.Payload)
	decoder := gob.NewDecoder(payloadBuf)
	return decoder.Decode(msg)
}

func ReadCommand(conn *net.Conn) (msg *VPNCommandMsg, err error) {
	msg = new(VPNCommandMsg)
	err = gob.NewDecoder(*conn).Decode(msg)
	return
}

func WriteCommand(conn *net.Conn, msg *VPNCommandMsg) (err error) {
	err = gob.NewEncoder(*conn).Encode(msg)
	return
}

type ClientRegisterRequestMsg struct {
	Ip       string
	Mac      string
	Mask     string
	UserName string
	Password string
}

type ClientRegisterResponseMsg struct {
	Ok bool
}

type ClientUnRegisterRequestMsg struct {
	Ok bool
}

type ClientUnRegisterResponseMsg struct {
	Ok bool
}

type ClientQueryOthersRequestMsg struct {
	Ok bool
}

type ClientQueryOthersResponseMsg struct {
	Ok bool
}

type ClientConnectToRequestMsg struct {
	Ok bool
}

type ClientConnectToResponseMsg struct {
	Ok bool
}

type ServerConnectToRequestMsg struct {
	Ok bool
}

type ServerConnectToResponseMsg struct {
	Ok bool
}
