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

	"github.com/HashCell/go-fileserver/mq"

	"github.com/HashCell/go-fileserver/common"
	"github.com/HashCell/go-fileserver/config"
	"github.com/HashCell/go-fileserver/db"
	"github.com/HashCell/go-fileserver/meta"
	"github.com/HashCell/go-fileserver/store/ceph"
	"github.com/HashCell/go-fileserver/store/oss"
	"github.com/HashCell/go-fileserver/util"
)

// UpdateFileMetaHandler 更新file meta
func UpdateFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	// fetch file hash-key to find file-meta
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	fileMeta := meta.GetFileMeta(fileSha1)
	fileMeta.FileName = newFileName
	meta.UpdateFileMeta(fileMeta)

	data, err := json.Marshal(fileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//GetFileMetaHandler 获取file meta
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// parse and get filehash
	// r.Form is ype Values map[string][]string, so its value is type of []string array
	// if we use Get method, it will return default values[0]
	fileSha1 := r.Form.Get("filehash")
	//fMeta := meta.GetFileMeta(fileSha1)

	// change to use mysql for query 2020.05.01 begin
	tableFile, err := meta.GetFileMetaDB(fileSha1)
	fmt.Println(tableFile)
	if err != nil {
		fmt.Printf("fail to get file meta, err %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Marshal return []byte
	data, err := json.Marshal(tableFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// use Write method to write []byte
	w.Write(data)
}

//QueryFileMetaHanler query recent file metas
func QueryFileMetaHanler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	//fileMetas := meta.GetLastFileMetas(limitCnt)

	fileMetas, err := meta.GetLastFileMetasDB(limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fileMetas)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

//DeleteFileHandler 删除文件
func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fileSha1 := r.Form.Get("filehash")
	fileMeta := meta.GetFileMeta(fileSha1)

	//remove the file
	err := os.Remove(fileMeta.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// remove file meta
	meta.RemoveFileMeta(fileSha1)
	// TODO: remove meta from database table

	w.WriteHeader(http.StatusOK)
}

//UploadHandler 文件上传
func UploadHandler(w http.ResponseWriter, req *http.Request) {

	// if get method, then return uploading-page
	if req.Method == "GET" {
		log.Println(config.WebRoot + "/static/view/index.html")
		data, err := ioutil.ReadFile(config.GetWebRoot() + "/static/view/index.html")
		if err != nil {
			io.WriteString(w, "internal server error")
			return
		}
		io.WriteString(w, string(data))
		//http.Redirect(w, req, "/static/view/index.html", http.StatusFound)
	} else if req.Method == "POST" {
		file, head, err := req.FormFile("file")
		if err != nil {
			fmt.Printf("fail to get data, err:%s\n", err.Error())
			return
		}
		defer file.Close()

		// read file, and save it the local file system
		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "/tmp/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		//  create a file on local file system
		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Printf("Fail to create file, err:%s\n", err.Error())
			return
		}
		defer newFile.Close()

		// copy the file
		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("fail to save save data into file, err: %s\n", err.Error())
			return
		}

		//compute file sha1
		newFile.Seek(0, 0)
		// save the file meta
		fileMeta.FileSha1 = util.FileSha1(newFile)
		//meta.UpdateFileMeta(fileMeta)

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
					w.Write([]byte("oss upload fail"))
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
		req.ParseForm()
		username := req.Form.Get("username")
		suc := db.OnUserFileUploadFinished(username, fileMeta.FileSha1,
			fileMeta.FileName, fileMeta.FileSize)

		if suc {
			http.Redirect(w, req, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("upload failed."))
		}
	}
}

//UploadSucHandler 上传成功后则重定向
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "upload file successfully!")
}

//DownloadHandler 获取下载url
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// r.Form is a type of map
	filesha1 := r.Form.Get("filehash")
	// find file meta according to filesha1
	fmt.Println("filesha1: " + filesha1)
	fileMeta, _ := meta.GetFileMetaDB(filesha1)
	// read file from file system, os.Open() is just used for reading
	f, err := os.Open(fileMeta.Location)
	if err != nil {
		fmt.Printf("fail to open file, err: %s\n", err.Error())
		return
	}
	defer f.Close()

	fmt.Printf("file name %s\n", f.Name())
	// read data from file and write them to response
	data, err := ioutil.ReadAll(f)
	//data, err := ioutil.ReadFile(f.Name())
	if err != nil {
		fmt.Printf("fail to read file, err: %s\n", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	// attachment means the file will be downloaded to save, instead of displaying on browser
	w.Header().Set("content-disposition", "attachment;filename=\""+fileMeta.FileName+"\"")
	w.Write(data)
}

//DownloadURLHandler 生成文件下载地址，为了向后兼容oss或ceph云存储
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	// 从文件表查找记录
	fmt.Println("filehash :" + filehash)
	row, _ := meta.GetFileMetaDB(filehash)

	// 根据location的前缀，判断文件下载的来源
	//本地/ceph集群/aliyun oss
	if strings.HasPrefix(row.Location, "/tmp") {
		username := r.Form.Get("username")
		token := r.Form.Get("token")
		tmpURL := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			r.Host, filehash, username, token)
		w.Write([]byte(tmpURL))
	} else if strings.HasPrefix(row.Location, "oss") {
		signedURL := oss.DownloadURL(row.Location)
		w.Write([]byte(signedURL))
	}
	return
}

// TryFastUploadHandler 实现秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	//1. 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))
	//2.　从文件表中查询是否有相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//3. 查不到记录就返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}
	//4. 否则通过秒传讲文件信息写入到用户文件表，返回成功
	suc := db.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}
	//可能出现数据库写入失败
	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败,请稍后重试",
	}
	w.Write(resp.JSONBytes())
	return
}

//FileQueryHandler 文件查询
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")

	userFiles, err := db.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
