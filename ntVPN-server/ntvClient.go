package main

import "net"

type VpnClient interface {
	GetName() string
	GetIp() string
	GetMac() string
	GetConn() net.Conn
}

func NewVpnClient(name string, ip string, mac string, conn net.Conn) (vpnClient VpnClient) {
	vpnClient = &vpnClientImpl{}
	return
}

type vpnClientImpl struct {
	name string
	ip   string
	mac  string
	conn net.Conn
}

func (v *vpnClientImpl) GetMac() string {
	return v.mac
}

func (v *vpnClientImpl) GetName() string {
	return v.name
}

func (v *vpnClientImpl) GetIp() string {
	return v.ip
}

func (v *vpnClientImpl) GetConn() net.Conn {
	return v.conn
}
