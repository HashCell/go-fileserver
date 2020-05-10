package config

const (
	//是否开启文件异步转移
	AsyncTransferEnable = true
	// rabbitmq服务的入口url, amqp协议
	RabbitURL = "amqp://guest:guest@127.0.0.1:5672/"
	//用于文件转移的交换机
	TransExchangeName = "uploadserver.transfer"
	//oss转移队列名
	TransOSSQueueName = "uploadserver.transfer.oss"
	//oss转移失败后写入到另一个队列
	TransOSSErrQueueName = "uploadserver.transfer.oss.err"
	//routing key
	TransOSSRoutingKey = "oss"
)
