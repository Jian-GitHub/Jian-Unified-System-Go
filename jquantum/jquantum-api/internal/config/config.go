package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/jus-core/types/jwt"
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
	RabbitMQ  rabbitMQ.RabbitMQ
	SubSystem jwt.TokenConfig
}
