package main

import (
	"net/http"
	"github.com/HashCell/go-fileserver/handler"
)

func main() {
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.ListenAndServe(":8080", nil)
}
