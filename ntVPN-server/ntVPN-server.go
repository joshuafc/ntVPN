package main

import (
	"context"
	"crypto/sha1"
	"encoding/gob"
	"errors"
	"flag"
	"github.com/joshuafc/ntVPN/network"
	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"log"
	"net"
	"strconv"
	"sync"
)

func main() {
	gob.Register(&net.UDPAddr{})
	serverContext, serverQuit := context.WithCancel(context.Background())
	defer serverQuit()
	var clientManager ClientManager
	clientManager.ConnMap = make(map[net.Addr]*ClientItem)
	clientManager.IpMap = make(map[string]*ClientItem)
	clientManager.MacMap = make(map[string]*ClientItem)
	clientManager.mutex = new(sync.Mutex)
	serverPort := flag.Int("port", 3455, "specify ntVPN server port")
	encryptKey := flag.String("encry_key", "password", "specify ntVPN transfer encrypt password")
	flag.Parse()

	key := pbkdf2.Key([]byte(*encryptKey), []byte(*encryptKey), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)
	if listener, err := kcp.ListenWithOptions("0.0.0.0:"+strconv.Itoa(*serverPort), block, 10, 3); err == nil {
		for {
			s, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go clientManager.handleConnection(serverContext, s)
		}
	} else {
		log.Fatal(err)
	}
}

type ClientItem struct {
	conn                    net.Conn
	writeLock               *sync.Mutex
	clientContext           context.Context
	clientContextCancelFunc func()
	clientInfo              network.ClientRegisterRequestMsg
	readChan                chan *network.VPNCommandMsg
	writeChan               chan *network.VPNCommandMsg
}

func (clientItem *ClientItem) ProcessMsg(msg *network.VPNCommandMsg) error {
	return nil
}

type ClientManager struct {
	ConnMap map[net.Addr]*ClientItem
	IpMap   map[string]*ClientItem
	MacMap  map[string]*ClientItem
	mutex   *sync.Mutex
}

func (clientManager *ClientManager) P2PRequest(toClientItem *ClientItem) error {
	clientManager.mutex.Lock()
	defer clientManager.mutex.Unlock()

	return nil
}

func (clientManager *ClientManager) P2PResponse(toClientItem *ClientItem) error {
	clientManager.mutex.Lock()
	defer clientManager.mutex.Unlock()
	return nil
}

func (clientManager *ClientManager) AddClient(clientItem *ClientItem) error {
	clientManager.mutex.Lock()
	defer clientManager.mutex.Unlock()
	clientManager.ConnMap[clientItem.conn.RemoteAddr()] = clientItem
	return nil
}

func (clientManager *ClientManager) UpdateClient(clientItem *ClientItem) error {
	clientManager.mutex.Lock()
	defer clientManager.mutex.Unlock()

	ip_val, ip_ok := clientManager.IpMap[clientItem.clientInfo.Ip]
	mac_val, mac_ok := clientManager.MacMap[clientItem.clientInfo.Mac]

	if ip_ok && ip_val.conn.RemoteAddr() != clientItem.conn.RemoteAddr() {
		return errors.New("Error: Ip addr Used!")
	}

	if mac_ok && mac_val.conn.RemoteAddr() != clientItem.conn.RemoteAddr() {
		return errors.New("Error: Mac addr Used!")
	}

	clientManager.IpMap[clientItem.clientInfo.Ip] = clientItem
	clientManager.MacMap[clientItem.clientInfo.Mac] = clientItem
	return nil
}

func (clientManager *ClientManager) DeleteClient(clientItem *ClientItem) error {
	clientManager.mutex.Lock()
	defer clientManager.mutex.Unlock()
	if clientItem.clientInfo.Ip != "" {
		delete(clientManager.IpMap, clientItem.clientInfo.Ip)
	}
	if clientItem.clientInfo.Mac != "" {
		delete(clientManager.MacMap, clientItem.clientInfo.Mac)
	}
	delete(clientManager.ConnMap, clientItem.conn.RemoteAddr())
	return nil
}

func (clientManager *ClientManager) GetClients(selfAddr net.Addr) (clientInfo []network.ClientInfo) {
	clientInfo = make([]network.ClientInfo, 0)
	clientManager.mutex.Lock()
	defer clientManager.mutex.Unlock()
	for k, v := range clientManager.ConnMap {
		if k == selfAddr {
			continue
		}
		clientInfo = append(clientInfo, network.ClientInfo{
			Ip:   v.clientInfo.Ip,
			Mac:  v.clientInfo.Mac,
			Mask: v.clientInfo.Mask,
			Addr: k,
		})
	}
	return
}

func (clientManager *ClientManager) handleConnection(serverContext context.Context, s net.Conn) {
	var clientItem ClientItem
	clientItem.writeLock = new(sync.Mutex)
	clientItem.clientContext, clientItem.clientContextCancelFunc = context.WithCancel(serverContext)
	clientItem.conn = s

	clientManager.AddClient(&clientItem)

	clientItem.readChan = make(chan *network.VPNCommandMsg, 5)
	clientItem.writeChan = make(chan *network.VPNCommandMsg, 5)

	defer func() {
		clientItem.clientContextCancelFunc()
		clientManager.DeleteClient(&clientItem)
		close(clientItem.writeChan)
	}()

	go func() {
		for {
			msg, err := network.ReadCommand(s)
			if err != nil && clientItem.clientContext.Err() != nil {
				close(clientItem.readChan)
				break
			}
			clientItem.readChan <- msg
		}
	}()

	go func() {
		for {
			select {
			case cmd, ok := <-clientItem.writeChan:
				if !ok {
					break
				}
				err := network.WriteCommand(s, cmd)
				if err != nil {
					s.Close()
					break
				}
			}
		}
	}()

Exit:
	for {
		select {
		case cmd, ok := <-clientItem.readChan:
			if !ok {
				break
			}
			switch true {
			case cmd.CommandType == network.ClientRegisterRequest:
				err := cmd.ToDetailMsg(&clientItem.clientInfo)
				if err != nil {
					break
				}
				err1 := clientManager.UpdateClient(&clientItem)
				var response network.ClientRegisterResponseMsg
				response.Ok = err1 == nil
				clientItem.writeChan <- network.FromDetailMsg(network.ClientRegisterResponse, &response)
				if !response.Ok {
					break Exit
				}
			case cmd.CommandType == network.ClientUnRegisterRequest:
				err := cmd.ToDetailMsg(&clientItem.clientInfo)
				if err != nil {
					break
				}
				clientManager.DeleteClient(&clientItem)
				var response network.ClientUnRegisterResponseMsg
				response.Ok = true
				clientItem.writeChan <- network.FromDetailMsg(network.ClientUnRegisterResponse, &response)

			case cmd.CommandType == network.ClientQueryOthersRequest:
				var response network.ClientQueryOthersResponseMsg
				response.Clients = clientManager.GetClients(s.RemoteAddr())
				clientItem.writeChan <- network.FromDetailMsg(network.ClientQueryOthersResponse, &response)
			default:
				err := clientItem.ProcessMsg(cmd)
				if err != nil {
					break
				}
			}
		case <-clientItem.clientContext.Done():
			break
		}
	}
}
