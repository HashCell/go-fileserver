package handler

import (
	"net/http"
	"fmt"
	"github.com/HashCell/go-fileserver/util"
	"github.com/HashCell/go-fileserver/db"
)
const (
	//自定义加密盐值
	pwdSalt = "*#890"
)

// 用户注册
func UserSignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w,r,"/static/view/signup.html", http.StatusFound)
	}
	//1. 获取用户账户密码＋简单校验
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	fmt.Println(username +":"+password)

	if len(username) < 3 || len(password) < 3 {
		w.Write([]byte("Invalid parameter"))
		return
	}
	//2. 对用户密码进行加密
	encPwd := util.Sha1([]byte(password+pwdSalt))
	//3. 将用户信息保存到数据库表
	res := db.UserSignup(username, encPwd)
	if res {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("sign up fail!"))
	}
}

//　用户登录
func UserSigninHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/static/view/signin.html",http.StatusFound)
		return
	}

	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	println(username+":"+password)
	//1.用户名和密码校验
	encPwd := util.Sha1([]byte(password+pwdSalt))
	pwdChecked := db.UserSignIn(username, encPwd)
	if !pwdChecked {
		w.Write([]byte("sign in fail, username or password invalid"))
		return
	}

	//2.校验成功，生成token
	token := GenToken(username)
	upRes := db.UpdateToken(username, token)
	if !upRes {
		w.Write([]byte("fail to update token"))
		return
	}

	//TODO: return response message
	resp := util.RespMsg{
		Code:0,
		Msg:"OK",
		Data: struct {
			Location string
			Username string
			Token string
		}{
			Location:"http://" + r.Host + "/static/view/home.html",
			Username:username,
			Token:token,
		},
	}
	w.Write(resp.JSONBytes())
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	//1.解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	token := r.Form.Get("token")
	//2.验证token是否合法
	if !IsTokenValid(token) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	user, err := db.GetUserInfo(username)
	fmt.Println(user)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	resp := util.RespMsg{
		Code:0,
		Msg:"OK",
		Data:user,
	}

	w.Write(resp.JSONBytes())
}


