package models

import (
	"time"
)

// Admin 管理员模型
type Admin struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password,omitempty"` // 密码在响应中不显示
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
	Admin Admin  `json:"admin"`
}