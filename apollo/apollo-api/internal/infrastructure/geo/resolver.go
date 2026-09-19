package geo

import (
	"net"
	"net/http"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

// Resolver maps the originating HTTP client address to an ISO 3166-1 country
// code. It owns the MaxMind reader and must be closed with the service.
type Resolver struct {
	db *geoip2.Reader
}

const UnknownCountry = "UNKNOWN"

func Open(databasePath string) (*Resolver, error) {
	db, err := geoip2.Open(databasePath)
	if err != nil {
		return nil, err
	}
	return &Resolver{db: db}, nil
}

func (r *Resolver) Close() error { return r.db.Close() }

func (r *Resolver) Country(request *http.Request) string {
	ip := net.ParseIP(clientIP(request))
	if ip == nil {
		return UnknownCountry
	}
	record, err := r.db.Country(ip)
	if err != nil || record == nil || record.Country.IsoCode == "" {
		return UnknownCountry
	}
	return record.Country.IsoCode
}

func clientIP(request *http.Request) string {
	if value := strings.TrimSpace(request.Header.Get("CF-Connecting-IP")); value != "" {
		return value
	}
	if value := request.Header.Get("X-Forwarded-For"); value != "" {
		if first, _, ok := strings.Cut(value, ","); ok {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(value)
	}
	if value := strings.TrimSpace(request.Header.Get("X-Real-IP")); value != "" {
		return value
	}
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil {
		return host
	}
	return request.RemoteAddr
}
