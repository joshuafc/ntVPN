package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"github.com/joshuafc/ntVPN/network"
	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"log"
	"net"
	"strconv"
)

func main() {
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
			go handleClient(s)
		}
	} else {
		log.Fatal(err)
	}
}

func handleClient(s net.Conn) {
	for {
		cmd, err := network.ReadCommand(&s)
		if err != nil {
			return
		}
		switch {
		case cmd.CommandType == network.ClientRegisterRequest:
			var detailMsg network.ClientRegisterRequestMsg
			err = cmd.ToDetailMsg(&detailMsg)
			fmt.Println("%+v", detailMsg)
		}
	}
}
