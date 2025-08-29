package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/jus-hermes/mq/rabbitMQ"
)

type Config struct {
	zrpc.RpcServerConf
	Kafka struct {
		Brokers []string
		Topic   string
		//GroupID  string
		Username string
		Password string
		SASL     string // 认证机制
		TLS      bool   // 是否启用 TLS
	}
	RabbitMQ  rabbitMQ.RabbitMQ
	RedisConf redis.RedisConf
	JQuantum  struct {
		BaseDir string
	}
}
