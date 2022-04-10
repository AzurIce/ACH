package middlewares

import (
	"ach/models"
	"ach/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminCheck() gin.HandlerFunc {
	return func(c *gin.Context) {

		claims := utils.MustGetClaims(c)
		log.Println(claims)

		if user, err := models.GetUserByUUID(claims.UUID); err == nil && user.IsAdmin {
			c.Next()
			return
		}

		c.Status(http.StatusForbidden)
		c.Abort()
	}
}
