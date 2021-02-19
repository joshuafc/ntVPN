package network

import (
	"context"
	"github.com/pion/ice"
)

/*
	1. Controller GetLocalSDP
	2. Controller Launch Connect Request
	3. Controlee Accept Connect Request
	4. Controlee GetLocalSDP
	5. Controlee SetRemoteSDP
	6. Controlee Accept ice.Conn
	7. Controlee Reply Connect Request
	8. Controller Set RemoteSDP
    9. Controller Dial ice.Conn
*/

type IceAgent struct {
	agent       *ice.Agent
	conn        *ice.Conn
	remoteUfrag string
	remotePwd   string
}

func NewIceAgent(stunAddr string, turnAddr string, turnUser string, turnPwd string) (*IceAgent, error) {
	var iceAgent IceAgent
	var err error
	stunURL, _ := ice.ParseURL(stunAddr)
	turnURL, _ := ice.ParseURL(turnAddr)
	turnURL.Username = turnUser
	turnURL.Password = turnPwd
	iceAgent.agent, err = ice.NewAgent(&ice.AgentConfig{
		NetworkTypes: []ice.NetworkType{ice.NetworkTypeUDP4},
		Urls:         []*ice.URL{stunURL, turnURL},
		CandidateTypes: []ice.CandidateType{
			ice.CandidateTypeHost,
			ice.CandidateTypeServerReflexive,
			ice.CandidateTypePeerReflexive,
			ice.CandidateTypeRelay,
		},
	})
	return &iceAgent, err
}

func (i *IceAgent) Accept() (*ice.Conn, error) {
	if i.agent == nil {
		return &ice.Conn{}, nil
	}
	return i.agent.Accept(context.TODO(), i.remoteUfrag, i.remotePwd)
}

func (i *IceAgent) Dial() (*ice.Conn, error) {
	if i.agent == nil {
		return &ice.Conn{}, nil
	}
	return i.agent.Dial(context.TODO(), i.remoteUfrag, i.remotePwd)
}

func (i *IceAgent) GetLocalSDP() (string, error) {
	return "", nil
}

func (i *IceAgent) SetRemoteSDP(sdp string) error {
	return nil
}
