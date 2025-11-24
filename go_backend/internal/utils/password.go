package utils

import (
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用bcrypt加密密码
// cost参数控制加密强度，默认使用10（推荐值）
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash 验证密码是否匹配
// password: 用户输入的明文密码
// hash: 数据库中存储的加密密码
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// IsValidPhone 验证手机号格式（中国大陆手机号）
func IsValidPhone(phone string) bool {
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	return matched
}
