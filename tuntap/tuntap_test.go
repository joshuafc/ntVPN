package tuntap

import (
	"runtime"
	"testing"
)

func TestNetInterface_Set(t *testing.T) {
	var deviceName string
	if runtime.GOOS == "linux" {
		deviceName = "testedge"
	}else{
		deviceName = ""
	}
	net, err := CreateTapDevice(deviceName)
	if err != nil {
		panic(err.Error())
	}

	err = net.SetHardwareAddr("AA:BB:CC:88:66:55")
	macaddr, _ := net.GetHardwareAddr()
	println(macaddr)
	err = net.Up()
	if err != nil {
		panic(err.Error())
	}
	err = net.SetIpAddr("192.168.150.7", "255.255.255.0")
	ipaddr, mask, _ := net.GetIpAddr()
	println(ipaddr)
	println(mask)
	for {
		frame, err := net.Read()
		if err != nil {
			panic(err.Error())
		}
		tmp := frame.Destination
		frame.Destination = frame.Source
		frame.Source = tmp
		err = net.Write(frame)
		if err != nil {
			panic(err.Error())
		}
	}
}
