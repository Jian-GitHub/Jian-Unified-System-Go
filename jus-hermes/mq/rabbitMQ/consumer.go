package rabbitMQ

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/streadway/amqp"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Consumer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	rabbitMQ RabbitMQ
	redis    *redis.Redis
	ctx      context.Context
	done     chan bool
	wg       sync.WaitGroup
	f        func(body []byte)
}

func NewConsumer(r RabbitMQ, redisClient *redis.Redis, f func(body []byte)) *Consumer {
	conn, err := amqp.DialConfig(
		r.URL,
		amqp.Config{
			Heartbeat: time.Second * 10,
		})
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	// 设置 QoS - 一次只预取一条消息
	err = channel.Qos(
		1,     // 预取计数
		0,     // 预取大小
		false, // 全局
	)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	return &Consumer{
		conn:     conn,
		channel:  channel,
		rabbitMQ: r,
		redis:    redisClient,
		ctx:      context.Background(),
		done:     make(chan bool),
		f:        f,
	}
}

func (c *Consumer) StartConsuming() {
	// 声明队列
	_, err := c.channel.QueueDeclare(
		c.rabbitMQ.Queue,
		true,  // 持久化
		false, // 自动删除
		false, // 排他性
		false, // 不等待
		nil,   // 参数
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// 开始消费消息（手动确认）
	msgs, err := c.channel.Consume(
		c.rabbitMQ.Queue,
		"",    // 消费者标签
		false, // 自动确认
		false, // 排他性
		false, // 不等待
		false, // 无本地
		nil,   // 参数
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	log.Printf("开始分布式顺序处理队列: %s", c.rabbitMQ.Queue)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for {
			select {
			case <-c.done:
				log.Println("接收到停止信号，停止消费")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("消息通道已关闭")
					c.keepAlive()
					return
				}

				// 使用更健壮的分布式锁
				lockKey := "global:job_queue_lock"
				lockValue := c.generateLockValue() // 生成唯一的锁值

				// 尝试获取分布式锁
				acquired, err := c.tryAcquireLockWithRetry(lockKey, lockValue, 30*time.Second, 5)
				if err != nil {
					log.Printf("获取分布式锁失败: %v", err)
					// 拒绝消息并重新入队
					if err := msg.Nack(false, true); err != nil {
						log.Printf("拒绝消息失败: %v", err)
					}
					time.Sleep(1 * time.Second)
					continue
				}

				if acquired {
					log.Printf("进程获取到全局锁，开始处理消息: %s", string(msg.Body))

					// 处理消息
					c.processMessage(msg.Body)

					// 确认消息
					if err := msg.Ack(false); err != nil {
						log.Printf("确认消息失败: %v", err)
					} else {
						log.Printf("消息处理完成: %s", string(msg.Body))
					}

					// 释放锁
					if err := c.releaseLock(lockKey, lockValue); err != nil {
						log.Printf("释放锁失败: %v", err)
					}
				} else {
					// 未获取到锁，拒绝消息并重新入队
					log.Printf("进程未获取到全局锁，消息将重新入队: %s", string(msg.Body))
					if err := msg.Nack(false, true); err != nil {
						log.Printf("拒绝消息失败: %v", err)
					}

					// 等待一段时间再尝试
					time.Sleep(1 * time.Second)
				}
			}
		}
	}()
}

func (c *Consumer) keepAlive() {
	log.Printf("消息队列监听关闭, 重新开启监听: " + c.rabbitMQ.Queue)
	fmt.Println(c.rabbitMQ.URL)
	*c = *NewConsumer(c.rabbitMQ, c.redis, c.f)
	c.StartConsuming()
}

// 生成唯一的锁值
func (c *Consumer) generateLockValue() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机数生成失败，使用时间戳作为后备
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// 尝试获取锁，带有重试机制
func (c *Consumer) tryAcquireLockWithRetry(lockKey, lockValue string, ttl time.Duration, maxRetries int) (bool, error) {
	for i := 0; i < maxRetries; i++ {
		acquired, err := c.tryAcquireLock(lockKey, lockValue, ttl)
		if err != nil {
			return false, err
		}
		if acquired {
			return true, nil
		}

		// 等待一段时间再重试，使用指数退避算法
		sleepTime := time.Duration(math.Pow(2, float64(i))) * time.Millisecond * 100
		if sleepTime > time.Second {
			sleepTime = time.Second
		}
		time.Sleep(sleepTime)
	}
	return false, nil
}

// 尝试获取分布式锁
func (c *Consumer) tryAcquireLock(lockKey, lockValue string, ttl time.Duration) (bool, error) {
	// 使用 SET NX EX 命令获取锁
	result, err := c.redis.SetnxExCtx(c.ctx, lockKey, lockValue, int(ttl.Seconds()))
	if err != nil {
		return false, err
	}

	return result, nil
}

// 释放分布式锁
func (c *Consumer) releaseLock(lockKey, lockValue string) error {
	// 使用 Lua 脚本确保只有锁的持有者才能释放锁
	luaScript := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `

	// 执行 Lua 脚本
	eval, err := c.redis.EvalCtx(c.ctx, luaScript, []string{lockKey}, []string{lockValue})
	if err != nil {
		return err
	}

	if eval.(int64) == 0 {
		log.Printf("释放锁失败: 不是锁的持有者或锁已过期")
	} else {
		log.Printf("锁释放成功")
	}

	return nil
}

// 处理消息
func (c *Consumer) processMessage(body []byte) {
	// 模拟耗时操作
	log.Printf("开始处理消息: %s", string(body))
	//var msg jquantum.JobStructureMsg
	//
	//err := json.Unmarshal(body, &msg)
	//if err != nil {
	//	return
	//}
	c.f(body)
	//executor := joblogic.NewExecutor(msg.UserID, msg.JobID, c.config.JQuantum.BaseDir)
	//executor.Compile()
	//time.Sleep(10 * time.Second)
	log.Printf("处理完成: %s", string(body))
}

func (c *Consumer) Stop() {
	close(c.done)
	c.wg.Wait()
}

func (c *Consumer) Close() {
	c.Stop()
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
	log.Println("RabbitMQ连接已关闭")
}
