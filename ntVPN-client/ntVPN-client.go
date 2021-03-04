package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"github.com/joshuafc/ntVPN/network"
	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"net"
	"strconv"
)

func main() {
	serverAddress := flag.String("server", "127.0.0.1", "specify ntVPN server IP address")
	serverPort := flag.Int("port", 3455, "specify ntVPN server port")
	encryptKey := flag.String("encry_key", "password", "specify ntVPN transfer encrypt password")
	flag.Parse()

	key := pbkdf2.Key([]byte(*encryptKey), []byte(*encryptKey), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)

	if sess, err := kcp.DialWithOptions(*serverAddress+":"+strconv.Itoa(*serverPort), block, 10, 3); err == nil {
		var conn net.Conn
		conn = sess

		var msg network.ClientRegisterRequestMsg
		msg.Ip = "192.168.100.4"
		network.WriteCommand(&conn, network.FromDetailMsg(network.ClientRegisterRequest, &msg))
		handleServer(conn)
	}
}

func handleServer(s net.Conn) {
	for {
		cmd, err := network.ReadCommand(&s)
		if err != nil {
			return
		}
		switch {
		case cmd.CommandType == network.ClientRegisterRequest:
			var detailMsg network.ClientRegisterRequestMsg
			err = cmd.ToDetailMsg(detailMsg)
			fmt.Println("%+v", detailMsg)
		}
	}
}
