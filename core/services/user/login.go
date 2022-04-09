package user

import (
	"time"

	"ach/core/models"
	"ach/utils"

	"github.com/golang-jwt/jwt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserLoginService struct {
	Code string `form:"code"`
}

func (service *UserLoginService) Login(c *gin.Context) (string) {
	// 获取账号对应 UUID
	// 1. Code -> Token
	token := utils.GetToken(service.Code)
	// 2. Token -> XBL Token
	xblRes := utils.GetXBLToken(token)
	// 3. XBL Token -> XSTS Token
	xstsRes := utils.GetXSTSToken(xblRes.Token)
	// 4. XSTS Token + UHS -> MC Token
	mcToken := utils.GetMCToken(xstsRes)
	// 5. MC Token -> Player Info
	playerInfo := utils.GetPlayerInfo(mcToken)
	// log.Println(playerInfo)

	// 若数据库中无此用户则为其创建
	if _, err := models.GetUserByUUID(playerInfo.UUID); err == gorm.ErrRecordNotFound {
		models.CreateUser(playerInfo.UUID, playerInfo.Name)
	}

	jwtToken, _ := createToken(playerInfo.UUID)
	return jwtToken
}

type MyCustomClaims struct {
	UUID string `json:"UUID"`
	jwt.StandardClaims
}

func createToken(uuid string) (string, error) {

	t := jwt.New(jwt.GetSigningMethod("HS256"))

	t.Claims = &MyCustomClaims{
		uuid,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
		},
	}

	return t.SignedString([]byte("azurcraft"))
}

