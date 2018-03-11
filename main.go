package main

import (
	"fmt"

	"gitlab.com/jinfagang/colorgo"
	"os"
	"flag"
	"time"
	"github.com/sirupsen/logrus"
	"net"
	"sort"
	"sync"
	"strings"
	"context"
	"./ips"
	"github.com/timest/gomanuf"
)

var VERSION = "0.1.1"

func PrintWelcome() {
	cg.PrintlnBlue(cg.BoldStart + "Peeper - man behind people. ")
	fmt.Println(`Star  : https://github.com/jinfagang/peeper
Author: Lucas Jin
Date  : 2018-3-7`)
	fmt.Println()
}


var log = logrus.New()
var ipNet *net.IPNet

var localHdAddress net.HardwareAddr
var iFace string

// 存放最终的数据，key[string] 存放的是IP地址
var data map[string]Info

// 计时器，在一段时间没有新的数据写入data中，退出程序，反之重置计时器
var t *time.Ticker
var do chan string

const (
	// 3秒的计时器
	START = "start"
	END = "end"
)

type Info struct {
	Mac      net.HardwareAddr
	Hostname string
	Vendor   string
}

func PrintData() {
	var keys ips.IPSlice
	for k := range data {
		keys = append(keys, ips.ParseIPString(k))
	}
	sort.Sort(keys)
	for _, k := range keys {
		d := data[k.String()]
		mac := ""
		if d.Mac != nil {
			mac = d.Mac.String()
		}
		fmt.Printf("%-15s %-17s %-30s %-10s\n", k.String(), mac, d.Hostname, d.Vendor)
	}
}

// 将抓到的数据集加入到data中，同时重置计时器
func pushData(ip string, mac net.HardwareAddr, hostname, vendor string) {
	// 停止计时器
	do <- START
	var mu sync.RWMutex
	mu.RLock()
	defer func() {
		// 重置计时器
		do <- END
		mu.RUnlock()
	}()
	if _, ok := data[ip]; !ok {
		data[ip] = Info{Mac: mac, Hostname: hostname, Vendor: vendor}
		return
	}
	info := data[ip]
	if len(hostname) > 0 && len(info.Hostname) == 0 {
		info.Hostname = hostname
	}
	if len(vendor) > 0 && len(info.Vendor) == 0 {
		info.Vendor = vendor
	}
	if mac != nil {
		info.Mac = mac
	}
	data[ip] = info
}

func setupNetInfo(f string) {
	var ifs []net.Interface
	var err error
	if f == "" {
		ifs, err = net.Interfaces()
	} else {
		var it *net.Interface
		it, err = net.InterfaceByName(f)
		if err == nil {
			ifs = append(ifs, *it)
		}
	}
	if err != nil {
		log.Fatal("无法获取本地网络信息:", err)
	}
	for _, it := range ifs {
		addr, _ := it.Addrs()
		for _, a := range addr {
			if ip, ok := a.(*net.IPNet); ok && !ip.IP.IsLoopback() {
				if ip.IP.To4() != nil {
					ipNet = ip
					localHdAddress = it.HardwareAddr
					iFace = it.Name
					goto END
				}
			}
		}
	}
END:
	if ipNet == nil || len(localHdAddress) == 0 {
		log.Fatal("无法获取本地网络信息")
	}
}

func localHost() {
	host, _ := os.Hostname()
	data[ipNet.IP.String()] = Info{
		Mac: localHdAddress,
		Hostname: strings.TrimSuffix(host, ".local"),
		Vendor: manuf.Search(localHdAddress.String())}
}

func sendARP() {
	// localIPs 是内网IP地址集合
	localIPs := ips.Table(ipNet)
	for _, ip := range localIPs {
		go sendArpPackage(ip)
	}
}

func main() {
	PrintWelcome()

	BasicInfo()

	// allow non root user to execute by compare with euid
	if os.Geteuid() != 0 {
		log.Fatal("peeper must run as root.")
	}
	flag.StringVar(&iFace, "I", "", "Network interface name")
	flag.Parse()

	// 初始化 data
	data = make(map[string]Info)
	do = make(chan string)

	// 初始化 网络信息
	setupNetInfo(iFace)

	ctx, cancel := context.WithCancel(context.Background())
	go listenARP(ctx)
	go listenMDNS(ctx)
	go listenNBNS(ctx)
	go sendARP()
	go localHost()

	t = time.NewTicker(4 * time.Second)
	for {
		select {
		case <-t.C:
			PrintData()
			cancel()
			goto END
		case d := <-do:
			switch d {
			case START:
				t.Stop()
			case END:
				// 接收到新数据，重置2秒的计数器
				t = time.NewTicker(2 * time.Second)
			}
		}
	}
END:

}


