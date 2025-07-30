package util

import (
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"github.com/zeromicro/go-zero/core/jsonx"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
	"net"
)

type GeoService struct {
	DB *geoip2.Reader
}

type GeoResult struct {
	Country string
	City    string
	Region  string
	IsoCode string
}

func NewGeoService() (*GeoService, error) {
	db, err := geoip2.Open("../../jus-core/data/GeoLite2-City.mmdb")
	if err != nil {
		return nil, err
	}
	return &GeoService{DB: db}, nil
}

func (g *GeoService) Lookup(ipStr string) (*GeoResult, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, nil
	}

	record, err := g.DB.City(ip)
	if err != nil {
		return nil, err
	}
	fmt.Println("record")
	fmt.Println(jsonx.MarshalToString(record))

	country := record.Country.Names["en"]
	city := record.City.Names["en"]
	isoCode := record.Country.IsoCode
	region := ""
	if len(record.Subdivisions) > 0 {
		region = record.Subdivisions[0].Names["en"]
	}

	return &GeoResult{
		Country: country,
		City:    city,
		Region:  region,
		IsoCode: isoCode,
	}, nil
}

func (g *GeoService) GetLocalizedRegionName(regionCode string, langCode string) (string, error) {
	// 解析目标语言标签
	langTag, err := language.Parse(langCode)
	if err != nil {
		return "", fmt.Errorf("无法解析语言代码: %v", err)
	}

	// 解析国家/地区标签
	regionTag, err := language.ParseRegion(regionCode)
	if err != nil {
		return "", fmt.Errorf("无法解析地区代码: %v", err)
	}

	// 获取 display.Regions() 提供器
	provider := display.Regions(langTag)

	// 获取本地化名称
	name := provider.Name(regionTag)

	return name, nil
}
