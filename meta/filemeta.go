package meta

import "fmt"

type FileMeta struct {
	FileSha1 string
	FileName string
	Location string
	UploadAt string
	FileSize int64
}

var fileMetas map[string] FileMeta

func init()  {
	fileMetas = make(map[string]FileMeta)
}

/////////////////////////////////////
/// memory storage
////////////////////////////////////

// add or update file meta
func UpdateFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
	fmt.Printf("file meta: %v\n", fmeta)
}

// get file meta
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// remove file meta
func RemoveFileMeta(fileSha1 string)  {
	delete(fileMetas,fileSha1)
}

////////////////////////////////////////
// Database storage
///////////////////////////////////////

