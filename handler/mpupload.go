package handler

import (
	"net/http"
	"strconv"
	"github.com/HashCell/go-fileserver/util"
	rPool "github.com/HashCell/go-fileserver/cache/redis"
	"fmt"
	"time"
	"math"
	"os"
	"path"
	"github.com/garyburd/redigo/redis"
	"github.com/HashCell/go-fileserver/db"
	"strings"
	"os/exec"
)

//分块初始化信息
type MultipartuploadInfo struct {
	FileHash 	string
	FileSize 	int
	//标识分块传输的唯一键
	UploadID	string
	//分块的块大小
	ChunkSize	int
	//分块总数
	ChunkCount	int
}

//初始化分块上传
func InitiateMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {

	//1. 解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid",nil).JSONBytes())
		return
	}

	//2.　获取redis的一个连接
	redisConn := rPool.RedisPool().Get()
	defer redisConn.Close()

	//3.　生成分块上传的初始化信息
	mpInfo := MultipartuploadInfo{
		FileHash:filehash,
		FileSize:filesize,
		UploadID:username+fmt.Sprintf("%x",time.Now().UnixNano()),
		ChunkSize:5 * 1024 * 1024, //5MB
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}
	fmt.Println(mpInfo)

	//4. 将初始化数据返回到客户端
	redisConn.Do("HSET", "MP_"+mpInfo.UploadID, "chunkcount", mpInfo.ChunkCount)
	redisConn.Do("HSET", "MP_"+mpInfo.UploadID, "filehash", mpInfo.FileHash)
	redisConn.Do("HSET","MP_"+mpInfo.UploadID, "filesize", mpInfo.FileSize)

	//5. 将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0,"OK",mpInfo).JSONBytes())
}

func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	//1. 解析请求参数
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

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
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	fmt.Println(fpath)

	buf := make([]byte, 1024 * 1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	//4. 更新redis缓存状态
	redisConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)
	//5. 返回处理结果给客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

//客户端完成分块上传，通知服务段合并分块
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	//1. 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	//2. 获得redis连接池的连接
	redisConn := rPool.RedisPool().Get()
	defer redisConn.Close()

	//3. 通过uploadid查询redis并判断是否所有分块都完成上传
	dataArr, err := redis.Values(redisConn.Do("HGETALL", "MP_"+uploadID))
	if err != nil {
		fmt.Println(err.Error())
		w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
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
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}
	//4. 合并分块，得到完整的文件,使用linux shell脚本完成合并
	targetDir := "/data/file-server/files/"
	targetFile := targetDir + filename
	srcFile := "/data/"+uploadID
	cmd := exec.Command("./script/shell/merge_file_blocks.sh", srcFile, targetFile)
	if _, err := cmd.Output(); err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(util.NewRespMsg(-2, "fail", nil).JSONBytes())
		return
	} else {
		fmt.Println(targetFile + " has been merged completely")
	}
	//4.1 TODO: 如果合并后的文件仍然存放在本地，则应该以uploadId区分文件目录
	//	  TODO: 后面使用私有云或公有云，结合rabbitMQ异步地将该大文件转移到云上
	//5. 更新唯一文件表和用户文件表
	fsize, _ := strconv.Atoi(filesize)
	//file address remains "" for future implement, such as ceph, oss
	db.OnFileUploadFinished(filehash, filename, int64(fsize), "")
	db.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	//6. 响应处理结果给客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

//取消分块上传
func CancelMultipartUpload(w http.ResponseWriter, r *http.Request) {

}

//查看分块上传的状态
func MultipartUploadStatusHandler(w http.ResponseWriter, r *http.Request) {

}