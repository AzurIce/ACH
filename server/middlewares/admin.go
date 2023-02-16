package middlewares

import (
	"ach/internal/models"
	"ach/internal/jwt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminCheck() gin.HandlerFunc {
	return func(c *gin.Context) {

		log.Println("[middleware/AdminCheck]")
		claims := jwt.MustGetClaims(c)
		log.Println(claims)

		if user, err := models.GetUserByUUID(claims.UUID); err == nil && user.IsAdmin {
			c.Set("isAdmin", true)
			c.Next()
			return
		}

		c.Status(http.StatusForbidden)
		c.Abort()
	}
}
