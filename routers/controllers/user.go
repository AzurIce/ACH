package controllers

import (
	"ach/services/user"

	"github.com/gin-gonic/gin"
)

func UserLogin(c *gin.Context) {
	var service user.LoginService
	if err := c.BindJSON(&service); err == nil {
		c.JSON(service.Login(c))
	}
}