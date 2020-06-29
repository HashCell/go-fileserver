package db

import (
	"database/sql"
	"fmt"

	mydb "github.com/HashCell/go-fileserver/db/mysql"
)

//TableFile 结构体
type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// OnFileUploadFinished finished to upload file then store file meta to database
func OnFileUploadFinished(filehash string, filename string, filesize int64, fileaddr string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_file (`file_sha1`,`file_name`,`file_size`," +
			"`file_addr`,`status`) values (?,?,?,?,?)")

	if err != nil {
		fmt.Printf("fail to prepare statement , err: %s\n", err.Error())
		return false
	}

	defer stmt.Close()

	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr, 1)
	if err != nil {
		fmt.Printf("fail to exe statement, err: %s\n", err.Error())
		return false
	}

	// use RowsAffected method to check whether the sql operation has effected
	if rf, err := ret.RowsAffected(); err == nil {
		if rf <= 0 {
			fmt.Printf("warning! statement execute succefully but not row effected.")
		}
		return true
	}
	return false
}

//GetFileMeta 获取file meta
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size " +
			" from tbl_file where file_sha1=? and status=1 limit 1")

	if err != nil {
		fmt.Printf("fail to prepare statement, err %s\n", err.Error())
		return nil, err
	}
	defer stmt.Close()

	tableFiles, err := GetFileMetaList(1)
	fmt.Println(tableFiles)

	tableFile := TableFile{}
	err = stmt.QueryRow(filehash).Scan(
		&tableFile.FileHash, &tableFile.FileAddr, &tableFile.FileName, &tableFile.FileSize)
	if err != nil {
		fmt.Println(err.Error())
	}
	return &tableFile, nil
}

// GetFileMetaList 获取file meta list
func GetFileMetaList(limit int) ([]TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_name,file_size,file_addr from tbl_file" +
			" where status=1 limit ?")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	colums, err := rows.Columns()
	values := make([]sql.RawBytes, len(colums))
	var resultFiles []TableFile
	for i := 0; i < len(values) && rows.Next(); i++ {
		tFile := TableFile{}
		err = rows.Scan(&tFile.FileHash, &tFile.FileName, &tFile.FileSize, &tFile.FileAddr)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		resultFiles = append(resultFiles, tFile)
	}
	return resultFiles, nil
}

//UpdateFileLocation 更新文件存储位置用于ceph或oss存储后
func UpdateFileLocation(filehash string, fileaddr string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_file set `file_addr`=? where `file_sha1`=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	defer stmt.Close()

	ret, err := stmt.Exec(fileaddr, filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if rf, err := ret.RowsAffected(); nil == err {
		if rf == 0 {
			fmt.Printf("文件 %s 已存在,无法更新\n", filehash)
		}
		return true
	}
	return false
}
