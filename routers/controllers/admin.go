package controllers

import (
	"ach/services/user"

	"github.com/gin-gonic/gin"
)

func AdminAddUser(c *gin.Context) {
	var service user.UserAddService
	if err := c.BindJSON(&service); err == nil {
		c.JSON(service.Add(c))
	}
}
