package middlewares

import (
	"ach/lib/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := utils.GetTokenStr(c)
		log.Println("[middlewares/JWTAuth]: Token: ", tokenStr)

		token, err := utils.DecodeTokenStr(tokenStr)
		// log.Println(token)

		if err != nil || !token.Valid {
			log.Println("[middlewares/JWTAuth]: Token not valid")
			c.Abort()
			return
		}

		c.Next()
	}
}
