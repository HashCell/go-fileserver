package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/HashCell/go-fileserver/config"
	"github.com/HashCell/go-fileserver/db"
	"github.com/HashCell/go-fileserver/mq"
	"github.com/HashCell/go-fileserver/store/oss"
)

// ProcessTransfer 转移数据
func ProcessTransfer(msg []byte) bool {
	log.Println(string(msg))
	//解析msg
	publishData := mq.TransferData{}
	err := json.Unmarshal(msg, &publishData)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	//根据临时存储路径，创建文件句柄
	fin, err := os.Open(publishData.CurLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	//通过文件句柄讲文件内容读出来并且上传到oss
	err = oss.Bucket().PutObject(
		publishData.DestLocation,
		bufio.NewReader(fin))

	if err != nil {
		log.Println(err.Error())
		return false
	}

	//更新文件的存储路径到文件表
	db.UpdateFileLocation(
		publishData.Filehash,
		publishData.DestLocation)
	log.Printf("文件 %s 成功转移到oss\n", publishData.Filehash)
	return true
}

func main() {

	if !config.AsyncTransferEnable {
		log.Println("异步转移文件功能未开启...")
		return
	}
	log.Println("开始监听转移任务队列...")
	mq.StartConsumer(
		config.TransOSSQueueName,
		"transfer_oss", //可以自定义
		ProcessTransfer)
}
