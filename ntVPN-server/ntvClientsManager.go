package main



type VpnClientsManager interface {
	AddClient(name string, vpnIp string, vpnMac string) error
	DelClientByName(clientName string) error
}

func NewVpnClientsManager() (manager VpnClientsManager) {
	manager = &vpnClientsManager{}
	return
}

type vpnClientsManager struct {

}

func (v *vpnClientsManager) AddClient(name string, vpnIp string, vpnMac string) error {
	panic("implement me")
}

func (v *vpnClientsManager) DelClientByName(clientName string) error {
	panic("implement me")
}




