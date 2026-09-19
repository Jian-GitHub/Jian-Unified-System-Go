package geo

import (
	"net/http/httptest"
	"testing"
)

func TestClientIPPrecedence(t *testing.T) {
	request := httptest.NewRequest("POST", "/registration", nil)
	request.RemoteAddr = "192.0.2.4:1234"
	request.Header.Set("X-Real-IP", "192.0.2.3")
	request.Header.Set("X-Forwarded-For", "192.0.2.2, 10.0.0.1")
	request.Header.Set("CF-Connecting-IP", "192.0.2.1")
	if got := clientIP(request); got != "192.0.2.1" {
		t.Fatalf("CF address lost: %q", got)
	}
	request.Header.Del("CF-Connecting-IP")
	if got := clientIP(request); got != "192.0.2.2" {
		t.Fatalf("forwarded address lost: %q", got)
	}
	request.Header.Del("X-Forwarded-For")
	if got := clientIP(request); got != "192.0.2.3" {
		t.Fatalf("real IP address lost: %q", got)
	}
	request.Header.Del("X-Real-IP")
	if got := clientIP(request); got != "192.0.2.4" {
		t.Fatalf("remote address lost: %q", got)
	}
}

func TestCountryAndFallback(t *testing.T) {
	resolver, err := Open("../../../../../jus-core/data/GeoLite2-City.mmdb")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resolver.Close() })

	request := httptest.NewRequest("POST", "/registration", nil)
	request.Header.Set("CF-Connecting-IP", "8.8.8.8")
	if got := resolver.Country(request); got != "US" {
		t.Fatalf("GeoIP country mismatch: %q", got)
	}
	request.Header.Set("CF-Connecting-IP", "invalid")
	if got := resolver.Country(request); got != UnknownCountry {
		t.Fatalf("fallback mismatch: %q", got)
	}
	request.Header.Set("CF-Connecting-IP", "192.168.1.10")
	if got := resolver.Country(request); got != UnknownCountry {
		t.Fatalf("private address mismatch: %q", got)
	}
}
