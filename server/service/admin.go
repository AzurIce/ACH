package service

import (
	"ach/internal/models"
	"ach/internal/utils"

	"github.com/gin-gonic/gin"
)

type GetUsersService struct{}

func (service *GetUsersService) Handle(c *gin.Context) (any, error) {
	return models.GetUserList()
}

type UserRegisterService struct{
	Code string `form:"code"`
}

func (service *UserRegisterService) Handle(c *gin.Context) (any, error) {
	playerInfo, err := utils.GetPlayerInfoByCode(service.Code)
	if err != nil {
		return nil, err
	}

	if _, err := models.GetUserByUUID(playerInfo.UUID); err != nil {
		_, err := models.CreateUser(playerInfo.UUID, playerInfo.Name)
		return nil, err
	} else {
		return nil, nil
	}
}