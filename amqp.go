package main

import (
	"fmt"
	// "time"

	"github.com/streadway/amqp"
)

// 定义全局变量,指针类型
var mqConn *amqp.Connection
var mqChan *amqp.Channel

//定义Rabbitmq连接
type MqConnect struct {
	hostname string
	port     uint16
	username string
	password string
}

// 定义RabbitMQ对象
type RabbitMQ struct {
	hostname     string
	port         uint16
	username     string
	password     string
	connection   *amqp.Connection
	channel      *amqp.Channel
	queueName    string // 队列名称
	routingKey   string // key名称
	exchangeName string // 交换机名称
	exchangeType string // 交换机类型
}

// 定义队列交换机对象
type QueueExchange struct {
	QuName string // 队列名称
	RtKey  string // key值
	ExName string // 交换机名称
	ExType string // 交换机类型
}

// 链接rabbitMQ
func (r *RabbitMQ) mqConnect() {
	var err error
	RabbitUrl := fmt.Sprintf("amqp://%s:%s@%s:%d/", r.username, r.password, r.hostname, r.port)
	mqConn, err = amqp.Dial(RabbitUrl)
	r.connection = mqConn // 赋值给RabbitMQ对象
	if err != nil {
		fmt.Printf("MQ打开链接失败:%s \n", err)
	}
	mqChan, err = mqConn.Channel()
	r.channel = mqChan // 赋值给RabbitMQ对象
	if err != nil {
		fmt.Printf("MQ打开管道失败:%s \n", err)
	}
}

// 关闭RabbitMQ连接
func (r *RabbitMQ) mqClose() {
	// 先关闭管道,再关闭链接
	err := r.channel.Close()
	if err != nil {
		fmt.Printf("MQ管道关闭失败:%s \n", err)
	}
	err = r.connection.Close()
	if err != nil {
		fmt.Printf("MQ链接关闭失败:%s \n", err)
	}
}

//创建新的连接
func DefaultConnect() *RabbitMQ {
	return &RabbitMQ{
		hostname:     "iw16.com",
		port:         5672,
		username:     "admin",
		password:     "wang*qing",
		routingKey:   "test",
		exchangeName: "test",
		exchangeType: "direct",
	}
}

//创建新的连接
func DefaultProducer(exchangeName, routingKey string) *RabbitMQ {
	rr := &RabbitMQ{
		hostname:     "iw16.com",
		port:         5672,
		username:     "admin",
		password:     "wang*qing",
		exchangeType: "direct",
	}
	if exchangeName == "" {
		rr.exchangeName = "test"
	} else {
		rr.exchangeName = exchangeName
	}

	if routingKey == "" {
		rr.routingKey = "test"
	} else {
		rr.routingKey = routingKey
	}
	return rr
}

func (r *RabbitMQ) MsgProducer(msg string) {
	// 验证链接是否正常,否则重新链接
	if r.channel == nil {
		r.mqConnect()
	}
	// 发送任务消息
	err := r.channel.Publish(r.exchangeName, r.routingKey, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msg),
	})
	if err != nil {
		fmt.Printf("MQ任务发送失败:%s \n", err)
		return
	}
}

func TestSend() {
	qq := DefaultConnect()
	qq.MsgProducer("12312")
	qq.MsgProducer("24234")
}
