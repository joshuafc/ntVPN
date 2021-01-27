package tuntap

import (
	w "golang.org/x/sys/windows"
	r "golang.org/x/sys/windows/registry"
)

type NetInterface struct {
}

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
