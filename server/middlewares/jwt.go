package middlewares

import (
	"ach/internal/jwt"
	"log"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := jwt.GetTokenStr(c)
		log.Println("[middlewares/JWTAuth]: Token: ", tokenStr)

		token, err := jwt.DecodeTokenStr(tokenStr)
		// log.Println(token)

		if err != nil || !token.Valid {
			log.Println("[middlewares/JWTAuth]: Token not valid", err)
			c.Abort()
			return
		}

		c.Next()
	}
}
