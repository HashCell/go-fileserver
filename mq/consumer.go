package mq

import "fmt"

var processDone chan bool

//StartConsumer 开始消费
func StartConsumer(queueName, consumerName string, callback func(msg []byte) bool) {
	//1. 获取消费者通信通道
	msgChan, err := channel.Consume(
		queueName,
		consumerName,
		true,  //自动应答
		false, //多个消费者
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//2. 循环获取队列的消息
	//　不阻塞主线程，放到协程去处理
	processDone = make(chan bool)

	go func() {
		//遍历通道传过来的任务消息
		for msg := range msgChan {
			//3. 调用callback处理消息
			procSuc := callback(msg.Body)
			if !procSuc {
				//TODO: 将任务转移到另一个队列，用于异常情况的重试
			}
		}
	}()

	//processDone没有消息过来的话，就一直阻塞，不让startConsumer退出
	<-processDone

	//有消息过来就close掉rabbit通道
	channel.Close()

}
