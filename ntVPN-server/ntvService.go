package main

import (
	"context"
	"github.com/joshuafc/ntVPN/network"
)

type NtvService interface {
	ClientReg(ctx context.Context, req *network.ClientRegReq, reply *network.ClientRegReply) error

}

func NewNtvService() (service NtvService) {
	service = &ntvServiceImpl{}
	return
}

type ntvServiceImpl struct {
	clientsManager *vpnClientsManager
}

func (t *ntvServiceImpl) ClientReg(ctx context.Context, req *network.ClientRegReq, reply *network.ClientRegReply) error {
	reply.Ok = true
	return nil
}