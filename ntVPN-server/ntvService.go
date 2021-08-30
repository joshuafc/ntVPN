package main

import (
	"context"
	"github.com/joshuafc/ntVPN/network"
	"github.com/smallnest/rpcx/server"
	"net"
)

type NtvService interface {
	ClientReg(ctx context.Context, req *network.ClientRegReq, reply *network.ClientRegReply) error
	ClientUnReg(ctx context.Context, req *network.ClientUnRegReq, reply *network.ClientUnRegReply) error
	ClientQueryOthers(ctx context.Context, req *network.ClientQueryOthersReq, reply *network.ClientQueryOthersReply) error
	ClientConnectTo(ctx context.Context, req *network.ClientConnectToReq, reply *network.ClientConnectToReply) error
	ClientBoardCast(ctx context.Context, req *network.ClientBroadcastReq, reply *network.ClientBroadcastReply) error
}

func NewNtvService() (service NtvService) {
	service = &ntvServiceImpl{}
	return
}

type ntvServiceImpl struct {
	clientsManager *vpnClientsManager
}

func (t *ntvServiceImpl) ClientUnReg(ctx context.Context, req *network.ClientUnRegReq, reply *network.ClientUnRegReply) error {
	panic("implement me")
}

func (t *ntvServiceImpl) ClientQueryOthers(ctx context.Context, req *network.ClientQueryOthersReq, reply *network.ClientQueryOthersReply) error {
	panic("implement me")
}

func (t *ntvServiceImpl) ClientConnectTo(ctx context.Context, req *network.ClientConnectToReq, reply *network.ClientConnectToReply) error {
	panic("implement me")
}

func (t *ntvServiceImpl) ClientBoardCast(ctx context.Context, req *network.ClientBroadcastReq, reply *network.ClientBroadcastReply) error {
	panic("implement me")
}

func (t *ntvServiceImpl) ClientReg(ctx context.Context, req *network.ClientRegReq, reply *network.ClientRegReply) error {
	err := t.clientsManager.AddClient(req.NodeName, req.Ip, req.Mac, ctx.Value(server.RemoteConnContextKey).(net.Conn))
	if err != nil {
		return err
	}
	reply.Ok = true
	return nil
}
