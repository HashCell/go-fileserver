package main

import (
	"net/http"
	"github.com/HashCell/go-fileserver/handler"
)

func main() {
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/update",handler.UpdateFileMetaHandler)
	http.HandleFunc("/file/delete",handler.DeleteFileHandler)
	http.ListenAndServe(":8080", nil)
}
