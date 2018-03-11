package main

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"./ips"
	"github.com/timest/gomanuf"
)

type ARP struct {
	layers.BaseLayer
	AddrType           layers.LinkType     // 硬件类型
	Protocol           layers.EthernetType // 协议类型
	HwAddressSize      uint8               // 硬件地址长度
	ProtoAddressSize   uint8               // 协议地址长度
	Operation          uint16              // 操作符(1代表request 2代表reply)
	SourceHwAddress    []byte              // 发送者硬件地址
	SourceProtoAddress []byte              // 发送者IP地址
	DstHwAddress       []byte              // 目标硬件地址（可以填写00:00:00:00:00:00)
	DstProtoAddress    []byte              // 目标IP地址
}

func listenARP(ctx context.Context) {
	handle, err := pcap.OpenLive(iFace, 1024, false, 10*time.Second)
	if err != nil {
		log.Fatal("pcap打开失败:", err)
	}
	defer handle.Close()
	handle.SetBPFFilter("arp")
	ps := gopacket.NewPacketSource(handle, handle.LinkType())
	for {
		select {
		case <-ctx.Done():
			return
		case p := <-ps.Packets():
			arp := p.Layer(layers.LayerTypeARP).(*layers.ARP)
			if arp.Operation == 2 {
				mac := net.HardwareAddr(arp.SourceHwAddress)
				m := manuf.Search(mac.String())

				pushData(ips.ParseIP(arp.SourceProtAddress).String(), mac, "", m)
				if strings.Contains(m, "Apple") {
					go sendMdns(ips.ParseIP(arp.SourceProtAddress), mac)
				} else {
					go sendNbns(ips.ParseIP(arp.SourceProtAddress), mac)
				}
			}
		}
	}
}

// send ARP bag to target IP
func sendArpPackage(ip ips.IP) {
	srcIp := net.ParseIP(ipNet.IP.String()).To4()
	dstIp := net.ParseIP(ip.String()).To4()
	if srcIp == nil || dstIp == nil {
		log.Fatal("IP parse error.")
	}
	// EthernetType 0x0806  ARP
	ether := &layers.Ethernet{
		SrcMAC:       localHdAddress,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}

	a := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     uint8(6),
		ProtAddressSize:   uint8(4),
		Operation:         uint16(1), // 0x0001 arp request 0x0002 arp response
		SourceHwAddress:   localHdAddress,
		SourceProtAddress: srcIp,
		DstHwAddress:      net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		DstProtAddress:    dstIp,
	}

	buffer := gopacket.NewSerializeBuffer()
	var opt gopacket.SerializeOptions
	gopacket.SerializeLayers(buffer, opt, ether, a)
	outgoingPacket := buffer.Bytes()

	handle, err := pcap.OpenLive(iFace, 2048, false, 30*time.Second)
	if err != nil {
		log.Fatal("pcap打开失败:", err)
	}
	defer handle.Close()

	err = handle.WritePacketData(outgoingPacket)
	if err != nil {
		log.Fatal("发送arp数据包失败..")
	}
}
