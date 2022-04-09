package middlewares

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// if c.Request.URL.Path == "/api/login" || c.Request.URL.Path == "/api/newlogin" {
		// 	c.Next()
		// 	return
		// }
		tokenStr := ""
		if c.Request.URL.Path == "/api/console" {
			// log.Print("dadsajdajdkasdlajdlasdkls")
			tokenStr = c.Query("token")
			// log.Println(tokenStr)
		} else {
			tokenStr = strings.ReplaceAll(c.Request.Header.Get("Authorization"), "Bearer ", "")
		}


		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("azurcraft"), nil
		})
		log.Println(token)

		if err != nil || !token.Valid {
			c.Abort()
			return
		}

		c.Next()
	}
}
