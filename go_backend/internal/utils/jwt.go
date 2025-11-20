package utils

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWT密钥
var jwtSecret = []byte("admin_jwt_secret_key") // 在实际生产环境中应该从配置文件中读取

// Claims 定义JWT的payload结构
type Claims struct {
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	jwt.StandardClaims
}

// GenerateToken 生成JWT token
func GenerateToken(username string, userID int) (string, error) {
	// 设置token过期时间
	nowTime := time.Now()
	expireTime := nowTime.Add(24 * time.Hour) // 24小时有效期

	claims := Claims{
		Username: username,
		UserID:   userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  nowTime.Unix(),
			Issuer:    "admin_console",
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)

	return token, err
}

// ParseToken 解析JWT token
func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}

// VerifyToken 验证token是否有效
func VerifyToken(token string) error {
	_, err := ParseToken(token)
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return errors.New("token is expired")
			}
		}
		return errors.New("invalid token")
	}
	return nil
}