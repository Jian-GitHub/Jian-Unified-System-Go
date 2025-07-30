package util

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// GenToken 生成JWT
func GenToken(secretKey string, seconds int64, payloads map[string]interface{}) (string, error) {
	iat := time.Now().UTC().Unix()
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	for k, v := range payloads {
		claims[k] = v
	}
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}
