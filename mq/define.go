package mq

import "github.com/HashCell/go-fileserver/common"

//定义消息队列的消息载体的结构格式
type TransferData struct {
	Filehash string
	CurLocation string
	DestLocation string
	DestStoreType common.StoreType
}