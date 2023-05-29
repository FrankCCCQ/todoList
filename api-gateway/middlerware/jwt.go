package middlerware

import (
	"api-gateway/pkg/e"
	"api-gateway/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		code = 200
		token := c.GetHeader("Authorization")
		fmt.Println("token", token)
		if token == "" {
			code = 404
		} else {
			claim, err := util.ParseToken(token)
			if err != nil {
				code = e.ErrorAuthCheckTokenFail
			} else if time.Now().Unix() > claim.ExpiresAt {
				code = e.ErrorAuthCheckTokenTimeout
			}
		}
		if code != e.Success {
			c.JSON(e.Success, gin.H{
				"status": code,
				"msg":    e.GetMsg(uint(code)),
			})

			//这个请求 不会被后续的handler处理
			c.Abort()
			return
		}
		c.Next()
	}

}
