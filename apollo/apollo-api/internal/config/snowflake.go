package config

import (
	"encoding/binary"
	"errors"
	"net"
	"os"
	"strings"
)

const maxSnowflakeNodeID int64 = 1023

// ResolveSnowflakeNodeID preserves the legacy Apollo behavior: a non-zero
// configured node is used directly; zero derives a node from POD_IP and then
// from the first non-loopback local IPv4 address.
func ResolveSnowflakeNodeID(configured int64) (int64, error) {
	if configured < 0 || configured > maxSnowflakeNodeID {
		return 0, errors.New("Snowflake.NodeID must be between 0 and 1023")
	}
	if configured != 0 {
		return configured, nil
	}
	if podIP := strings.TrimSpace(os.Getenv("POD_IP")); podIP != "" {
		return nodeIDFromIP(podIP)
	}
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return 0, err
	}
	for _, address := range addresses {
		var raw string
		switch value := address.(type) {
		case *net.IPNet:
			raw = value.IP.String()
		case *net.IPAddr:
			raw = value.IP.String()
		default:
			continue
		}
		ip := net.ParseIP(raw)
		if ip != nil && !ip.IsLoopback() && ip.To4() != nil {
			return nodeIDFromIPv4(ip.To4()), nil
		}
	}
	return 0, errors.New("cannot derive Snowflake.NodeID from a non-loopback IPv4 address")
}

func nodeIDFromIP(raw string) (int64, error) {
	ip := net.ParseIP(raw)
	if ip == nil || ip.To4() == nil {
		return 0, errors.New("POD_IP must be a valid IPv4 address")
	}
	return nodeIDFromIPv4(ip.To4()), nil
}

func nodeIDFromIPv4(ip net.IP) int64 {
	return int64(binary.BigEndian.Uint32(ip)) & maxSnowflakeNodeID
}
