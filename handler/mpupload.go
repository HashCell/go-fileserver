package handler

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	rPool "github.com/HashCell/go-fileserver/cache/redis"
	"github.com/HashCell/go-fileserver/db"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
)

//MultipartuploadInfo 分块初始化信息
type MultipartuploadInfo struct {
	FileHash string
	FileSize int
	//标识分块传输的唯一键
	UploadID string
	//分块的块大小
	ChunkSize int
	//分块总数
	ChunkCount int
}

const (
	chunkSize     int    = 5 * 1024 * 1024
	hSetKeyPrefix string = "MP_"
)

//InitiateMultipartUploadHandler 初始化分块上传
func InitiateMultipartUploadHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filesize, err := strconv.Atoi(c.Request.FormValue("filesize"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "params invalid",
		})
		return
	}

	//2.　获取redis的一个连接
	redisConn := rPool.RedisPool().Get()
	defer redisConn.Close()

	//3.　生成分块上传的初始化信息
	mpInfo := MultipartuploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  chunkSize, //5MB
		ChunkCount: int(math.Ceil(float64(filesize / chunkSize))),
	}
	fmt.Println(mpInfo)

	//4. 将初始化数据返回到客户端
	redisConn.Do("HSET", hSetKeyPrefix+mpInfo.UploadID, "chunkcount", mpInfo.ChunkCount)
	redisConn.Do("HSET", hSetKeyPrefix+mpInfo.UploadID, "filehash", mpInfo.FileHash)
	redisConn.Do("HSET", hSetKeyPrefix+mpInfo.UploadID, "filesize", mpInfo.FileSize)

	//5. 将响应初始化数据返回到客户端
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "OK",
		"data": mpInfo,
	})
}

//UploadPartHandler 分块上传
func UploadPartHandler(c *gin.Context) {
	//1. 解析请求参数
	uploadID := c.Request.FormValue("uploadid")
	chunkIndex := c.Request.FormValue("index")

	//2. 获取redis连接
	redisConn := rPool.RedisPool().Get()
	defer redisConn.Close()

	//3.　获取文件句柄，用于存储分块内容
	fpath := "/data/" + uploadID + "/" + chunkIndex
	fmt.Println(path.Dir(fpath))
	err := os.MkdirAll(path.Dir(fpath), 0744)
	if err != nil {
		fmt.Println(err.Error())
	}
	fd, err := os.Create(fpath)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "upload part failed",
			"data": nil,
		})
		return
	}
	defer fd.Close()

	fmt.Println(fpath)
	//读取内存中分块内容写入到文件
	buf := make([]byte, 1024*1024)
	for {
		n, err := c.Request.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	//4. 更新redis缓存状态
	redisConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)
	//5. 返回处理结果给客户端
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": nil,
	})
}

//CompleteUploadHandler 客户端完成分块上传，通知服务段合并分块
func CompleteUploadHandler(c *gin.Context) {
	//1. 解析请求参数
	username := c.Request.FormValue("username")
	uploadID := c.Request.FormValue("uploadid")
	filehash := c.Request.FormValue("filehash")
	filesize := c.Request.FormValue("filesize")
	filename := c.Request.FormValue("filename")

	//2. 获得redis连接池的连接
	redisConn := rPool.RedisPool().Get()
	defer redisConn.Close()

	//3. 通过uploadid查询redis并判断是否所有分块都完成上传
	dataArr, err := redis.Values(redisConn.Do("HGETALL", "MP_"+uploadID))
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "complete upload failed",
			"data": nil,
		})
		return
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(dataArr); i += 2 {
		k := string(dataArr[i].([]byte))
		v := string(dataArr[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}
	fmt.Println("total: ", totalCount, "chunkcount: ", chunkCount)
	if totalCount != chunkCount {
		c.JSON(http.StatusOK, gin.H{
			"code": -2,
			"msg":  "invalid request",
			"data": nil,
		})
		return
	}
	//4. 合并分块，得到完整的文件,使用linux shell脚本完成合并
	targetDir := "/data/file-server/files/"
	targetFile := targetDir + filename
	srcFile := "/data/" + uploadID
	cmd := exec.Command("./script/shell/merge_file_blocks.sh", srcFile, targetFile)
	if _, err := cmd.Output(); err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -2,
			"msg":  "merge file blocks fail",
			"data": nil,
		})
		return
	}
	//4.1 TODO: 如果合并后的文件仍然存放在本地，则应该以uploadId区分文件目录
	//	  TODO: 后面使用私有云或公有云，结合rabbitMQ异步地将该大文件转移到云上
	//5. 更新唯一文件表和用户文件表
	fsize, _ := strconv.Atoi(filesize)
	//file address remains "" for future implement, such as ceph, oss
	fileuploadfinished := db.OnFileUploadFinished(filehash, filename, int64(fsize), "")
	userfileuploadfinished := db.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	fmt.Printf("fileuploadfinished:%t	userfileuploadfinished:%t", fileuploadfinished, userfileuploadfinished)
	//6. 响应处理结果给客户端
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "finished",
		"data": nil,
	})
}

//CancelMultipartUpload 取消分块上传
func CancelMultipartUpload(w http.ResponseWriter, r *http.Request) {

}

//MultipartUploadStatusHandler 查看分块上传的状态
func MultipartUploadStatusHandler(w http.ResponseWriter, r *http.Request) {

}
