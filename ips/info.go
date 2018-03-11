package ips

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"gitlab.com/jinfagang/colorgo"
)

type InfoData struct {
	Country   string `json:"country"`
	CountryId string `json:"country_id"`
	Area      string `json:"area"`
	AreaId    string `json:"area_id"`
	Region    string `json:"region"`
	RegionId  string `json:"region_id"`
	City      string `json:"city"`
	CityId    string `json:"city_id"`
	Isp       string `json:"isp"`
}

type IPInfo struct {
	Code int `json:"code"`
	Data InfoData  `json:"data"`
}

func GetIpInfo(ip string) (*IPInfo, error) {
	url := "https://ip.taobao.com/service/getIpInfo.php?ip="
	url += ip

	resp, err := http.Get(url)
	fmt.Println(resp)
	if err != nil {
		cg.PrintlnRed("[Error]: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		cg.PrintlnRed("[Error]: io error" + err.Error())
		return nil, err
	}
	var result IPInfo
	if err := json.Unmarshal(out, &result); err != nil {
		cg.PrintlnRed("[Error]: " + err.Error())
		return nil, err
	}
	return &result, nil
}
