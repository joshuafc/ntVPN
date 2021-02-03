package tuntap

import (
	"github.com/mdlayher/ethernet"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

type TapLinux struct {
	fd int
	io.ReadWriteCloser
	devName string
}

const (
	cIFFTUN        = 0x0001
	cIFFTAP        = 0x0002
	cIFFNOPI       = 0x1000
	cIFFMULTIQUEUE = 0x0100
)

type ifReq struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

func ioctl(fd uintptr, request uintptr, argp uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(request), argp)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	return nil
}

func CreateTapDevice(deviceName string) (TapDevice, error) {
	if len(deviceName) > 0x9 {
		panic("deviceName too len!")
	}

	var tapLinux TapLinux
	tapLinux.devName = deviceName
	var err error
	tapLinux.fd, err = syscall.Open(
		"/dev/net/tun", os.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}

	var req ifReq
	req.Flags = cIFFNOPI | cIFFTAP | cIFFMULTIQUEUE
	copy(req.Name[:], deviceName)

	err = ioctl(uintptr(tapLinux.fd), syscall.TUNSETIFF, uintptr(unsafe.Pointer(&req)))
	if err != nil {
		panic(err.Error())
	}
	tapLinux.ReadWriteCloser = os.NewFile(uintptr(tapLinux.fd), "tun")
	return &tapLinux, nil
}

func DestroyTapDevice(tapDevice TapDevice) error {
	syscall.Close(tapDevice.(TapLinux).fd)
	return nil
}

func (t TapLinux) SetHardwareAddr(addr string) error {
	isDeviceUp := true
	command := "ip link show " + t.devName + " | grep \"state DOWN\""
	cmd := exec.Command("sh", "-c", command)
	output, _ := cmd.Output()
	if len(output) != 0 {
		isDeviceUp = false
	}
	var err error
	if isDeviceUp {
		cmd = exec.Command("ip", "link", "set", "dev", t.devName, "down")
		output, err = cmd.Output()
		println(output)
		if err != nil {
			return err
		}
	}
	cmd = exec.Command("ip", "link", "set", "dev", t.devName, "address", addr)
	output, err = cmd.Output()
	println(output)
	if err != nil {
		return err
	}
	if isDeviceUp {
		cmd = exec.Command("ip", "link", "set", "dev", t.devName, "up")
		output, err = cmd.Output()
		println(output)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t TapLinux) GetHardwareAddr() (string, error) {
	command := "ip link show dev " + t.devName + " | grep \"link/ether\" | awk '{print $2}'"
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	println(output)
	return strings.Trim(string(output), "\n"), err
}

func (t TapLinux) SetIpAddr(addr, mask string) error {
	ipaddr := net.ParseIP(mask).To4()
	sz_,_ := net.IPv4Mask(ipaddr[0],ipaddr[1],ipaddr[2],ipaddr[3]).Size()
	cidr := addr + "/" + strconv.Itoa(sz_)
	cmd := exec.Command("ip", "addr", "flush", "dev", t.devName)
	output, err := cmd.Output()
	println(output)
	if err != nil {
		return err
	}
	cmd = exec.Command("ip", "addr", "add", cidr, "dev", t.devName)
	output, err = cmd.Output()
	println(output)
	if err != nil {
		return err
	}
	return nil
}

func (t TapLinux) GetIpAddr() (string, string, error) {
	command := "ip addr show dev " + t.devName + " | grep 'inet ' | awk '{print $2}'"
	cmd := exec.Command("sh", "-c", command)
	cidr_byte, err := cmd.Output()
	cidr := strings.Trim(string(cidr_byte), "\n")
	if err != nil {
		return "","",err
	}
	ip, ipnet, err1 := net.ParseCIDR(cidr)
	if err1 != nil {
		return "","",err
	}
	return ip.String(), (net.IP)(ipnet.Mask).String(), err1
}

func (t TapLinux) Up() error {
	cmd := exec.Command("ip", "link", "set", "dev", t.devName, "up")
	output, err := cmd.Output()
	println(output)
	return err
}

func (t TapLinux) Down() error {
	cmd := exec.Command("ip", "link", "set", "dev", t.devName, "down")
	output, err := cmd.Output()
	println(output)
	return err
}

func (t TapLinux) Read() (ethernet.Frame, error) {
	frame := ethernet.Frame{}

	buf := make([]byte, 2048)
	length, err := t.ReadWriteCloser.Read(buf)
	if err != nil {
		println(err.Error())
		return frame, err
	}
	err = frame.UnmarshalBinary(buf[0:length])
	return frame, err
}

func (t TapLinux) Write(frame ethernet.Frame) error {
	buf, err := frame.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = t.ReadWriteCloser.Write( buf)
	return err
}
