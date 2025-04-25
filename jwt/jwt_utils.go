package jwt

import (
	"fmt"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	gConfig JWTConfig
	once    sync.Once
)

type JWTConfig struct {
	SecretKey    string `yaml:"secret_key"`
	ExpireMinute int    `yaml:"expire_minute"`
}

func InitJWT(config JWTConfig) {
	once.Do(func() {
		gConfig = config
	})
}

func GenerateToken(claimMap map[string]any) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claimMap))
	expireTime := time.Now().Add(time.Minute * time.Duration(gConfig.ExpireMinute))
	token.Claims.(jwt.MapClaims)["exp"] = expireTime.Unix()
	return token.SignedString([]byte(gConfig.SecretKey))
}

func ValidateToken(token string) (map[string]any, error) {
	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(gConfig.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	return claims, nil
}
