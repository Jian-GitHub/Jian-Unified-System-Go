package config

import (
	"net"
	"os"
)

type SnowflakeConfig struct {
	NodeID int64 `json:",optional"` // 可选配置，未配置时自动生成
}

// SetupSnowflake 自动初始化雪花节点ID
func (c *Config) SetupSnowflake() error {
	if c.Snowflake.NodeID == 0 {
		// 基于Pod IP最后一段自动计算节点ID (0-1023)
		ip, err := getLastIPSegment()
		if err != nil {
			return err
		}
		c.Snowflake.NodeID = ip % 1024
	}
	return nil
}

// getLocalIPSegment 获取本机IP最后一段作为种子
func getLocalIPSegment() (int64, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return 0, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				seg := ipnet.IP.To4()[3]
				return int64(seg), nil
			}
		}
	}
	return 0, nil
}
func getLastIPSegment() (int64, error) {
	// 优先从环境变量获取（K8S环境）
	if podIP := os.Getenv("POD_IP"); podIP != "" {
		ip := net.ParseIP(podIP).To4()
		return int64(ip[3]), nil
	}
	// 非K8S环境回退到本地IP
	return getLocalIPSegment()
}
