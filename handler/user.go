package handler

import (
	"fmt"
	"net/http"

	"github.com/HashCell/go-fileserver/db"
	"github.com/HashCell/go-fileserver/util"
	"github.com/gin-gonic/gin"
)

const (
	//自定义加密盐值
	pwdSalt = "*#890"
)

//DoGetUserSignupHandler 用户注册
func DoGetUserSignupHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signup.html")
}

//DoPostUserSignupHandler 用户注册
func DoPostUserSignupHandler(c *gin.Context) {

	username := c.Request.FormValue("username")
	passwd := c.Request.FormValue("password")

	//TODO 定义code枚举
	if len(username) < 3 || len(passwd) < 3 {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "invalid parameter",
			"code": -1,
		})
		return
	}
	//2. 对用户密码进行加密
	encPwd := util.Sha1([]byte(passwd + pwdSalt))
	//3. 将用户信息保存到数据库表
	res := db.UserSignup(username, encPwd)
	if res {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "signup success",
			"data": nil,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "signup fail",
			"data": nil,
		})
	}
	return
}

//DoGetUserSigninHandler 获取登录界面
func DoGetUserSigninHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signin.html")
}

//DoPostUserSigninHandler 用户登录
func DoPostUserSigninHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")

	println(username + ":" + password)
	//1.用户名和密码校验
	encPwd := util.Sha1([]byte(password + pwdSalt))
	pwdChecked := db.UserSignIn(username, encPwd)
	if !pwdChecked {
		println("sign in fail")
		c.JSON(http.StatusOK, gin.H{
			"msg":  "signin fail",
			"code": 0,
		})
		return
	}

	//2.校验成功，生成token
	token := GenToken(username)
	upRes := db.UpdateToken(username, token)
	if !upRes {
		println("update token fail")
		c.JSON(http.StatusOK, gin.H{
			"msg":  "signin fail",
			"code": 0,
		})
		return
	}

	//TODO: return response message
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	c.String(http.StatusOK, string(resp.JSONBytes()))
	//c.Data(http.StatusOK, "octet-stream", resp.JSONBytes())
	return
}

//DoGetUserInfoHandler 获取用户信息
func DoGetUserInfoHandler(c *gin.Context) {
	//1.解析请求参数
	username := c.Request.FormValue("username")
	token := c.Request.FormValue("token")
	//2.验证token是否合法
	if !IsTokenValid(token) {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}
	user, err := db.GetUserInfo(username)
	fmt.Println(user)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}

	c.Data(http.StatusOK, "application/octet-stream", resp.JSONBytes())
	return
}
