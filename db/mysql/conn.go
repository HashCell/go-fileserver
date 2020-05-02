package mysql

// we must import github.com/go-sql-driver/mysql to execute its init method
import (
	"database/sql"
	"fmt"
	"os"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB

func init() {
	db,_ = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/fileserver?charset=utf8")
	db.SetMaxOpenConns(1000)
	db.SetMaxIdleConns(100)
	err := db.Ping()
	if err != nil {
		fmt.Printf("fail to connect to mysql, err %s \n", err.Error())
		os.Exit(1)
	}
}

// return the DB instance
func DBConn() *sql.DB {
	return db
}

func ParseRows(rows *sql.Rows) []map[string]interface{} {
	colums, _ := rows.Columns()

	//用于从row获取每一行的值
	scanArgs := make([]interface{}, len(colums))
	values := make([]interface{}, len(colums))

	for j := range values {
		scanArgs[j] = &values[j]
	}

	// 用字典来存放每一行的列名和对应的值
	record := make(map[string]interface{})
	// 用字典数组来存放所有行
	records := make([]map[string]interface{}, 0)


	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			log.Fatal(err)
			panic(err)
		}

		//将获取到的一行记录写入到record
		for i, colValue := range values {
			if colValue != nil {
				record[colums[i]] = colValue
			}
		}

		records = append(records, record)
	}
	return records
}