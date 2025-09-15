package apolloUtil

import (
	"fmt"
	"jian-unified-system/jus-core/util"
	"net"
	"net/http"
	"strings"
)

// GetLocate Get Locate str short code.
// eg. CN
func GetLocate(r *http.Request, f func(string) (*util.GeoResult, error)) string {
	ip := GetRealIP(r)
	fmt.Println("ip", ip)
	locate := "CN"
	info, err := f(ip)
	if err == nil && info != nil {
		if len(info.IsoCode) != 0 {
			locate = info.IsoCode
		}
	}
	return locate
}

// GetRealIP HTTP request -> IP Addr
func GetRealIP(r *http.Request) string {
	// Cloudflare
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		return cfip
	}
	// 尝试从 X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// 可能多个 IP 用逗号分隔，取第一个
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// 尝试从 X-Real-IP
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// 否则 fallback 到 RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
