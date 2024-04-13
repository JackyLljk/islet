package middleware

import (
	"fmt"
	"islet/controller"
	"islet/pkg/jwt_auth"
	"strings"

	"github.com/gin-gonic/gin"
)

// token认证中间件

func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			controller.ResponseErrorWithMsg(c, controller.CodeInvalidToken, "请求缺少Auth Token")
			// 不调用该请求的剩余处理函数
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			controller.ResponseErrorWithMsg(c, controller.CodeInvalidToken, "Token 格式不正确")
			c.Abort()
			return
		}

		mc, err := jwt_auth.ParseToken(parts[1])
		if err != nil {
			fmt.Println(err)
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}

		c.Set(controller.ContextUserIDKey, mc.UserID)

		c.Next()
	}
}
