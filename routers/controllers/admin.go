package controllers

import (
	"ach/services/user"

	"github.com/gin-gonic/gin"
)

func AdminAddUser(c *gin.Context) {
	var service user.AddUserService
	if err := c.BindJSON(&service); err == nil {
		c.JSON(service.Register(c))
	}
}
