package routers

import (
	"ach/bootstrap"
	"ach/middlewares"
	"ach/routers/controllers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	config := cors.DefaultConfig()
	config.ExposeHeaders = []string{"Authorization"}
	config.AllowCredentials = true
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))
	r.Use(middlewares.Frontend(bootstrap.StaticFS))

	api := r.Group("/api")

	/*
		路由
	*/
	{
		user := api.Group("user")
		{
			user.POST("login", controllers.UserLogin)
			// TODO: user.POST("reset", controllers.UserReset)
		}

		auth := api.Group("")
		auth.Use(middlewares.JWTAuth())
		{
			admin := api.Group("admin")
			admin.Use(middlewares.AdminCheck())
			{
				server := admin.Group("server")
				{
					server.GET("console", controllers.Console)
				}

				user := admin.Group("user")
				{
					user.POST("", controllers.AdminAddUser)
					// TODO: user.POST("delete", controllers.UserRegister)
				}
			}
		}
	}

	return r
}