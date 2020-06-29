package config

import "os"

var root = os.Getenv("WEBROOT")

//WebRoot the root dir of web
var WebRoot = root

func init() {
	//os.Getwd获取当前文件所在路径
	if len(root) == 0 {
		webRoot, err := os.Getwd()
		if err != nil || len(webRoot) == 0 {
			panic("could not retrive workding directory")
		}
		WebRoot = webRoot + "/../.."
	}
}

//GetWebRoot 获取webroot
func GetWebRoot() string {
	return WebRoot
}
