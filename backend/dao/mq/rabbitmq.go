package mq

import (
	"backend/settings"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

var RabbitMQConn *amqp.Connection
var RabbitMQChannel *amqp.Channel

// InitRabbitMQ InitRabbitMQ 初始化 RabbitMQ 连接和通道
func InitRabbitMQ(config settings.RabbitMQConfig) error {
	var err error
	// 1. 建立连接
	RabbitMQConn, err = amqp.Dial(config.URL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
		return err
	}

	// 2. 打开通道
	RabbitMQChannel, err = RabbitMQConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
		return err
	}

	if RabbitMQConn == nil || RabbitMQChannel == nil {
		log.Fatalf("RabbitMQ connection or channel is not initialized")
		return fmt.Errorf("RabbitMQ not properly initialized")
	}

	// 3. 循环声明队列、交换机和绑定
	for _, queue := range config.Queues {
		// 声明交换机
		err = RabbitMQChannel.ExchangeDeclare(
			queue.Exchange, // 交换机名称
			"direct",       // 类型
			true,           // 持久化
			false,          // 自动删除
			false,          // 内部使用
			false,          // 阻塞
			nil,            // 额外参数
		)
		if err != nil {
			log.Fatalf("Failed to declare exchange %s: %v", queue.Exchange, err)
			return err
		}

		// 声明队列
		_, err = RabbitMQChannel.QueueDeclare(
			queue.Name, // 队列名称
			true,       // 持久化
			false,      // 自动删除
			false,      // 独占
			false,      // 阻塞
			nil,        // 额外参数
		)
		if err != nil {
			log.Fatalf("Failed to declare queue %s: %v", queue.Name, err)
			return err
		}

		// 绑定队列到交换机
		err = RabbitMQChannel.QueueBind(
			queue.Name,       // 队列名称
			queue.RoutingKey, // 路由键
			queue.Exchange,   // 交换机名称
			false,            // 阻塞
			nil,              // 额外参数
		)
		if err != nil {
			log.Fatalf("Failed to bind queue %s: %v", queue.Name, err)
			return err
		}

		log.Printf("Initialized queue: %s, exchange: %s, routing key: %s",
			queue.Name, queue.Exchange, queue.RoutingKey)
	}

	log.Println("RabbitMQ initialization completed")
	return nil
}

// PublishMessage 发送消息到队列
func PublishMessage(exchange, routingKey string, body []byte) error {
	err := RabbitMQChannel.Publish(
		exchange,
		routingKey,
		false, // Mandatory
		false, // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			//DeliveryMode: amqp.Persistent, // 可选: 设置消息持久化
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish a message: %s", err)
	}
	return err
}

// CloseRabbitMQ 关闭连接
func CloseRabbitMQ() {
	if RabbitMQChannel != nil {
		err := RabbitMQChannel.Close()
		if err != nil {
			return
		}
	}
	if RabbitMQConn != nil {
		err := RabbitMQConn.Close()
		if err != nil {
			return
		}
	}
}
