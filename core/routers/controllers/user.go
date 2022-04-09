package controllers

import (
	"ach/core/services/user"

	"github.com/gin-gonic/gin"
)

func UserLogin(c *gin.Context) {
	var service user.UserLoginService
	if err := c.BindJSON(&service); err == nil {
		res := service.Login(c)
		c.JSON(200, res)
	}
}