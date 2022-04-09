package middlewares

import (
	"ach/models"
	"ach/utils"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AdminCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		
		tokenStr := ""
		if c.Request.URL.Path == "/api/console" {
			// log.Print("dadsajdajdkasdlajdlasdkls")
			tokenStr = c.Query("token")
			// log.Println(tokenStr)
		} else {
			tokenStr = strings.ReplaceAll(c.Request.Header.Get("Authorization"), "Bearer ", "")
		}

		token, _ := jwt.ParseWithClaims(tokenStr, &utils.MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("azurcraft"), nil
		})
		log.Println(token.Claims.(*utils.MyCustomClaims))

		if user, err := models.GetUserByUUID(token.Claims.(*utils.MyCustomClaims).UUID); err == nil {
			if !user.IsAdmin {
				c.Status(http.StatusForbidden)
				c.Abort()
				return
			}
			c.Next()
			return
		}

		c.Status(http.StatusForbidden)
		c.Abort()
	}
}
