package model

import (
	"database/sql"
	"time"
)

// Admin 管理员结构体
type Admin struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应结构体
type LoginResponse struct {
	Token string `json:"token"`
	Admin Admin  `json:"admin"`
}

// GetAdminByUsername 根据用户名获取管理员
func GetAdminByUsername(db *sql.DB, username string) (*Admin, error) {
	var admin Admin
	query := "SELECT id, username, password, created_at, updated_at FROM admins WHERE username = ?"
	err := db.QueryRow(query, username).Scan(&admin.ID, &admin.Username, &admin.Password, &admin.CreatedAt, &admin.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

// GetAdminByUsernameAndPassword 根据用户名和密码获取管理员（已废弃，请使用GetAdminByUsername配合密码验证）
// 保留此函数以保持向后兼容，但不再使用
func GetAdminByUsernameAndPassword(db *sql.DB, username, password string) (*Admin, error) {
	var admin Admin
	query := "SELECT id, username, password, created_at, updated_at FROM admins WHERE username = ? AND password = ?"
	err := db.QueryRow(query, username, password).Scan(&admin.ID, &admin.Username, &admin.Password, &admin.CreatedAt, &admin.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

// GetAdminByID 根据ID获取管理员
func GetAdminByID(db *sql.DB, id int) (*Admin, error) {
	var admin Admin
	query := "SELECT id, username, password, created_at, updated_at FROM admins WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&admin.ID, &admin.Username, &admin.Password, &admin.CreatedAt, &admin.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

// UpdateAdminPassword 更新管理员密码
// id: 管理员ID
// hashedPassword: 已加密的密码（使用bcrypt等加密方法）
func UpdateAdminPassword(db *sql.DB, id int, hashedPassword string) error {
	query := "UPDATE admins SET password = ?, updated_at = NOW() WHERE id = ?"
	_, err := db.Exec(query, hashedPassword, id)
	return err
}
