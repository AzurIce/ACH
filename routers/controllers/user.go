package controllers

import (
	"ach/models"
	"ach/services/user"
	"ach/utils"
	"net/http"

	// "io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
)

func UserLogin(c *gin.Context) {
	// body, _ := ioutil.ReadAll(c.Request.Body)

	var service user.LoginService
	// log.Print("[UserLogin]", string(body))
	if err := c.BindJSON(&service); err == nil {
		log.Println("[UserLogin] BindJSON Succesed")
		c.JSON(service.Login(c))
	} else {
		log.Printf("[UserLogin] BindJSON Failed, %s", err)
	}
}

func UserIsAdmin(c *gin.Context) {
	claims := utils.MustGetClaims(c)

	log.Print("[UserIsAdmin] claims:", claims, '\n')
	if user, err := models.GetUserByUUID(claims.UUID); err == nil {
		if !user.IsAdmin {
			log.Println("[UserIsAdmin]: false")
			c.Status(http.StatusForbidden)
		} else {
			log.Println("[UserIsAdmin]: true")
			c.Status(http.StatusOK)
		}
	}
}
