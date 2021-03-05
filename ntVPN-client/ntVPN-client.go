package main

import (
	"crypto/sha1"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/joshuafc/ntVPN/network"
	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

func main() {
	gob.Register(&net.UDPAddr{})

	serverAddress := flag.String("server", "127.0.0.1", "specify ntVPN server IP address")
	serverPort := flag.Int("port", 3455, "specify ntVPN server port")
	encryptKey := flag.String("encry_key", "password", "specify ntVPN transfer encrypt password")

	vpnIP := flag.String("vpn_ip", "192.168.101.2", "specify vpn node ip address")
	vpnMask := flag.String("vpn_mask", "255.255.255.0", "specify vpn node mask")
	vpnMac := flag.String("vpn_mac", "", "specify vpn node mac address")

	readChan := make(chan *network.VPNCommandMsg, 5)
	writeChan := make(chan *network.VPNCommandMsg, 5)

	flag.Parse()

	if len(*vpnMac) == 0 {
		buf := make([]byte, 6)
		_, err := rand.Read(buf)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		// Set the local bit
		buf[0] |= 2
		*vpnMac = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
	}

	key := pbkdf2.Key([]byte(*encryptKey), []byte(*encryptKey), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)

	if sess, err := kcp.DialWithOptions(*serverAddress+":"+strconv.Itoa(*serverPort), block, 10, 3); err == nil {
		var conn net.Conn
		conn = sess
		defer conn.Close()

		go func() {
			for {
				msg, err := network.ReadCommand(conn)
				if err != nil {
					break
				}
				readChan <- msg
			}
		}()

		go func() {
			for {
				select {
				case cmd, ok := <-writeChan:
					if !ok {
						break
					}
					err := network.WriteCommand(conn, cmd)
					if err != nil {
						break
					}
				}
			}
		}()

		timer1 := time.NewTimer(0)
		defer timer1.Stop()
		go func() {
			for range timer1.C {
				var msg network.ClientRegisterRequestMsg
				msg.Ip = *vpnIP
				msg.Mac = *vpnMac
				msg.Mask = *vpnMask
				writeChan <- network.FromDetailMsg(network.ClientRegisterRequest, msg)
				timer1.Reset(time.Second * 30)
			}
		}()

		timer2 := time.NewTimer(time.Second * 2)
		defer timer2.Stop()
		for {
			select {
			case cmd, ok := <-readChan:
				if !ok {
					return
				}
				switch {
				case cmd.CommandType == network.ClientRegisterResponse:
					var detailMsg network.ClientRegisterResponseMsg
					err = cmd.ToDetailMsg(&detailMsg)
					if !detailMsg.Ok {
						log.Print("Register Failed!")
					}
					log.Println(fmt.Sprintf("%+v", detailMsg))
				case cmd.CommandType == network.ClientQueryOthersResponse:
					var detailMsg network.ClientQueryOthersResponseMsg
					err = cmd.ToDetailMsg(&detailMsg)
					log.Println(fmt.Sprintf("%+v", detailMsg))
				default:
					log.Print("Data Not Process.")
				}
			case <-timer2.C:
				timer2.Reset(time.Second * 2)
				var msg network.ClientQueryOthersRequestMsg
				writeChan <- network.FromDetailMsg(network.ClientQueryOthersRequest, &msg)
			}
		}
	}
}
