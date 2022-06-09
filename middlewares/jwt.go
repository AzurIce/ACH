package middlewares

import (
	"ach/lib/utils"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := utils.GetTokenStr(c)

		token, err := utils.DecodeTokenStr(tokenStr)
		// log.Println(token)

		if err != nil || !token.Valid {
			c.Abort()
			return
		}

		c.Next()
	}
}
