package utils

import (
	"errors"
	"fmt"
	"perfect-pic-server/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

//var jwtSecret = []byte(config.Get().JWT.Secret)

// JWTClaims 自定义的 JWT Claims 结构体（单管理员模式）
type JWTClaims struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	jwt.RegisteredClaims
}

func getSecret() []byte {
	return []byte(config.Get().JWT.Secret)
}

func GenerateToken(id uint, username string, admin bool, duration time.Duration) (string, error) {
	claims := JWTClaims{
		ID:       id,
		Username: username,
		Admin:    admin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			Issuer:    "perfect-pic-server",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecret())
}

func ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法是否预期
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
