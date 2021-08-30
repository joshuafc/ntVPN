package main

import (
	"context"
	"crypto/sha1"
	"flag"
	"fmt"
	"github.com/joshuafc/ntVPN/network"
	"github.com/op/go-logging"
	"github.com/smallnest/rpcx/client"
	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"math/rand"
	"net"
	"time"
)

var (
	addr      = flag.String("addr", "localhost:8972", "server address")
	cryptKey  = flag.String("cryptKey", "rpcx-key", "cryptKey")
	cryptSalt = flag.String("cryptSalt", "rpcx-salt", "cryptSalt")
	ip        = flag.String("vpn_ip", "192.168.101.2", "specify vpn node ip address")
	mask      = flag.String("vpn_mask", "255.255.255.0", "specify vpn node mask")
	mac       = flag.String("vpn_mac", "", "specify vpn node mac address")

	format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfile} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)

	log = logging.MustGetLogger("ntvc")
)

func main() {
	logging.SetFormatter(format)

	flag.Parse()

	if len(*mac) == 0 {
		buf := make([]byte, 6)
		_, err := rand.Read(buf)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		// Set the local bit
		buf[0] |= 2
		*mac = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
	}

	pass := pbkdf2.Key([]byte(*cryptKey), []byte(*cryptSalt), 4096, 32, sha1.New)
	bc, _ := kcp.NewAESBlockCrypt(pass)
	option := client.DefaultOption
	option.Block = bc
	option.ConnectTimeout = 30 * 1e9
	option.Heartbeat = true
	option.MaxWaitForHeartbeat = 30 * 1e9
	option.HeartbeatInterval = 30 * 1e9

	d, _ := client.NewPeer2PeerDiscovery("kcp@"+*addr, "")
	xclient := client.NewXClient("NtvService", client.Failtry, client.RoundRobin, d, option)
	defer xclient.Close()

	// plugin
	cs := &ConfigUDPSession{}
	pc := client.NewPluginContainer()
	pc.Add(cs)
	xclient.SetPlugins(pc)

	args := network.ClientRegReq{}

	start := time.Now()
	for i := 0; i < 10000; i++ {
		reply := &network.ClientRegReply{}
		err := xclient.Call(context.Background(), "Mul", args, reply)
		if err != nil {
			log.Fatalf("failed to call: %v", err)
		}
		//log.Printf("%d * %d = %d", args.A, args.B, reply.C)
	}
	dur := time.Since(start)
	qps := 10000 * 1000 / int(dur/time.Millisecond)
	fmt.Printf("qps: %d call/s", qps)
}

type ConfigUDPSession struct{}

func (p *ConfigUDPSession) ConnCreated(conn net.Conn) (net.Conn, error) {
	session, ok := conn.(*kcp.UDPSession)
	if !ok {
		return conn, nil
	}

	session.SetACKNoDelay(true)
	session.SetStreamMode(true)
	return conn, nil
}
