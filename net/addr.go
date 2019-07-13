package net

import (
	"net"
)

// get all local addrs
func GetLocalAddr() ([]string, error) {
	var ret []string

	addrs, err := net.InterfaceAddrs()

	if err == nil {
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil || ipnet.IP.To16() != nil {
					ret = append(ret, ipnet.IP.String())
				}
			}
		}
	}

	return ret, err
}
