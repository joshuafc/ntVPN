package main

import (
	"errors"
	"net"
)

type VpnClientsManager interface {
	AddClient(name string, vpnIp string, vpnMac string, conn net.Conn) error
	DelClientByConn(conn net.Conn) error
}

func NewVpnClientsManager() (manager VpnClientsManager) {
	var managerImpl *vpnClientsManager = new(vpnClientsManager)
	managerImpl.macMap = make(map[string]VpnClient)
	managerImpl.ipMap = make(map[string]VpnClient)
	managerImpl.connMap = make(map[net.Conn]VpnClient)
	manager = managerImpl
	return
}

type vpnClientsManager struct {
	connMap map[net.Conn]VpnClient
	macMap  map[string]VpnClient
	ipMap   map[string]VpnClient
}

func (v *vpnClientsManager) AddClient(name string, vpnIp string, vpnMac string, conn net.Conn) error {
	vpnClient := NewVpnClient(name, vpnIp, vpnMac, conn)
	v.ipMap[vpnIp] = vpnClient
	v.macMap[vpnMac] = vpnClient
	v.connMap[conn] = vpnClient
	return nil
}

func (v *vpnClientsManager) DelClientByConn(conn net.Conn) error {
	vpnClient := v.connMap[conn]
	if vpnClient == nil {
		return errors.New("conn not exists")
	}
	delete(v.ipMap, vpnClient.GetIp())
	delete(v.macMap, vpnClient.GetMac())
	delete(v.connMap, vpnClient.GetConn())
	return nil
}
