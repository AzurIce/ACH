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

	// TODO: Figure these things out
	config := cors.DefaultConfig()
	config.ExposeHeaders = []string{"Authorization"}
	config.AllowCredentials = true
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// FrontendFS
	r.Use(middlewares.Frontend(bootstrap.StaticFS))

	/*
		路由
	*/
	api := r.Group("api")
	{
		// No login required
		user := api.Group("user")
		{
			user.POST("login", controllers.UserLogin) // POST api/user/login
			// TODO: user.POST("reset", controllers.UserReset)
			user.GET("isAdmin", controllers.UserIsAdmin) // GET api/user/isAdmin
		}
		

		 // Login required
		auth := api.Group("")
		auth.Use(middlewares.JWTAuth())
		{
			auth.GET("server", controllers.GetServers)

			admin := api.Group("admin")
			admin.Use(middlewares.AdminCheck())
			{
				server := admin.Group("server")
				{
					server.GET("", controllers.GetServers)
					server.GET("log", controllers.Log)
					server.GET("console", controllers.Console) // GET api/admin/server
				}

				user := admin.Group("user")
				{
					user.GET("", controllers.AdminGetUserList) // POST api/admin/user
					user.POST("register", controllers.AdminAddUser)
					// TODO: user.POST("delete", controllers.UserRegister)
				}
			}
		}
	}

	return r
}