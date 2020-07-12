package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/HashCell/go-fileserver/mq"

	"github.com/HashCell/go-fileserver/common"
	"github.com/HashCell/go-fileserver/config"
	"github.com/HashCell/go-fileserver/db"
	"github.com/HashCell/go-fileserver/meta"
	"github.com/HashCell/go-fileserver/store/ceph"
	"github.com/HashCell/go-fileserver/store/oss"
	"github.com/HashCell/go-fileserver/util"
)

// DoPostUpdateFileMetaHandler 更新file meta
func DoPostUpdateFileMetaHandler(c *gin.Context) {
	opType := c.Request.FormValue("op")
	fileSha1 := c.Request.FormValue("filehash")
	newFileName := c.Request.FormValue("filename")

	if opType != "0" {
		c.Status(http.StatusForbidden)
		return
	}

	fileMeta, err := meta.GetFileMetaDB(fileSha1)
	if err != nil {
		fmt.Println("post update filemeta fail")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -2,
			"msg":  "file is not exist",
		})
		return
	}
	if fileMeta.FileName != newFileName {
		fileMeta.FileName = newFileName
		meta.UpdateFileMetaDB(fileMeta)
	}

	data, err := json.Marshal(fileMeta)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}

//DoGetFileMetaHandler 获取file meta
func DoGetFileMetaHandler(c *gin.Context) {
	fileSha1 := c.Request.FormValue("filehash")
	tableFile, err := meta.GetFileMetaDB(fileSha1)
	fmt.Println(tableFile)
	if err != nil {
		fmt.Printf("fail to get file meta, err %s\n", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -2,
			"msg":  "cannot find file meta",
		})
		return
	}

	if tableFile.FileSha1 != "" {
		data, err := json.Marshal(tableFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": -3,
				"msg":  "marshal error",
			})
			return
		}
		c.Data(http.StatusOK, "application/json", data)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": -4,
			"msg":  "no such file",
		})
	}
}

//DoGetQueryFileMetaHanler 查询最近的filemeta
func DoGetQueryFileMetaHanler(c *gin.Context) {
	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	fileMetas, err := meta.GetLastFileMetasDB(limitCnt)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fileMetas)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

//DoDeleteFileHandler 删除文件
func DoDeleteFileHandler(c *gin.Context) {
	fileSha1 := c.Request.FormValue("filehash")
	fileMeta := meta.GetFileMeta(fileSha1)
	err := os.Remove(fileMeta.Location)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	meta.RemoveFileMeta(fileSha1)
	c.Status(http.StatusOK)
}

//DoGetUploadHandler 获取文件上传页面
func DoGetUploadHandler(c *gin.Context) {
	data, err := ioutil.ReadFile("./static/view/index.html")
	if err != nil {
		fmt.Println(err)
		c.String(http.StatusNotFound, "not found")
		return
	}
	c.Header("Content-Type", "text/html;charset=utf-8")
	c.String(http.StatusOK, string(data))
	//c.Data(http.StatusOK, "text/html", data)
	//http.Redirect(w, req, "/static/view/index.html", http.StatusFound)
}

