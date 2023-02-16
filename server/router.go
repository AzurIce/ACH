package server

import (
	"ach/internal/bootstrap"
	"ach/server/middlewares"
	"ach/server/service"

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
			user.POST("login", service.Handler(&service.UserLoginService{})) // POST api/user/login
			// TODO: user.POST("", controllers.UserRegister) // POST api/user
		}

		// Login required
		auth := api.Group("")
		auth.Use(middlewares.JWTAuth())
		{
			server := auth.Group("servers")
			{
				server.GET("", service.Handler(&service.GetServersService{})) // GET api/servers 获取服务器列表
				// TODO: server.GET(":name", controllers.GetServer) // GET api/servers/:name
				// TODO: server.POST(":id/start", middlewares.AdminCheck(), controllers.StartServer) //POST api/server/:name/start
				// TODO: server.POST(":id/stop", middlewares.AdminCheck(), controllers.StopServer) // POST api/server/:name/stop

				server.GET("console", service.Handler(&service.ServerConsoleService{})) // GET api/server/console
				// TODO: server.GET("log", middlewares.AdminCheck(), controllers.Log)         // GET api/server/log
			}

			admin := api.Group("admin")
			admin.Use(middlewares.AdminCheck())
			{
				user := admin.Group("user")
				{
					// TODO: user.GET("", services.Handler()) // GET  api/admin/user
					user.POST("", service.Handler(&service.UserUpdateService{})) // POST api/admin/user
					// TODO: user.POST("delete", controllers.UserRegister)
				}
			}
		}
	}

	return r
}
