package service

import (
	"ach/internal/jwt"
	"ach/internal/models"
	"ach/internal/utils"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserLoginService struct {
	Code     string `form:"code"`
	Name     string `form:"name"`
	Password string `form:"password"`
}

func (service *UserLoginService) Handle(c *gin.Context) (any, error) {
	var user models.User
	var err error
	// 提供了 Code，使用微软登录
	if service.Code != "" {
		// 获取游戏账号信息
		playerInfo, err := utils.GetPlayerInfoByCode(service.Code)
		if err != nil {
			return nil, errors.New("Invalid code")
		}
		// log.Println(playerInfo)

		// 数据库中无此用户(未授权)
		if user, err = models.GetUserByUUID(playerInfo.UUID); err == gorm.ErrRecordNotFound {
			return nil, errors.New("Not Authenticated")
		}
	} else {
		if user, err = models.GetUserByName(service.Name); err == gorm.ErrRecordNotFound {
			return nil, errors.New("Not exist")
		}

		if !user.CheckPassword(service.Password) {
			return nil, errors.New("Incorrect password")
		}
	}

	var jwtToken string
	if user.Name == "Admin" {
		jwtToken, _ = jwt.CreateToken("Admin")
	} else {
		jwtToken, _ = jwt.CreateToken(user.PlayerUUID)
	}

	res := make(map[string]any)
	res["token"] = jwtToken
	res["user_name"] = user.Name

	return res, nil
}

type UserUpdateService struct {
	Code string `form:"code"`
}

func (service *UserUpdateService) Handle(c *gin.Context) (any, error) {
	playerInfo, err := utils.GetPlayerInfoByCode(service.Code)
	if err != nil {
		return nil, err
	}

	if _, err := models.GetUserByUUID(playerInfo.UUID); err != nil {
		_, err := models.CreateUser(playerInfo.UUID, playerInfo.Name)
		return nil, err
	} else {
		// TODO: 更新用户信息
		return nil, nil
	}
}
