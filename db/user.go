package db

import (
	mydb "github.com/HashCell/go-fileserver/db/mysql"
	"fmt"
)

type UserInfo struct {
	Username 		string
	Email 			string
	Phone			string
	SignupAt		string
	LastActiveAt	string
	Status 			int
}




// 用户注册
func UserSignup(username string, password string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user (`user_name`,`user_pwd`) values (?,?)")

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	defer stmt.Close()

	ret, err := stmt.Exec(username, password)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if rowsAffected, err := ret.RowsAffected(); nil == err && rowsAffected > 0 {
		println("okkk")
		return true
	}

	return false
}

// 用户登录
func UserSignIn(username string, encpwd string) bool {
	stmt, err := mydb.DBConn().Prepare("select * from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		fmt.Println("username not found: " + username)
	}

	pRowsMapArr := mydb.ParseRows(rows)
	// string(pRowsMapArr[0]["user_pwd"].([]byte)) !!!
	if len(pRowsMapArr) > 0 && string(pRowsMapArr[0]["user_pwd"].([]byte)) == encpwd {
		return true
		println("heheh")
	}
	return false
}

func UpdateToken(username string, token string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"replace into tbl_user_token (`user_name`,`user_token`) values (?,?)")

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	defer stmt.Close()
	_,err = stmt.Exec(username,token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func GetUserInfo(username string) (UserInfo, error) {
	user := UserInfo{}
	stmt, err := mydb.DBConn().Prepare(
		"select `user_name`, `signup_at` from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.Username,&user.SignupAt)
	if err != nil {
		return user, err
	}
	return user, nil
}