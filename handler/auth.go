package handler

import (
	"net/http"
	"fmt"
	"time"
	"github.com/HashCell/go-fileserver/util"
)

func HttpInterceptor(h http.HandlerFunc) http.HandlerFunc {
	//convert the func(w,r) to http.HandlerFunc
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			username := r.Form.Get("username")
			token := r.Form.Get("token")
			
			if len(username) < 3 || !IsTokenValid(token) {
				http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
				return
			}
			h(w,r)
		})
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