package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Frontend(fs http.FileSystem) gin.HandlerFunc {
	fileServer := http.FileServer(fs)
	return func(c *gin.Context) {
		// if bootstrap.Dev {
		// 	c.Next()
        //     return
		// }
		path := c.Request.URL.Path

		// API 跳过
		if strings.HasPrefix(path, "/api") {
			c.Next()
		} else {
			fileServer.ServeHTTP(c.Writer, c.Request)
		}
	}
}
