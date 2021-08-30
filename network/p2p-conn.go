package network

import (
	"context"
	"github.com/pion/ice/v2"
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
	agent     *ice.Agent
	conn      *ice.Conn
	remoteSDP *IceSDP
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
	return i.agent.Accept(context.TODO(), i.remoteSDP.frag, i.remoteSDP.pwd)
}

func (i *IceAgent) Dial() (*ice.Conn, error) {
	if i.agent == nil {
		return &ice.Conn{}, nil
	}
	return i.agent.Dial(context.TODO(), i.remoteSDP.frag, i.remoteSDP.pwd)
}

func (i *IceAgent) GetLocalSDP() (IceSDP, error) {
	var localSDP IceSDP
	localSDP.frag, localSDP.pwd, _ = i.agent.GetLocalUserCredentials()
	candidateChan := make(chan ice.Candidate)
	i.agent.OnCandidate(func(c ice.Candidate) {
		if c == nil {
			close(candidateChan)
		} else {
			candidateChan <- c
		}
	})
	i.agent.GatherCandidates()
	for {
		candidate, ok := <-candidateChan
		if !ok || candidate == nil {
			break
		}
		i.remoteSDP.candidates = append(i.remoteSDP.candidates, candidate.Marshal())
	}
	return localSDP, nil
}

func (i *IceAgent) SetRemoteSDP(sdp IceSDP) error {
	*i.remoteSDP = sdp
	for _, v := range i.remoteSDP.candidates {
		candidate, _ := ice.UnmarshalCandidate(v)
		i.agent.AddRemoteCandidate(candidate)
	}
	return nil
}
