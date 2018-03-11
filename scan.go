package main

import (

	"./common"
	"./ips"
)

func SendARP() {
	ipNet, err := ips.GetLocalIpNet()
	common.CheckError(err)
	ips := ips.Table(ipNet)
	for _, ip := range ips {
		go sendArpPackage(ip)
	}
}
