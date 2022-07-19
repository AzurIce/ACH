package utils

import (
	"log"
	"strings"
	// "time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type MyCustomClaims struct {
	UUID string `json:"UUID"`
	jwt.StandardClaims
}

func CreateToken(uuid string) (string, error) {

	t := jwt.New(jwt.GetSigningMethod("HS256"))

	t.Claims = &MyCustomClaims{
		uuid,
		jwt.StandardClaims{
			// ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
		},
	}

	return t.SignedString([]byte("azurcraft"))
}

func GetTokenStr(c *gin.Context) string {
	tokenStr := ""
	if c.Request.URL.Path == "/api/admin/server/console" {
		tokenStr = c.Query("token")
	} else {
		tokenStr = strings.ReplaceAll(c.Request.Header.Get("Authorization"), "Bearer ", "")
	}
	return tokenStr
}

func DecodeTokenStr(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("azurcraft"), nil
	})
	if err != nil {
		return token, err
	}
	return token, nil
}

func MustGetClaims(c *gin.Context) *MyCustomClaims {
	log.Println("[MustGetClaims]")
	tokenStr := GetTokenStr(c)
	log.Printf("[MustGetClaims] tokenStr: %s\n", tokenStr)
	token, _ := DecodeTokenStr(tokenStr)
	return token.Claims.(*MyCustomClaims)
}
