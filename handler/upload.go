package handler

import (
	"io"
	"io/ioutil"
	"net/http"
	"fmt"
	"github.com/HashCell/go-fileserver/meta"
	"time"
	"os"
	"github.com/HashCell/go-fileserver/util"
)

func UploadHandler(w http.ResponseWriter, req *http.Request) {

	// if get method, then return uploading-page
	if req.Method == "GET" {
		data, err := ioutil.ReadFile("./static/view/index.html")
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
		newFile.Seek(0,0)
		// save the file meta
		fileMeta.FileSha1 = util.FileSha1(newFile)
		meta.UpdateFileMeta(fileMeta)

		io.WriteString(w, "upload successfully")
	}
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// r.Form is a type of map
	filesha1 := r.Form.Get("filehash")
	// find file meta according to filesha1
	fileMeta := meta.GetFileMeta(filesha1)
	// read file from file system, os.Open() is just used for reading
	f, err := os.Open(fileMeta.Location)
	if err != nil {
		fmt.Printf("fail to open file, err: %s\n", err.Error())
		return
	}

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
	w.Header().Set("content-disposition","attachment;filename=\""+fileMeta.FileName+"\"")
	w.Write(data)
}

/**
* update file meta, only support rename file
 */
func UpdateFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	// fetch file hash-key to find file-meta
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFilename := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
	}

}

func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {

}

