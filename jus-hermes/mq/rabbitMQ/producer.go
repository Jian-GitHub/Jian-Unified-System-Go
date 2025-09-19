package rabbitMQ

import (
	"github.com/zeromicro/go-zero/core/errorx"

	//"github.com/streadway/amqp"
	"github.com/rabbitmq/amqp091-go"
	"log"
)

type Producer struct {
	conn     *amqp091.Connection
	channel  *amqp091.Channel
	rabbitMQ RabbitMQ
}

func NewProducer(r RabbitMQ) (*Producer, error) {
	conn, err := amqp091.Dial(r.URL)
	if err != nil {
		return nil, errorx.Wrap(err, "Failed to connect to RabbitMQ")
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, errorx.Wrap(err, "Failed to open a channel: %v")
	}

	// 声明交换机
	err = channel.ExchangeDeclare(
		r.Exchange, // name
		"direct",   // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return nil, errorx.Wrap(err, "Failed to declare an exchange: %v")
	}

	// 声明队列
	_, err = channel.QueueDeclare(
		r.Queue, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return nil, errorx.Wrap(err, "Failed to declare a queue: %v")
	}

	// 绑定队列到交换机
	err = channel.QueueBind(
		r.Queue,      // queue name
		r.RoutingKey, // routing key
		r.Exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind a queue: %v", err)
	}

	return &Producer{
		conn:     conn,
		channel:  channel,
		rabbitMQ: r,
	}, nil
}

func (p *Producer) Publish(message []byte) error {
	return p.channel.Publish(
		p.rabbitMQ.Exchange,   // exchange
		p.rabbitMQ.RoutingKey, // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        message,
		})
}

func (p *Producer) PublishWithHeaders(message []byte, headers *amqp091.Table) error {
	return p.channel.Publish(
		p.rabbitMQ.Exchange,   // exchange
		p.rabbitMQ.RoutingKey, // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        message,
			Headers:     *headers,
			//MessageId:     "msg-001",
			//CorrelationId: "corr-001",
			//DeliveryMode:  2,               // 持久化
			//Timestamp:     time.Now(),
			AppId: "Jian Unified System - Hermes - Go",
		})
}

func (p *Producer) Close() {
	_ = p.channel.Close()
	_ = p.conn.Close()
}
