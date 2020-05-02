package util

import (
	"encoding/json"
	"log"
	"fmt"
)

type RespMsg struct {
	Code int 			`json:"code"`
	Msg string 			`json:"msg"`
	Data interface{} 	`json:"data"`
}

// JSONBytes: 对象转json格式的二进制数组
func (resp *RespMsg) JSONBytes() []byte {
	r, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	return r
}

func (resp *RespMsg) JSONString() string {
	r, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	return string(r)
}

//GenSimpleRespStream:只包含code和message的response body
func GenSimpleRespStream(code int, msg string) []byte {
	return []byte(fmt.Sprintf(`{"code":%d,"msg":"%s"}`, code, msg))
}

//GenSimpleRespString:只包含code和message的response body
func GenSimpleRespString(code int, msg string) string {
	return fmt.Sprintf(`{"code":%d,"msg":"%s"}`, code, msg)
}