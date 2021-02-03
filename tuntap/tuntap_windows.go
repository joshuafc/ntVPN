package tuntap

import (
	w "golang.org/x/sys/windows"
	r "golang.org/x/sys/windows/registry"
	"unsafe"
)

type NetInterface struct {
}

func CTL_CODE(DeviceType uint32, Function uint32, Method uint32, Access uint32) uint32 {
	return (((DeviceType) << 16) | ((Access) << 14) | ((Function) << 2) | (Method))
}

func TAP_CONTROL_CODE(request uint32,method uint32) uint32{
	var FILE_DEVICE_UNKNOWN, FILE_ANY_ACCESS uint32
	FILE_DEVICE_UNKNOWN = 0x00000022
	FILE_ANY_ACCESS = 0
	return CTL_CODE (FILE_DEVICE_UNKNOWN, request, method, FILE_ANY_ACCESS)
}
var METHOD_BUFFERED uint32 = 0
var TAP_IOCTL_GET_MAC               = TAP_CONTROL_CODE (1, METHOD_BUFFERED)
var TAP_IOCTL_GET_VERSION           = TAP_CONTROL_CODE (2, METHOD_BUFFERED)
var TAP_IOCTL_GET_MTU               = TAP_CONTROL_CODE (3, METHOD_BUFFERED)
var TAP_IOCTL_GET_INFO              = TAP_CONTROL_CODE (4, METHOD_BUFFERED)
var TAP_IOCTL_CONFIG_POINT_TO_POINT = TAP_CONTROL_CODE (5, METHOD_BUFFERED)
var TAP_IOCTL_SET_MEDIA_STATUS      = TAP_CONTROL_CODE (6, METHOD_BUFFERED)
var TAP_IOCTL_CONFIG_DHCP_MASQ      = TAP_CONTROL_CODE (7, METHOD_BUFFERED)
var TAP_IOCTL_GET_LOG_LINE          = TAP_CONTROL_CODE (8, METHOD_BUFFERED)
var TAP_IOCTL_CONFIG_DHCP_SET_OPT   = TAP_CONTROL_CODE (9, METHOD_BUFFERED)

func (*NetInterface) Set() bool {
	NETWORK_CONNECTIONS_KEY := "SYSTEM\\CurrentControlSet\\Control\\Network\\{4D36E972-E325-11CE-BFC1-08002BE10318}"
	key, ok := r.OpenKey(r.LOCAL_MACHINE, NETWORK_CONNECTIONS_KEY, r.READ)
	if ok != nil {
		println(ok.Error())
		return false
	}
	keys, ok := key.ReadSubKeyNames(0)
	_ = key.Close()
	handler := w.Handle(0)
	for _, subkey := range keys {
		keypath := NETWORK_CONNECTIONS_KEY + "\\" + subkey + "\\Connection"
		key, ok = r.OpenKey(r.LOCAL_MACHINE, keypath, r.READ)
		if ok != nil {
			println(ok.Error())
			continue
		}
		name, _, _ := key.GetStringValue("Name")
		tapname := "\\\\.\\Global\\" + subkey + ".tap"
		filepath := w.StringToUTF16Ptr(tapname)
		handler, ok = w.CreateFile(filepath, w.GENERIC_WRITE|w.GENERIC_READ, 0, nil, w.OPEN_EXISTING, w.FILE_ATTRIBUTE_SYSTEM|w.FILE_FLAG_OVERLAPPED, 0)
		if ok == nil {
			var length uint32 = 0
			var status uint32 = 1
			err := w.DeviceIoControl(handler, TAP_IOCTL_SET_MEDIA_STATUS,
				(*byte)(unsafe.Pointer(&status)), 8,
				(*byte)(unsafe.Pointer(&status)),8, &length, nil)
			if err != nil {
				println(err.Error())
			}
			println(handler)
		} else {
			println(ok.Error())
		}
		println(name)
		_ = key.Close()
	}

	println(handler)
	_ = w.Close(handler)
	println(keys)
	println(key)
	println(ok)
	return false
}
