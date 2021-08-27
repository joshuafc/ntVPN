package main

import (
	"crypto/sha1"
	"flag"
	"net"

	"github.com/op/go-logging"
	"github.com/smallnest/rpcx/server"
	kcp "github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
)

var (
	addr = flag.String("addr", "localhost:8972", "server address")
	cryptKey = flag.String("cryptKey", "rpcx-key", "cryptKey")
	cryptSalt = flag.String("cryptSalt", "rpcx-salt", "cryptSalt")
	log = logging.MustGetLogger("ntvd")
	format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfile} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
)

func main() {
	logging.SetFormatter(format)

	flag.Parse()

	pass := pbkdf2.Key([]byte(*cryptKey), []byte(*cryptSalt), 4096, 32, sha1.New)
	bc, err := kcp.NewAESBlockCrypt(pass)
	if err != nil {
		panic(err)
	}

	s := server.NewServer(server.WithBlockCrypt(bc), server.WithReadTimeout(60*1e9), server.WithWriteTimeout(60*1e9))
	err = s.Register(new(NtvService), "")
	if err != nil {
		log.Fatal(err.Error())
	}

	cs := &ConfigUDPSession{}
	s.Plugins.Add(cs)

	err = s.Serve("kcp", *addr)
	if err != nil {
		log.Fatal(err.Error())
	}
}

type ConfigUDPSession struct{}

func (p *ConfigUDPSession) HandleConnAccept(conn net.Conn) (net.Conn, bool) {
	session, ok := conn.(*kcp.UDPSession)
	if !ok {
		return conn, true
	}

	session.SetACKNoDelay(true)
	session.SetStreamMode(true)
	return conn, true
}
