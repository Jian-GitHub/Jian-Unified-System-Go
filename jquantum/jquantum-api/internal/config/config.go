package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/jus-hermes/mq/rabbitMQ"
)

type Config struct {
	rest.RestConf
	JQuantumRpc zrpc.RpcClientConf
	Kafka       struct {
		Brokers  []string
		Topic    string
		Username string
		Password string
		SASL     string // 认证机制
		TLS      bool   // 是否启用 TLS
	}
	RabbitMQ rabbitMQ.RabbitMQ
	Auth     struct { // JWT 认证需要的密钥和过期时间配置
		AccessSecret string
		AccessExpire int64
	}
}
