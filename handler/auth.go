package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/HashCell/go-fileserver/util"
	"github.com/gin-gonic/gin"
)

//HTTPInterceptor 拦截器
func HTTPInterceptor() gin.HandlerFunc {
	//convert the func(w,r) to http.HandlerFunc
	return func(c *gin.Context) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")

		if len(username) < 3 || !IsTokenValid(token) {
			c.Abort() // 终止传递
			resp := util.NewRespMsg(-4, "token无效", nil)
			c.JSON(http.StatusOK, resp)
			return
		}
		c.Next()
	}
}

// GenToken : 生成token
func GenToken(username string) string {
	// 40位字符:md5(username+timestamp+token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

// IsTokenValid : token是否有效
func IsTokenValid(token string) bool {
	//if len(token) != 40 {
	//	return false
	//}
	// TODO: 判断token的时效性，是否过期
	// TODO: 从数据库表tbl_user_token查询username对应的token信息
	// TODO: 对比两个token是否一致
	return true
}
