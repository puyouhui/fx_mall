package model

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go_backend/internal/database"
)

// MiniAppUser 表示小程序用户
type MiniAppUser struct {
	ID               int       `json:"id"`
	UniqueID         string    `json:"unique_id"`
	UserCode         string    `json:"user_code,omitempty"` // 用户编号（4-5位数）
	Name             string    `json:"name,omitempty"`       // 用户姓名
	Avatar           string    `json:"avatar,omitempty"`
	Phone            string    `json:"phone,omitempty"`
	SalesCode        string    `json:"sales_code,omitempty"`
	StoreType        string    `json:"store_type,omitempty"`
	UserType         string    `json:"user_type"`
	ProfileCompleted bool      `json:"profile_completed"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// GetMiniAppUserByUniqueID 根据唯一ID获取用户
func GetMiniAppUserByUniqueID(uniqueID string) (*MiniAppUser, error) {
	query := `
		SELECT id, unique_id, user_code, name, avatar, phone, sales_code, store_type, user_type, profile_completed, created_at, updated_at
		FROM mini_app_users
		WHERE unique_id = ?
		LIMIT 1`

	var (
		user                          MiniAppUser
		userCode, name, avatar, phone sql.NullString
		salesCode, storeType           sql.NullString
		profileCompleted               sql.NullInt64
	)

	err := database.DB.QueryRow(query, uniqueID).Scan(
		&user.ID,
		&user.UniqueID,
		&userCode,
		&name,
		&avatar,
		&phone,
		&salesCode,
		&storeType,
		&user.UserType,
		&profileCompleted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	user.UserCode = nullString(userCode)
	user.Name = nullString(name)
	user.Avatar = nullString(avatar)
	user.Phone = nullString(phone)
	user.SalesCode = nullString(salesCode)
	user.StoreType = nullString(storeType)
	user.ProfileCompleted = profileCompleted.Valid && profileCompleted.Int64 == 1

	return &user, nil
}

// GetMiniAppUserByID 根据ID获取用户详情（后台管理使用）
func GetMiniAppUserByID(id int) (*MiniAppUser, error) {
	query := `
		SELECT id, unique_id, user_code, name, avatar, phone, sales_code, store_type, user_type, profile_completed, created_at, updated_at
		FROM mini_app_users
		WHERE id = ?
		LIMIT 1`

	var (
		user                          MiniAppUser
		userCode, name, avatar, phone sql.NullString
		salesCode, storeType          sql.NullString
		profileCompleted              sql.NullInt64
	)

	err := database.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.UniqueID,
		&userCode,
		&name,
		&avatar,
		&phone,
		&salesCode,
		&storeType,
		&user.UserType,
		&profileCompleted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	user.UserCode = nullString(userCode)
	user.Name = nullString(name)
	user.Avatar = nullString(avatar)
	user.Phone = nullString(phone)
	user.SalesCode = nullString(salesCode)
	user.StoreType = nullString(storeType)
	user.ProfileCompleted = profileCompleted.Valid && profileCompleted.Int64 == 1

	return &user, nil
}

// GenerateUserCode 生成用户编号（4-5位数，优先4位数）
func GenerateUserCode() (string, error) {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())
	
	// 先尝试生成4位数（1000-9999）
	for i := 0; i < 100; i++ {
		code := fmt.Sprintf("%04d", 1000+rand.Intn(9000))
		var exists int
		err := database.DB.QueryRow(`
			SELECT COUNT(*) FROM mini_app_users WHERE user_code = ?
		`, code).Scan(&exists)
		if err != nil {
			return "", err
		}
		if exists == 0 {
			return code, nil
		}
	}
	
	// 如果4位数都用完了，使用5位数（10000-99999）
	for i := 0; i < 100; i++ {
		code := fmt.Sprintf("%05d", 10000+rand.Intn(90000))
		var exists int
		err := database.DB.QueryRow(`
			SELECT COUNT(*) FROM mini_app_users WHERE user_code = ?
		`, code).Scan(&exists)
		if err != nil {
			return "", err
		}
		if exists == 0 {
			return code, nil
		}
	}
	
	return "", fmt.Errorf("无法生成唯一的用户编号")
}

// CreateMiniAppUser 创建用户（仅记录唯一ID，其他信息后续完善）
func CreateMiniAppUser(uniqueID string) (*MiniAppUser, error) {
	// 检查用户是否已存在
	existingUser, err := GetMiniAppUserByUniqueID(uniqueID)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		// 如果用户已存在但没有编号，生成一个
		if existingUser.UserCode == "" {
			userCode, err := GenerateUserCode()
			if err != nil {
				return nil, err
			}
			_, err = database.DB.Exec(`
				UPDATE mini_app_users SET user_code = ?, updated_at = NOW() WHERE unique_id = ?
			`, userCode, uniqueID)
			if err != nil {
				return nil, err
			}
			existingUser.UserCode = userCode
		}
		return existingUser, nil
	}
	
	// 生成用户编号
	userCode, err := GenerateUserCode()
	if err != nil {
		return nil, err
	}
	
	_, err = database.DB.Exec(`
		INSERT INTO mini_app_users (unique_id, user_code, user_type, profile_completed, created_at, updated_at)
		VALUES (?, ?, 'unknown', 0, NOW(), NOW())
	`, uniqueID, userCode)
	if err != nil {
		return nil, err
	}

	return GetMiniAppUserByUniqueID(uniqueID)
}

// GetMiniAppUsers 获取用户列表
func GetMiniAppUsers(pageNum, pageSize int, keyword string) ([]MiniAppUser, int, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (pageNum - 1) * pageSize

	args := make([]interface{}, 0)
	where := ""
	if keyword != "" {
		like := "%" + keyword + "%"
		// 支持搜索：用户ID（数字匹配）、唯一ID、用户编号、姓名、电话、地址
		// 使用 EXISTS 子查询来搜索地址信息
		where = `WHERE (
			id = ? OR 
			unique_id LIKE ? OR 
			user_code LIKE ? OR 
			name LIKE ? OR 
			phone LIKE ? OR
			EXISTS (
				SELECT 1 FROM mini_app_addresses 
				WHERE user_id = mini_app_users.id 
				AND (name LIKE ? OR contact LIKE ? OR phone LIKE ? OR address LIKE ?)
			)
		)`
		// 尝试将关键词转换为数字（用于ID搜索）
		var idValue int
		if _, idErr := fmt.Sscanf(keyword, "%d", &idValue); idErr == nil {
			args = append(args, idValue)
		} else {
			args = append(args, 0) // 如果不是数字，使用0（不会匹配任何ID）
		}
		// 添加9个LIKE参数：unique_id, user_code, name, phone, address.name, address.contact, address.phone, address.address
		args = append(args, like, like, like, like, like, like, like, like)
	}

	countQuery := "SELECT COUNT(*) FROM mini_app_users " + where
	var total int
	if err := database.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, unique_id, user_code, name, avatar, phone, sales_code, store_type, user_type, profile_completed, created_at, updated_at
		FROM mini_app_users
	`
	if where != "" {
		query += where + " "
	}
	query += "ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]MiniAppUser, 0)
	for rows.Next() {
		var (
			user                          MiniAppUser
			userCode, name, avatar, phone sql.NullString
			salesCode, storeType          sql.NullString
			profileCompleted             sql.NullInt64
		)

		if err := rows.Scan(
			&user.ID,
			&user.UniqueID,
			&userCode,
			&name,
			&avatar,
			&phone,
			&salesCode,
			&storeType,
			&user.UserType,
			&profileCompleted,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		user.UserCode = nullString(userCode)
		user.Name = nullString(name)
		user.Avatar = nullString(avatar)
		user.Phone = nullString(phone)
		user.SalesCode = nullString(salesCode)
		user.StoreType = nullString(storeType)
		user.ProfileCompleted = profileCompleted.Valid && profileCompleted.Int64 == 1

		users = append(users, user)
	}

	return users, total, nil
}

// UpdateMiniAppUserType 更新用户的身份类型
func UpdateMiniAppUserType(uniqueID, userType string) error {
	_, err := database.DB.Exec(`
		UPDATE mini_app_users
		SET user_type = ?, updated_at = NOW()
		WHERE unique_id = ?
	`, userType, uniqueID)
	return err
}

// UpdateMiniAppUserAvatar 更新用户头像
func UpdateMiniAppUserAvatar(uniqueID, avatarURL string) error {
	_, err := database.DB.Exec(`
		UPDATE mini_app_users
		SET avatar = ?, updated_at = NOW()
		WHERE unique_id = ?
	`, avatarURL, uniqueID)
	return err
}

// UpdateMiniAppUserName 更新用户姓名
func UpdateMiniAppUserName(uniqueID, name string) error {
	_, err := database.DB.Exec(`
		UPDATE mini_app_users
		SET name = ?, updated_at = NOW()
		WHERE unique_id = ?
	`, name, uniqueID)
	return err
}

// UpdateMiniAppUserPhone 更新用户电话
func UpdateMiniAppUserPhone(uniqueID, phone string) error {
	_, err := database.DB.Exec(`
		UPDATE mini_app_users
		SET phone = ?, updated_at = NOW()
		WHERE unique_id = ?
	`, phone, uniqueID)
	return err
}

// UpdateMiniAppUserProfile 已废弃，现在使用地址表管理地址信息
// 此函数保留用于兼容，但不再更新用户表的地址相关字段
func UpdateMiniAppUserProfile(uniqueID string, profileData map[string]interface{}) error {
	// 此函数已不再使用，地址信息现在存储在地址表中
	// 保留函数签名以避免编译错误
	return nil
}

// UpdateMiniAppUserByAdmin 管理员更新用户信息（可修改所有字段包括用户类型）
func UpdateMiniAppUserByAdmin(id int, updateData map[string]interface{}) error {
	// 构建更新SQL
	updates := []string{}
	args := []interface{}{}

	if name, ok := updateData["name"].(string); ok {
		updates = append(updates, "name = ?")
		args = append(args, name)
	}
	if phone, ok := updateData["phone"].(string); ok {
		updates = append(updates, "phone = ?")
		args = append(args, phone)
	}
	if storeType, ok := updateData["storeType"].(string); ok {
		updates = append(updates, "store_type = ?")
		args = append(args, storeType)
	}
	if salesCode, ok := updateData["salesCode"].(string); ok {
		updates = append(updates, "sales_code = ?")
		args = append(args, salesCode)
	}
	if avatar, ok := updateData["avatar"].(string); ok {
		updates = append(updates, "avatar = ?")
		args = append(args, avatar)
	}
	// 管理员可以修改用户类型
	if userType, ok := updateData["userType"].(string); ok {
		updates = append(updates, "user_type = ?")
		args = append(args, userType)
	}
	// 管理员可以修改资料完善状态
	if profileCompleted, ok := updateData["profileCompleted"].(bool); ok {
		var completedValue int
		if profileCompleted {
			completedValue = 1
		} else {
			completedValue = 0
		}
		updates = append(updates, "profile_completed = ?")
		args = append(args, completedValue)
	}

	// 更新更新时间
	updates = append(updates, "updated_at = NOW()")

	if len(updates) == 0 {
		return nil
	}

	args = append(args, id)
	query := "UPDATE mini_app_users SET " + strings.Join(updates, ", ") + " WHERE id = ?"
	_, err := database.DB.Exec(query, args...)
	return err
}

func nullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
