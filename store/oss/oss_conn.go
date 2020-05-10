package oss

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	cfg "github.com/HashCell/go-fileserver/config"
	"fmt"
)

var ossCli *oss.Client

func Client() *oss.Client {
	if ossCli != nil {
		return ossCli
	}

	//1. 获取aliyun　oss的认证连接
	ossCli, err := oss.New(cfg.OSSEndpoint,
		cfg.OSSAccesskeyID, cfg.OSSAccesskeySecret)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return ossCli
}

// Bucket: 获取bucket存储空间
func Bucket()  *oss.Bucket {
	cli := Client()
	if cli != nil {
		bucket, err := cli.Bucket(cfg.OSSBucket)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		return bucket
	}
	return nil
}

// 临时授权下载URL
/**
* objName: ossPath
 */
func DownloadURL(objName string) string {
	signedURL, err := Bucket().SignURL(objName, oss.HTTPGet, 3600)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return signedURL
}

//设置bucket里test/前缀路径下的文件对象的生命周期为30天
func BuildLifecycleRule(bucketName string)  {
	ruleTest1 := oss.BuildLifecycleRuleByDays("rule1", "test/", true, 30)
	rules := []oss.LifecycleRule{ruleTest1}
	Client().SetBucketLifecycle(bucketName,rules)
}

