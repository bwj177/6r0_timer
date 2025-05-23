package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xiaoxuxiansheng/xtimer/common/model/vo"
	"github.com/xiaoxuxiansheng/xtimer/pkg/jwt"
	"github.com/xiaoxuxiansheng/xtimer/pkg/log"
	"net/http"
	"strings"
)

// JWTAuthMiddleware 基于JWT的认证中间件
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URI
		// Authorization: Bearer xxxxxxx.xxx.xxx  / X-TOKEN: xxx.xxx.xx
		authHeader := c.Request.Header.Get("Authorization")
		log.Infof("token: %v", authHeader)
		if authHeader == "" {
			c.JSON(http.StatusForbidden, vo.NewCodeMsg(-1, fmt.Sprintf("not auth")))
			c.Abort()
			return
		}
		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusForbidden, vo.NewCodeMsg(-1, fmt.Sprintf("auth invalid all")))
			c.Abort()
			return
		}
		// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			log.Errorf("parse token failed,err:%v", err.Error())
			c.JSON(http.StatusForbidden, vo.NewCodeMsg(-1, fmt.Sprintf("auth invalid header")))
			c.Abort()
			return
		}
		// 将当前请求的userID信息保存到请求的上下文c上
		c.Set("userID", mc.UserID)
		c.Next() // 后续的处理请求的函数中 可以用过c.Get(CtxUserIDKey) 来获取当前请求的用户信息
	}
}