//DoPostUploadHandler 文件上传
func DoPostUploadHandler(c *gin.Context) {

	file, head, err := c.Request.FormFile("file")
	if err != nil {
		fmt.Printf("fail to get data, err:%s\n", err.Error())
		return
	}
	defer file.Close()

	fileMeta := &meta.FileMeta{
		FileName: head.Filename,
		Location: "/tmp/" + head.Filename,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	//本地创建文件接收
	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		fmt.Printf("Fail to create file, err:%s\n", err.Error())
		return
	}
	defer newFile.Close()

	//将网络文件拷贝到本地文件
	fileMeta.FileSize, err = io.Copy(newFile, file)
	if err != nil {
		fmt.Printf("fail to save save data into file, err: %s\n", err.Error())
		return
	}

	//游标回到文件头部，计算sha1值
	newFile.Seek(0, 0)
	fileMeta.FileSha1 = util.FileSha1(newFile)

	//游标回到文件头部
	newFile.Seek(0, 0)
	if config.CurrentStoreType == common.StoreCeph {
		data, _ := ioutil.ReadAll(newFile)
		cephPath := "/ceph/" + fileMeta.FileSha1
		_ = ceph.PutObject("userfile", cephPath, data)
		fileMeta.Location = cephPath
	} else if config.CurrentStoreType == common.StoreOSS {
		// 上传到阿里云oss
		//为了方便在阿里云预览文件，将文件名带上从而带上具体文件格式
		// ossPath中的文件路径将会在阿里云oss创建，path不要以 / 开头
		ossPath := "oss/" + fileMeta.FileSha1 + "_" + fileMeta.FileName

		//判断是使用同步还是异步，异步则使用rabbitmq
		if !config.AsyncTransferEnable {
			err = oss.Bucket().PutObject(ossPath, newFile)
			if err != nil {
				fmt.Println(err.Error())
				c.JSON(http.StatusOK, gin.H{
					"code": -1,
					"msg":  "oss upload fail",
				})
				return
			}
			fileMeta.Location = ossPath
		} else {
			//写入rabbitmq异步转移队列
			data := mq.TransferData{
				Filehash:      fileMeta.FileSha1,
				CurLocation:   fileMeta.Location,
				DestLocation:  ossPath,
				DestStoreType: common.StoreOSS,
			}

			publishData, _ := json.Marshal(data)
			isSuc := mq.Publish(
				config.TransExchangeName,
				config.TransOSSRoutingKey,
				publishData)
			if !isSuc {
				log.Println("转移到异步队列失败")
			}
		}
	}

	// 更新文件表
	_ = meta.UpdateFileMetaDB(fileMeta)

	//由于引入了秒传功能，所以还需要更新用户文件表
	username := c.Request.FormValue("username")
	suc := db.OnUserFileUploadFinished(username, fileMeta.FileSha1,
		fileMeta.FileName, fileMeta.FileSize)

	if suc {
		c.Redirect(http.StatusFound, "/static/view/home.html")
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "upload fail caused updating userfile failed",
		})
	}
}

//DoGetUploadSucHandler 上传成功后则重定向
func DoGetUploadSucHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "upload file successfully!",
	})
	//io.WriteString(w, "upload file successfully!")
}

//DoGetFileDownloadHandler 获取下载url
func DoGetFileDownloadHandler(c *gin.Context) {
	filesha1 := c.Request.FormValue("filehash")
	fileMeta, _ := meta.GetFileMetaDB(filesha1)
	c.FileAttachment(fileMeta.Location, fileMeta.FileName)
}

//DoGetDownloadURLHandler 生成文件下载地址
func DoGetDownloadURLHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")
	// 从文件表查找记录
	row, _ := meta.GetFileMetaDB(filehash)

	// 根据location的前缀，判断文件下载的来源
	//本地/ceph集群/aliyun oss
	if strings.HasPrefix(row.Location, "/tmp") {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		tmpURL := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			c.Request.Host, filehash, username, token)
		c.Data(http.StatusOK, "octet-stream", []byte(tmpURL))
		//w.Write([]byte(tmpURL))
	} else if strings.HasPrefix(row.Location, "/ceph") {
		//TODO:ceph download url
	} else if strings.HasPrefix(row.Location, "oss") {
		signedURL := oss.DownloadURL(row.Location)
		c.Data(http.StatusOK, "octet-stream", []byte(signedURL))
	}
	return
}

// TryFastUploadHandler 实现秒传接口
func TryFastUploadHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize, _ := strconv.Atoi(c.Request.FormValue("filesize"))
	//2.　从文件表中查询是否有相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	//3. 查不到记录就返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}
	//4. 否则通过秒传讲文件信息写入到用户文件表，返回成功
	suc := db.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}
	//可能出现数据库写入失败
	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败,请稍后重试",
	}
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
	return
}

//DoPostFileQueryHandler 文件查询
func DoPostFileQueryHandler(c *gin.Context) {
	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	username := c.Request.FormValue("username")

	userFiles, err := db.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"msg":  "queue failed",
		})
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -2,
			"msg":  "marshal error",
		})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}
