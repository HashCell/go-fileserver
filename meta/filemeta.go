package meta

import (
	"fmt"
	"sort"

	"github.com/HashCell/go-fileserver/db"
)

type FileMeta struct {
	FileSha1 string
	FileName string
	Location string
	UploadAt string
	FileSize int64
}

var fileMetas map[string]FileMeta

func init() {
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

// UpdateFileMetaDB store into database
func UpdateFileMetaDB(fmete *FileMeta) bool {
	return db.OnFileUploadFinished(
		fmete.FileSha1, fmete.FileName, fmete.FileSize, fmete.Location)
}

// get file meta
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

func GetFileMetaDB(filesha1 string) (*FileMeta, error) {
	tableFile, err := db.GetFileMeta(filesha1)
	if err != nil || tableFile == nil {
		return nil, err
	}

	var fileMeta = FileMeta{
		FileSha1: tableFile.FileHash,
		FileSize: tableFile.FileSize.Int64,
		FileName: tableFile.FileName.String,
		Location: tableFile.FileAddr.String,
	}
	fmt.Println(fileMeta)
	return &fileMeta, nil
}

// remove file meta
func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}

func GetLastFileMetas(count int) []FileMeta {
	fileMetaArray := make([]FileMeta, len(fileMetas))
	for _, v := range fileMetas {
		fileMetaArray = append(fileMetaArray, v)
	}

	// transfer []FileMeta to ByUploadTime
	sort.Sort(ByUploadTime(fileMetaArray))
	return fileMetaArray[0:count]
}

func GetLastFileMetasDB(limit int) ([]FileMeta, error) {

	tfiles, err := db.GetFileMetaList(limit)
	if err != nil {
		return make([]FileMeta, 0), err
	}

	resultFiles := make([]FileMeta, len(tfiles))
	for i := 0; i < len(tfiles); i++ {
		resultFiles[i] = FileMeta{
			FileSha1: tfiles[i].FileHash,
			FileName: tfiles[i].FileName.String,
			FileSize: tfiles[i].FileSize.Int64,
			Location: tfiles[i].FileAddr.String,
		}
	}

	return resultFiles, nil
}
