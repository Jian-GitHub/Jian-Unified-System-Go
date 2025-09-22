package svc

import (
	"errors"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jquantum/jquantum-rpc/internal/config"
	"jian-unified-system/jquantum/jquantum-rpc/internal/service/deploy"
	"jian-unified-system/jus-core/data/mysql/jquantum"
	"jian-unified-system/jus-core/util"
	"jian-unified-system/jus-hermes/email/service"
	"os/exec"

	//"jian-unified-system/jquantum/jquantum-rpc/internal/mq"

	"jian-unified-system/jus-hermes/mq/rabbitMQ"
)

type ServiceContext struct {
	Config        config.Config
	ApolloAccount apollo.AccountClient
	//ApolloSecurityAccount apollo.SecurityClient
	//KafkaReader *kafka.Reader
	Producer *rabbitMQ.Producer
	//Consumer    *rabbitMQ.Consumer
	Redis                   *redis.Redis
	JobModel                jquantum.JobModel
	EmailService            service.EmailService
	KubernetesDeployService deploy.K8sDeployService
}

func NewServiceContext(c config.Config) *ServiceContext {
	var (
		sqlConn sqlx.SqlConn
		client  zrpc.Client
		//kafkaReader             *kafka.Reader
		redisClient             *redis.Redis
		producer                *rabbitMQ.Producer
		emailService            service.EmailService
		kubernetesDeployService deploy.K8sDeployService
		cc                      grpc.ClientConnInterface
	)

	for {
		if err := util.RetryWithBackoff("start sshd", func() error {
			return exec.Command("/usr/sbin/sshd").Start()
		}); err != nil {
			continue
		}

		if err := util.RetryWithBackoff("ApolloRpc", func() error {
			var err error
			client, err = zrpc.NewClient(c.ApolloRpc)
			return err
		}); err != nil {
			continue
		}

		if err := util.RetryWithBackoff("ApolloRpc.client.conn", func() error {
			if cc = client.Conn(); cc == nil {
				return errors.New("client.Conn() is null")
			}
			return nil
		}); err != nil {
			continue
		}

		if err := util.RetryWithBackoff("Redis", func() error {
			var err error
			redisClient, err = redis.NewRedis(c.RedisConf)
			return err
		}); err != nil {
			continue
		}

		if err := util.RetryWithBackoff("RabbitMQ", func() error {
			var err error
			producer, err = rabbitMQ.NewProducer(c.RabbitMQ)
			return err
		}); err != nil {
			continue
		}

		if err := util.RetryWithBackoff("RabbitMQ", func() error {
			var err error
			kubernetesDeployService, err = deploy.NewK8sDeployService(c.JQuantum.Namespace)
			return err
		}); err != nil {
			continue
		}

		break
	}
	// MySQL
	sqlConn = sqlx.NewMysql(c.DB.DataSource)
	// Email
	emailService = service.DefaultEmailService()

	// Kafka
	// 创建认证机制
	//mechanism := plain.Mechanism{
	//	Username: c.Kafka.Username,
	//	Password: c.Kafka.Password,
	//}

	// 正确的方式：使用 Dialer
	//dialer := &kafka.Dialer{
	//	Timeout:       10 * time.Second,
	//	DualStack:     true,
	//	SASLMechanism: mechanism,
	//}

	// 创建 Reader
	//kafkaReader = kafka.NewReader(kafka.ReaderConfig{
	//	Brokers: c.Kafka.Brokers,
	//	//GroupID:        c.Kafka.GroupID,
	//	Topic:    c.Kafka.Topic,
	//	Dialer:   dialer,
	//	MinBytes: 0,
	//	MaxBytes: 10e6,
	//	//MaxWait:        2 * time.Second,
	//	//CommitInterval: time.Second,
	//	StartOffset: kafka.LastOffset,
	//	//Logger:         kafka.LoggerFunc(log.Printf),
	//	ErrorLogger: kafka.LoggerFunc(log.Printf),
	//})

	return &ServiceContext{
		Config:        c,
		ApolloAccount: apollo.NewAccountClient(cc),
		//ApolloSecurityAccount: apollo.NewSecurityClient(client.Conn()),
		//KafkaReader: kafkaReader,
		Redis:    redisClient,
		JobModel: jquantum.NewJobModel(sqlConn, c.Cache),
		Producer: producer,
		//Consumer:    consumer,
		EmailService:            emailService,
		KubernetesDeployService: kubernetesDeployService,
	}
}

//func (sc *ServiceContext) StartKafkaConsumer() {
//	defer func(KafkaReader *kafka.Reader) {
//		_ = KafkaReader.Close()
//	}(sc.KafkaReader)
//
//	ctx := context.Background()
//
//	log.Println("Kafka 消费者启动...")
//
//	for {
//		m, err := sc.KafkaReader.ReadMessage(ctx)
//		if err != nil {
//			log.Printf("读取 Kafka 消息失败: %v\n", err)
//			continue
//		}
//
//		// 在这里处理 Kafka 消息
//		fmt.Printf("[Kafka] 收到消息: %s = %s (offset=%d)\n", string(m.Key), string(m.Value), m.Offset)
//
//		// 消息 ->channel，交给 worker pool 异步处理
//	}
//}
