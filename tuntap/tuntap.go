package tuntap

import (
	"github.com/mdlayher/ethernet"
)

type TapDevice interface {
	SetHardwareAddr(addr string) error
	GetHardwareAddr() (string, error)

	SetIpAddr(addr, mask string) error
	GetIpAddr() (string, string, error)

	Up() error
	Down() error

	Read() (ethernet.Frame, error)
	Write(frame ethernet.Frame) error
}
