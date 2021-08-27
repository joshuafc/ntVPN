package main

type VpnClient interface {

}

func NewVpnClient() (vpnClient VpnClient) {
	vpnClient = &vpnClientImpl{}
	return
}

type vpnClientImpl struct {


}