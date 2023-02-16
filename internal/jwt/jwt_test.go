package jwt

import (
	"ach/internal/bootstrap"
	"fmt"
	"testing"

	"github.com/golang-jwt/jwt/v4"
)

func TestCreateToken(t *testing.T) {
	token, err := CreateToken("Admin")
	fmt.Println(token, err)
}

func TestJwtPlayground(t *testing.T) {
	// jwt.New 使用指定签名方法创建 Payload（Claims）为空的 Token（签名方法置于 Header 部分）
	// 创建得到的 Token 的 Valid 值为 false，此值只有在 Token 被 Parse/Verify 时才会更新
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	fmt.Printf("%#v\n", token)

	// jwt.RegisteredClaims 是对 JMT 的 7 个保留 Claim 的包装
	// token.Claims = jwt.RegisteredClaims{}
	// fmt.Printf("%#v\n", token)

	// jwt.NewWithClaims 是上面两步的合并
	// token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), jwt.RegisteredClaims{})

	// SignedString 获得 签名后的 Token 的字符串
	//（对 base64 编码后的 Header 和 Payload 使用指定 Token 的 Header 中 alg 指定的算法加密并将结果置于 Signature 部分）
	tokenString, err := token.SignedString([]byte("test"))
	fmt.Println(tokenString, err)

	// Parse 使用由 KeyFunction 指定的 Key 对 tokenString 进行解析
	parsedToken, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(bootstrap.Config.JWTSigningString), nil
	})
	fmt.Printf("%#v %v\n", parsedToken, err)

	// t.Claims = &MyCustomClaims{
	// 	uuid,
	// 	jwt.StandardClaims{
	// 		// ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
	// 	},
	// }

	// return t.SignedString([]byte("azurcraft"))
}
