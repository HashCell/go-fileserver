package mq

import (
	"fmt"
	"log"

	"github.com/HashCell/go-fileserver/config"
	"github.com/streadway/amqp"
)

//rabbitmq服务的连接
var conn *amqp.Connection

//与rabbitmq服务通信的通道
var channel *amqp.Channel

var notifyClose chan *amqp.Error

func init() {
	//没有开启异步转移，仅当开启时才初始化rabbitMQ连接
	if !config.AsyncTransferEnable {
		return
	}
	if initChannel() {
		channel.NotifyClose(notifyClose)
	}

	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				log.Printf("on notify channel closed: %+v\n", msg)
				initChannel()
			}
		}
	}()
}

// 初始化channel
func initChannel() bool {
	//1. 检查channel是否已经初始化
	if channel != nil {
		return true
	}

	//2. 获取rabbitmq的连接
	conn, err := amqp.Dial(config.RabbitURL)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	//3.通过amqp的连接，获取channel
	channel, err = conn.Channel()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

//Publish 生产者发布消息
func Publish(exchanger string, routingKey string, msg []byte) bool {
	//1. 判断channel是否正常
	if !initChannel() {
		return false
	}
	//2. 执行消息发布
	err := channel.Publish(
		exchanger,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		},
	)

	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
