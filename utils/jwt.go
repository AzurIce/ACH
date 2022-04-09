package utils

import (
	"time"

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
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
		},
	}

	return t.SignedString([]byte("azurcraft"))
}