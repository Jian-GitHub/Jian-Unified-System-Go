package svc

import (
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/jquantum/jquantum-api/internal/config"
	"jian-unified-system/jquantum/jquantum-rpc/jquantum"
	"jian-unified-system/jus-hermes/mq/rabbitMQ"
	"time"
)

type ServiceContext struct {
	Config         config.Config
	KafkaWriter    *kafka.Writer
	Producer       *rabbitMQ.Producer
	JQuantumClient jquantum.JQuantumClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	client := zrpc.MustNewClient(c.JQuantumRpc)

	// SASL/PLAIN 机制
	mechanism := plain.Mechanism{
		Username: c.Kafka.Username,
		Password: c.Kafka.Password,
	}

	// Dialer 不使用 TLS，因为是 SASL_PLAINTEXT
	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		SASLMechanism: mechanism,
		// TLS: nil -> 明文 TCP
	}

	// 使用 WriterConfig 初始化 Writer（替代 NewWriter）
	writer := &kafka.Writer{
		Addr:     kafka.TCP(c.Kafka.Brokers...),
		Topic:    c.Kafka.Topic,
		Balancer: &kafka.LeastBytes{},
		Transport: &kafka.Transport{
			SASL:     dialer.SASLMechanism,
			TLS:      dialer.TLS,
			ClientID: dialer.ClientID,
		},
		AllowAutoTopicCreation: true,
	}

	producer := rabbitMQ.NewProducer(c.RabbitMQ)
	return &ServiceContext{
		Config:         c,
		JQuantumClient: jquantum.NewJQuantumClient(client.Conn()),
		KafkaWriter:    writer,
		Producer:       producer,
	}
}
