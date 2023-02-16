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
	Username string `form:"username"`
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
			return nil, errors.New("invalid code")
		}
		// log.Println(playerInfo)

		// 数据库中无此用户(未授权)
		if user, err = models.GetUserByUUID(playerInfo.UUID); err == gorm.ErrRecordNotFound {
			return nil, errors.New("not Authenticated")
		}
	} else {
		if user, err = models.GetUserByUsername(service.Username); err == gorm.ErrRecordNotFound {
			return nil, errors.New("not exist")
		}

		if !user.CheckPassword(service.Password) {
			return nil, errors.New("incorrect password")
		}
	}

	var jwtToken string
	if user.Username == "Admin" {
		jwtToken, err = jwt.CreateToken("Admin")
		// fmt.Println(err)
	} else {
		jwtToken, err = jwt.CreateToken(user.PlayerUUID)
		// fmt.Println(err)
	}

	res := make(map[string]any)
	res["token"] = jwtToken
	res["user"] = user
	// res["user_name"] = user.Username

	return res, nil
}

type UserUpdateService struct {
	Username string `form:"username"` // 用户名
	Password string `form:"password"` // 密码
}

func (service *UserUpdateService) Handle(c *gin.Context) (any, error) {
	// playerInfo, err := utils.GetPlayerInfoByCode(service.Code)
	// if err != nil {
	// 	return nil, err
	// }

	// if _, err := models.GetUserByUUID(playerInfo.UUID); err != nil {
	// 	_, err := models.CreateUser(playerInfo.UUID, playerInfo.Name)
	// 	return nil, err
	// } else {
	// 	// TODO: 更新用户信息
	// 	return nil, nil
	// }
	return nil, nil
}

type GetUserService struct {
	UUID string `form:"uuid"`
}

func (service *GetUserService) Handle(c *gin.Context) (any, error) {
	if user, err := models.GetUserByUUID(service.UUID); err != nil {
		return user, nil
	} else {
		return nil, err
	}
}
