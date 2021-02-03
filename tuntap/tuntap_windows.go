package tuntap

import (
	"encoding/hex"
	"errors"
	"github.com/mdlayher/ethernet"
	w "golang.org/x/sys/windows"
	r "golang.org/x/sys/windows/registry"
	"os/exec"
	"regexp"
	"time"
	"unsafe"
)

func CTL_CODE(DeviceType uint32, Function uint32, Method uint32, Access uint32) uint32 {
	return (((DeviceType) << 16) | ((Access) << 14) | ((Function) << 2) | (Method))
}

func TAP_CONTROL_CODE(request uint32, method uint32) uint32 {
	var FILE_DEVICE_UNKNOWN, FILE_ANY_ACCESS uint32
	FILE_DEVICE_UNKNOWN = 0x00000022
	FILE_ANY_ACCESS = 0
	return CTL_CODE(FILE_DEVICE_UNKNOWN, request, method, FILE_ANY_ACCESS)
}

var METHOD_BUFFERED uint32 = 0
var TAP_IOCTL_GET_MAC = TAP_CONTROL_CODE(1, METHOD_BUFFERED)
var TAP_IOCTL_GET_VERSION = TAP_CONTROL_CODE(2, METHOD_BUFFERED)
var TAP_IOCTL_GET_MTU = TAP_CONTROL_CODE(3, METHOD_BUFFERED)
var TAP_IOCTL_GET_INFO = TAP_CONTROL_CODE(4, METHOD_BUFFERED)
var TAP_IOCTL_CONFIG_POINT_TO_POINT = TAP_CONTROL_CODE(5, METHOD_BUFFERED)
var TAP_IOCTL_SET_MEDIA_STATUS = TAP_CONTROL_CODE(6, METHOD_BUFFERED)
var TAP_IOCTL_CONFIG_DHCP_MASQ = TAP_CONTROL_CODE(7, METHOD_BUFFERED)
var TAP_IOCTL_GET_LOG_LINE = TAP_CONTROL_CODE(8, METHOD_BUFFERED)
var TAP_IOCTL_CONFIG_DHCP_SET_OPT = TAP_CONTROL_CODE(9, METHOD_BUFFERED)

var NETWORK_CONNECTIONS_KEY = "SYSTEM\\CurrentControlSet\\Control\\Network\\{4D36E972-E325-11CE-BFC1-08002BE10318}"
var ADAPTER_INFO_KEY = "SYSTEM\\CurrentControlSet\\Control\\Class\\{4D36E972-E325-11CE-BFC1-08002BE10318}"

type TapWindows struct {
	handler        w.Handle
	devName        string
	deviceRegistry string
}

func lookup_adapter_reg_path(deviceRegistry string) string {
	key, ok := r.OpenKey(r.LOCAL_MACHINE, ADAPTER_INFO_KEY, r.READ)
	if ok != nil {
		println(ok.Error())
		return ""
	}
	keys, ok := key.ReadSubKeyNames(0)
	_ = key.Close()
	for _, subkey := range keys {
		keypath := ADAPTER_INFO_KEY + "\\" + subkey
		key, ok = r.OpenKey(r.LOCAL_MACHINE, keypath, r.READ)
		if ok != nil {
			println(ok.Error())
			continue
		}
		devRegID, _, _ := key.GetStringValue("NetCfgInstanceId")
		if devRegID == deviceRegistry {
			return keypath
		}
		_ = key.Close()
	}
	return ""
}

func CreateTapDevice(deviceName string) (TapDevice, error) {
	key, ok := r.OpenKey(r.LOCAL_MACHINE, NETWORK_CONNECTIONS_KEY, r.READ)
	if ok != nil {
		println(ok.Error())
		return nil, errors.New("Cannot Open Registry!")
	}
	keys, ok := key.ReadSubKeyNames(0)
	_ = key.Close()
	for _, subkey := range keys {
		tapWindows := TapWindows{}
		tapWindows.deviceRegistry = subkey
		keypath := NETWORK_CONNECTIONS_KEY + "\\" + subkey + "\\Connection"
		key, ok = r.OpenKey(r.LOCAL_MACHINE, keypath, r.READ)
		if ok != nil {
			println(ok.Error())
			continue
		}
		tapWindows.devName, _, _ = key.GetStringValue("Name")
		if len(deviceName) > 0 && tapWindows.devName != deviceName {
			continue
		}
		tapname := "\\\\.\\Global\\" + subkey + ".tap"
		filepath := w.StringToUTF16Ptr(tapname)
		tapWindows.handler, ok = w.CreateFile(filepath, w.GENERIC_WRITE|w.GENERIC_READ, 0, nil, w.OPEN_EXISTING, w.FILE_ATTRIBUTE_SYSTEM, 0)
		if ok == nil {
			return &tapWindows, nil
		} else {
			println(ok.Error())
		}
		_ = key.Close()
	}
	return &TapWindows{}, nil
}

