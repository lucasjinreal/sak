package main

import (
	"./common"
	"./ips"
	"gitlab.com/jinfagang/colorgo"
)

func BasicInfo() {
	allLocalIp, err := ips.GetLocalIp()
	common.CheckError(err)
	for _, v := range allLocalIp {
		cg.PrintBlue(cg.BoldStart + "[local IP]: " + cg.BoldEnd)
		cg.PrintlnGreen(v)
	}
	externalIp := ips.GetExternalIp()
	cg.PrintBlue(cg.BoldStart + "[external IP]: " + cg.BoldEnd)
	cg.PrintlnCyan(externalIp)

	// this line code not tested yet, still under construct
	// ipInfo, _ := ips.GetIpInfo(externalIp)
	// fmt.Println(ipInfo.Data)
	// cg.PrintlnCyan(ipInfo.Data)
}

// scan all the devices under current local net
func ScanLocalDevices() {
	ipNet, err := ips.GetLocalIpNet()
	common.CheckError(err)

	onlineIPs := ips.Table(ipNet)
	cg.PrintlnRed(onlineIPs)
}
