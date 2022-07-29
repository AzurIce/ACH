package controllers

import (
	"ach/models"
	"ach/services/user"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminAddUser(c *gin.Context) {
	log.Println("[AdminAddUser]")
	var service user.UserAddService
	if err := c.BindJSON(&service); err == nil {
		c.JSON(service.Add(c))
	}
}

func AdminGetUserList(c *gin.Context) {
	log.Println("[AdminGetUserList]")
	if userList, err := models.GetUserList(); err == nil {
		c.JSON(http.StatusOK, userList)
	} else {
		c.Status(http.StatusInternalServerError)
	}
}