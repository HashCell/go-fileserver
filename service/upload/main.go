package main

import (
	"log"
	"net/http"

	"github.com/HashCell/go-fileserver/config"

	"github.com/HashCell/go-fileserver/handler"
)

func main() {

	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/downloadurl", handler.DownloadURLHandler)

	http.HandleFunc("/file/delete", handler.DeleteFileHandler)
	http.HandleFunc("/file/meta/update", handler.UpdateFileMetaHandler)
	http.HandleFunc("/file/meta", handler.HttpInterceptor(handler.GetFileMetaHandler))
	http.HandleFunc("/file/meta/query", handler.HttpInterceptor(handler.QueryFileMetaHanler))
	http.HandleFunc("/file/query", handler.FileQueryHandler)

	http.HandleFunc("/user/signup", handler.UserSignupHandler)
	http.HandleFunc("/user/signin", handler.UserSigninHandler)
	http.HandleFunc("/user/info", handler.UserInfoHandler)

	// multipart upload
	http.HandleFunc("/file/mpupload/init", handler.InitiateMultipartUploadHandler)
	http.HandleFunc("/file/mpupload/split", handler.UploadPartHandler)
	http.HandleFunc("/file/mpupload/complete", handler.CompleteUploadHandler)

	//webRoot+"/static"绝对路径
	//http://localhost:8080/static/view/signin.html trip掉/static，剩下/view/signin.html到webRoot+"/static目录去找
	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(webRoot+"/static"))))
	log.Println(http.Dir(config.WebRoot + "/static"))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(config.GetWebRoot()+"/static"))))
	log.Println("文件上传服务开始监听...")
	http.ListenAndServe(":8080", nil)

}
