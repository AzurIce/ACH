package user

import (
	"ach/models"
	"ach/pkg/jwt"
	"ach/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserLoginService struct {
	Code     string `form:"code"`
	Name     string `form:"name"`
	Password string `form:"password"`
}

func (service *UserLoginService) Login(c *gin.Context) (int, string) {
	var user models.User
	var err error
	// 提供了 Code，使用微软登录
	if service.Code != "" {
		// 获取游戏账号信息
		playerInfo, err := utils.GetPlayerInfoByCode(service.Code)
		if err != nil {
			return http.StatusBadRequest, `{"error":"Invalid code"}`
		}
		// log.Println(playerInfo)

		// 数据库中无此用户(未授权)
		if user, err = models.GetUserByUUID(playerInfo.UUID); err == gorm.ErrRecordNotFound {
			return http.StatusUnauthorized, `{"error":"Not Authenticated"}`
		}
	} else {
		if user, err = models.GetUserByName(service.Name); err == gorm.ErrRecordNotFound {
			return http.StatusUnauthorized, `{"error":"Not exist"}`
		}

		if !user.CheckPassword(service.Password) {
			return http.StatusUnauthorized, `{"error":"Incorrect password"}`
		}
	}

	var jwtToken string
	if user.Name == "Admin" {
		jwtToken, _ = jwt.CreateToken("Admin")
	} else {
		jwtToken, _ = jwt.CreateToken(user.PlayerUUID)
	}
	return http.StatusOK, `{"token": "` + jwtToken + `", "user_name": "` + user.Name + `"}`
}
