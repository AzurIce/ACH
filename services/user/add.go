package user

import (
	"ach/models"
	"ach/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserAddService struct {
	Code string `form:"code"`
}

func (service *UserAddService) Add(c *gin.Context) (int, string) {
	println("添加用户 code: " + service.Code)
	playerInfo, err := utils.GetPlayerInfoByCode(service.Code)
	if err != nil {
		return http.StatusBadRequest, `{"error":"Invalid code"}`
	}

	if _, err := models.GetUserByUUID(playerInfo.UUID); err == nil {
		return http.StatusConflict, `{"error":"User already exist"}`
	} else {
		models.CreateUser(playerInfo.UUID, playerInfo.Name)
		return http.StatusOK, ""
	}

}
