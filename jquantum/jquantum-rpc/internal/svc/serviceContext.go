package svc

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"jian-unified-system/jquantum/jquantum-rpc/internal/config"
	jobService "jian-unified-system/jquantum/jquantum-rpc/internal/service/job"

	//"jian-unified-system/jquantum/jquantum-rpc/internal/mq"

	"jian-unified-system/jus-hermes/mq/rabbitMQ"
	"log"
	"time"
)

type ServiceContext struct {
	Config      config.Config
	KafkaReader *kafka.Reader
	Consumer    *rabbitMQ.Consumer
	Redis       *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {

	// 创建认证机制
	mechanism := plain.Mechanism{
		Username: c.Kafka.Username,
		Password: c.Kafka.Password,
	}

	// 正确的方式：使用 Dialer
	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
	}

	// 创建 Reader
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.Kafka.Brokers,
		//GroupID:        c.Kafka.GroupID,
		Topic:    c.Kafka.Topic,
		Dialer:   dialer,
		MinBytes: 0,
		MaxBytes: 10e6,
		//MaxWait:        2 * time.Second,
		//CommitInterval: time.Second,
		StartOffset: kafka.LastOffset,
		//Logger:         kafka.LoggerFunc(log.Printf),
		ErrorLogger: kafka.LoggerFunc(log.Printf),
	})

	redisClient, err := redis.NewRedis(c.RedisConf)
	consumer := rabbitMQ.NewConsumer(c.RabbitMQ, redisClient, jobService.NewExecutor(c.JQuantum.BaseDir).Process)
	// 启动消费者
	consumer.StartConsuming()

	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:      c,
		KafkaReader: r,
		Redis:       redisClient,
		Consumer:    consumer,
	}
}

func (sc *ServiceContext) StartKafkaConsumer() {
	defer sc.KafkaReader.Close()

	ctx := context.Background()

	log.Println("Kafka 消费者启动...")

	for {
		m, err := sc.KafkaReader.ReadMessage(ctx)
		if err != nil {
			log.Printf("读取 Kafka 消息失败: %v\n", err)
			continue
		}

		// 在这里处理 Kafka 消息
		fmt.Printf("[Kafka] 收到消息: %s = %s (offset=%d)\n", string(m.Key), string(m.Value), m.Offset)

		// 消息 ->channel，交给 worker pool 异步处理
	}
}
