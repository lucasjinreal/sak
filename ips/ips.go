package ips

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"math"

	log "github.com/sirupsen/logrus"
)


// this function get external ip
func GetExternalIp() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	//buf := new(bytes.Buffer)
	//buf.ReadFrom(resp.Body)
	//s := buf.String()
	return string(content)
}

// this function get local ip
// return a list of ips
func GetLocalIp() ([]string, error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var allIp []string
	for _, address := range addrs {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				allIp = append(allIp, ipNet.IP.String())
			}
		}
	}
	return allIp, nil
}

// get local ip net object, from this
// ip and mask can be accessed
type IPNetError struct {
	info string
	Err  error
}

func (ipError *IPNetError) Error() string {
	errorInfo := fmt.Sprintf("info: %s ,, original err info : %s ", ipError.info, ipError.Err.Error())
	return errorInfo
}

// this method will only return one IP
func GetLocalIpNet() (*net.IPNet, error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for _, address := range addrs {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet, nil
			}
		}
	}
	err = &IPNetError{
		info: "IPNet get error.",
		Err:  errors.New("test custom err"),
	}
	return nil, err
}

// get the IP table
func Table(ipNet *net.IPNet) []IP {
	ip := ipNet.IP.To4()
	log.Info("本机ip: ", ip)
	var min, max IP
	var data []IP
	for i := 0; i < 4; i++ {
		b := IP(ip[i] & ipNet.Mask[i])
		min += b << ((3 - uint(i)) * 8)
	}
	one, _ := ipNet.Mask.Size()
	max = min | IP(math.Pow(2, float64(32-one))-1)
	log.Infof("内网IP范围: %s ~ %s", min, max)
	// max 是广播地址，忽略
	// i & 0x000000ff  == 0 是尾段为0的IP，根据RFC的规定，忽略
	for i := min; i < max; i++ {
		if i&0x000000ff == 0 {
			continue
		}
		data = append(data, i)
	}
	return data
}
