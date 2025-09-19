package svc

import (
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jquantum/jquantum-api/internal/config"
	"jian-unified-system/jquantum/jquantum-api/internal/middleware"
	"jian-unified-system/jquantum/jquantum-rpc/jquantum"
	"jian-unified-system/jus-hermes/mq/rabbitMQ"
	"log"
	"time"
)

type ServiceContext struct {
	Config               config.Config
	KafkaWriter          *kafka.Writer
	Producer             *rabbitMQ.Producer
	JQuantumClient       jquantum.JQuantumClient
	ApolloSecurityClient apollo.SecurityClient
	TokenMiddleware      rest.Middleware
	jqClient             jquantum.JQuantumClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	var (
		loop           = true
		err            error
		jquantumClient zrpc.Client
		apolloClient   zrpc.Client
		writer         *kafka.Writer
		producer       *rabbitMQ.Producer
		jqClient       jquantum.JQuantumClient
		apolloSecurity apollo.SecurityClient
	)

	for loop {
		jquantumClient, err = zrpc.NewClient(c.JQuantumRpc)
		if err != nil {
			logx.Error("初始化JQuantumClient失败: " + err.Error())
			log.Println("30 秒后重试")
			time.Sleep(time.Second * 30)
			continue
		}
		apolloClient, err = zrpc.NewClient(c.ApolloRpc)
		if err != nil {
			logx.Error("初始化ApolloClient失败: " + err.Error())
			log.Println("30 秒后重试")
			time.Sleep(time.Second * 30)
			continue
		}

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
		writer = &kafka.Writer{
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

		producer, err = rabbitMQ.NewProducer(c.RabbitMQ)
		if err != nil {
			logx.Error("初始化RabbitMQ Producer失败: " + err.Error())
			log.Println("30 秒后重试")
			time.Sleep(time.Second * 30)
			continue
		}

		jqClient = jquantum.NewJQuantumClient(jquantumClient.Conn())
		apolloSecurity = apollo.NewSecurityClient(apolloClient.Conn())

		loop = false
	}
	return &ServiceContext{
		Config:               c,
		JQuantumClient:       jqClient,
		ApolloSecurityClient: apolloSecurity,
		KafkaWriter:          writer,
		Producer:             producer,
		TokenMiddleware:      middleware.NewTokenMiddleware(c.SubSystem.AccessSecret, apolloSecurity).Handle,
	}
}
