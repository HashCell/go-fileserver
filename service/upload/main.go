package upload

import (
	"net/http"
	"github.com/HashCell/go-fileserver/handler"
	"os"
)

func main() {

	webRoot := os.Getenv("WEBROOT")
	//os.Getwd获取当前文件所在路径
	if len(webRoot) == 0 {
		if root, err := os.Getwd();err != nil {
			panic("could not retrive workding directory")
		} else {
			webRoot = root
		}
	}

	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/downloadurl",handler.DownloadURLHandler)

	http.HandleFunc("/file/delete",handler.DeleteFileHandler)
	http.HandleFunc("/file/meta/update",handler.UpdateFileMetaHandler)
	http.HandleFunc("/file/meta",handler.HttpInterceptor(handler.GetFileMetaHandler))
	http.HandleFunc("/file/meta/query", handler.HttpInterceptor(handler.QueryFileMetaHanler))
	http.HandleFunc("/file/query",handler.FileQueryHandler)

	http.HandleFunc("/user/signup",handler.UserSignupHandler)
	http.HandleFunc("/user/signin",handler.UserSigninHandler)
	http.HandleFunc("/user/info",handler.UserInfoHandler)

	// multipart upload
	http.HandleFunc("/file/mpupload/init", handler.InitiateMultipartUploadHandler)
	http.HandleFunc("/file/mpupload/split", handler.UploadPartHandler)
	http.HandleFunc("/file/mpupload/complete", handler.CompleteUploadHandler)

	//webRoot+"/static"绝对路径
	//http://localhost:8080/static/view/signin.html trip掉/static，剩下/view/signin.html到webRoot+"/static目录去找
	http.Handle("/static/",http.StripPrefix("/static/",http.FileServer(http.Dir(webRoot+"/static"))))
	http.ListenAndServe(":8080", nil)
}