func DestroyTapDevice(tapDevice TapDevice) error {
	_ = w.Close(tapDevice.(*TapWindows).handler)
	return nil
}

func (t *TapWindows) SetHardwareAddr(addr string) error {
	addrarr := ([]byte)(addr)
	newaddr := make([]byte, 0, 12)
	for _, val := range addrarr {
		if val != ':' {
			newaddr = append(newaddr, val)
		}
	}
	regpath := lookup_adapter_reg_path(t.deviceRegistry)

	cmd := exec.Command("reg", "add", "HKEY_LOCAL_MACHINE\\"+regpath, "/v", "MAC", "/d", string(newaddr), "/f")
	output, err := cmd.Output()
	println(output)

	cmd = exec.Command("reg", "add", "HKEY_LOCAL_MACHINE\\"+regpath, "/v", "NetworkAddress", "/d", string(newaddr), "/f")
	output, err = cmd.Output()
	println(output)

	_ = w.Close(t.handler)

	cmd = exec.Command("netsh", "interface", "set", "interface", t.devName, "disabled")
	output, err = cmd.Output()
	println(output)

	cmd = exec.Command("netsh", "interface", "set", "interface", t.devName, "enabled")
	output, err = cmd.Output()
	println(output)

	tapname := "\\\\.\\Global\\" + t.deviceRegistry + ".tap"
	filepath := w.StringToUTF16Ptr(tapname)
	var ok error
	t.handler, ok = w.CreateFile(filepath, w.GENERIC_WRITE|w.GENERIC_READ, 0, nil, w.OPEN_EXISTING, w.FILE_ATTRIBUTE_SYSTEM, 0)
	if ok != nil {
		panic(ok.Error())
	}
	return err
}

func (t *TapWindows) GetHardwareAddr() (string, error) {
	var reslen uint32 = 0
	macaddr := make([]byte, 6)
	err := w.DeviceIoControl(t.handler, TAP_IOCTL_GET_MAC,
		&macaddr[0], 6,
		&macaddr[0], 6, &reslen, nil)
	macres := ""
	if err == nil {
		macres = hex.EncodeToString(macaddr[0:1])
		macres = macres + ":" + hex.EncodeToString(macaddr[1:2])
		macres = macres + ":" + hex.EncodeToString(macaddr[2:3])
		macres = macres + ":" + hex.EncodeToString(macaddr[3:4])
		macres = macres + ":" + hex.EncodeToString(macaddr[4:5])
		macres = macres + ":" + hex.EncodeToString(macaddr[5:6])
	}
	return macres, err
}

func (t *TapWindows) SetIpAddr(addr, mask string) error {
	cmd := exec.Command("netsh", "interface", "ip", "set", "address", t.devName, "static", addr, mask)
	output, err := cmd.Output()
	println(output)
	time.Sleep(time.Second * 3)
	return err
}

func (t *TapWindows) GetIpAddr() (string, string, error) {
	cmd := exec.Command("netsh", "interface", "ip", "show", "address", t.devName)
	output, err := cmd.Output()
	println(output)
	ipregex, err1 := regexp.Compile("[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+\\s")
	if err1 != nil {
		println(err1.Error())
		return "", "", err1
	}
	ipres := ipregex.FindString(string(output))
	maskregex, err2 := regexp.Compile("[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+\\)")
	if err2 != nil {
		println(err2.Error())
		return "", "", err2
	}
	maskres := maskregex.FindString(string(output))
	if len(ipres) == 0 || len(maskres) == 0 {
		return "", "", errors.New("Not Found IP")
	}
	return ipres[0 : len(ipres)-1], maskres[0 : len(maskres)-1], err
}

func (t *TapWindows) Up() error {
	var length uint32 = 0
	var status uint64 = 1
	return w.DeviceIoControl(t.handler, TAP_IOCTL_SET_MEDIA_STATUS,
		(*byte)(unsafe.Pointer(&status)), 8,
		(*byte)(unsafe.Pointer(&status)), 8, &length, nil)
}

func (t *TapWindows) Down() error {
	var length uint32 = 0
	var status uint64 = 0
	return w.DeviceIoControl(t.handler, TAP_IOCTL_SET_MEDIA_STATUS,
		(*byte)(unsafe.Pointer(&status)), 8,
		(*byte)(unsafe.Pointer(&status)), 8, &length, nil)
}

func (t *TapWindows) Read() (ethernet.Frame, error) {
	frame := ethernet.Frame{}

	buf := make([]byte, 2048)
	len, err := w.Read(t.handler, buf)
	if err != nil {
		println(err.Error())
		return frame, err
	}
	err = frame.UnmarshalBinary(buf[0:len])
	return frame, err
}

func (t *TapWindows) Write(frame ethernet.Frame) error {
	buf, err := frame.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = w.Write(t.handler, buf)
	return err
}
