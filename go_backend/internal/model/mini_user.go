package model

import (
	"database/sql"
	"strings"
	"time"

	"go_backend/internal/database"
)

// MiniAppUser 表示小程序用户
type MiniAppUser struct {
	ID               int       `json:"id"`
	UniqueID         string    `json:"unique_id"`
	Avatar           string    `json:"avatar,omitempty"`
	Name             string    `json:"name,omitempty"`
	Contact          string    `json:"contact,omitempty"`
	Phone            string    `json:"phone,omitempty"`
	Address          string    `json:"address,omitempty"`
	Latitude         *float64  `json:"latitude,omitempty"`
	Longitude        *float64  `json:"longitude,omitempty"`
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
		SELECT id, unique_id, avatar, name, contact, phone, address, latitude, longitude, sales_code, store_type, user_type, profile_completed, created_at, updated_at
		FROM mini_app_users
		WHERE unique_id = ?
		LIMIT 1`

	var (
		user                          MiniAppUser
		avatar, name, contact, phone  sql.NullString
		address, salesCode, storeType sql.NullString
		latitude, longitude           sql.NullFloat64
		profileCompleted              sql.NullInt64
	)

	err := database.DB.QueryRow(query, uniqueID).Scan(
		&user.ID,
		&user.UniqueID,
		&avatar,
		&name,
		&contact,
		&phone,
		&address,
		&latitude,
		&longitude,
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

	user.Avatar = nullString(avatar)
	user.Name = nullString(name)
	user.Contact = nullString(contact)
	user.Phone = nullString(phone)
	user.Address = nullString(address)
	user.SalesCode = nullString(salesCode)
	user.StoreType = nullString(storeType)
	if latitude.Valid {
		val := latitude.Float64
		user.Latitude = &val
	}
	if longitude.Valid {
		val := longitude.Float64
		user.Longitude = &val
	}
	user.ProfileCompleted = profileCompleted.Valid && profileCompleted.Int64 == 1

	return &user, nil
}

// GetMiniAppUserByID 根据ID获取用户详情（后台管理使用）
func GetMiniAppUserByID(id int) (*MiniAppUser, error) {
	query := `
		SELECT id, unique_id, avatar, name, contact, phone, address, latitude, longitude, sales_code, store_type, user_type, profile_completed, created_at, updated_at
		FROM mini_app_users
		WHERE id = ?
		LIMIT 1`

	var (
		user                          MiniAppUser
		avatar, name, contact, phone  sql.NullString
		address, salesCode, storeType sql.NullString
		latitude, longitude           sql.NullFloat64
		profileCompleted              sql.NullInt64
	)

	err := database.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.UniqueID,
		&avatar,
		&name,
		&contact,
		&phone,
		&address,
		&latitude,
		&longitude,
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

	user.Avatar = nullString(avatar)
	user.Name = nullString(name)
	user.Contact = nullString(contact)
	user.Phone = nullString(phone)
	user.Address = nullString(address)
	user.SalesCode = nullString(salesCode)
	user.StoreType = nullString(storeType)
	if latitude.Valid {
		val := latitude.Float64
		user.Latitude = &val
	}
	if longitude.Valid {
		val := longitude.Float64
		user.Longitude = &val
	}
	user.ProfileCompleted = profileCompleted.Valid && profileCompleted.Int64 == 1

	return &user, nil
}

// CreateMiniAppUser 创建用户（仅记录唯一ID，其他信息后续完善）
func CreateMiniAppUser(uniqueID string) (*MiniAppUser, error) {
	_, err := database.DB.Exec(`
		INSERT INTO mini_app_users (unique_id, user_type, profile_completed, created_at, updated_at)
		VALUES (?, 'unknown', 0, NOW(), NOW())
		ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at)
	`, uniqueID)
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
		where = "WHERE unique_id LIKE ? OR name LIKE ? OR phone LIKE ?"
		args = append(args, like, like, like)
	}

	countQuery := "SELECT COUNT(*) FROM mini_app_users " + where
	var total int
	if err := database.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, unique_id, avatar, name, contact, phone, address, latitude, longitude, sales_code, store_type, user_type, profile_completed, created_at, updated_at
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
			avatar, name, contact, phone  sql.NullString
			address, salesCode, storeType sql.NullString
			latitude, longitude           sql.NullFloat64
			profileCompleted              sql.NullInt64
		)

		if err := rows.Scan(
			&user.ID,
			&user.UniqueID,
			&avatar,
			&name,
			&contact,
			&phone,
			&address,
			&latitude,
			&longitude,
			&salesCode,
			&storeType,
			&user.UserType,
			&profileCompleted,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		user.Avatar = nullString(avatar)
		user.Name = nullString(name)
		user.Contact = nullString(contact)
		user.Phone = nullString(phone)
		user.Address = nullString(address)
		user.SalesCode = nullString(salesCode)
		user.StoreType = nullString(storeType)
		if latitude.Valid {
			val := latitude.Float64
			user.Latitude = &val
		}
		if longitude.Valid {
			val := longitude.Float64
			user.Longitude = &val
		}
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

// UpdateMiniAppUserProfile 更新用户资料，提交后自动设为零售身份并标记资料已完善
func UpdateMiniAppUserProfile(uniqueID string, profileData map[string]interface{}) error {
	// 构建更新SQL
	updates := []string{}
	args := []interface{}{}

	if name, ok := profileData["name"].(string); ok && name != "" {
		updates = append(updates, "name = ?")
		args = append(args, name)
	}
	if contact, ok := profileData["contact"].(string); ok && contact != "" {
		updates = append(updates, "contact = ?")
		args = append(args, contact)
	}
	if phone, ok := profileData["phone"].(string); ok && phone != "" {
		updates = append(updates, "phone = ?")
		args = append(args, phone)
	}
	if address, ok := profileData["address"].(string); ok && address != "" {
		updates = append(updates, "address = ?")
		args = append(args, address)
	}
	if storeType, ok := profileData["storeType"].(string); ok {
		updates = append(updates, "store_type = ?")
		args = append(args, storeType)
	}
	if salesCode, ok := profileData["salesCode"].(string); ok {
		updates = append(updates, "sales_code = ?")
		args = append(args, salesCode)
	}
	if latitude, ok := profileData["latitude"].(float64); ok {
		updates = append(updates, "latitude = ?")
		args = append(args, latitude)
	}
	if longitude, ok := profileData["longitude"].(float64); ok {
		updates = append(updates, "longitude = ?")
		args = append(args, longitude)
	}

	// 提交资料后，自动设置为零售身份，并标记资料已完善
	updates = append(updates, "user_type = ?")
	args = append(args, "retail")
	updates = append(updates, "profile_completed = ?")
	args = append(args, 1)
	updates = append(updates, "updated_at = NOW()")

	if len(updates) == 0 {
		return nil
	}

	args = append(args, uniqueID)
	query := "UPDATE mini_app_users SET " + strings.Join(updates, ", ") + " WHERE unique_id = ?"
	_, err := database.DB.Exec(query, args...)
	return err
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
	if contact, ok := updateData["contact"].(string); ok {
		updates = append(updates, "contact = ?")
		args = append(args, contact)
	}
	if phone, ok := updateData["phone"].(string); ok {
		updates = append(updates, "phone = ?")
		args = append(args, phone)
	}
	if address, ok := updateData["address"].(string); ok {
		updates = append(updates, "address = ?")
		args = append(args, address)
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
	if latitude, ok := updateData["latitude"].(float64); ok {
		updates = append(updates, "latitude = ?")
		args = append(args, latitude)
	}
	if longitude, ok := updateData["longitude"].(float64); ok {
		updates = append(updates, "longitude = ?")
		args = append(args, longitude)
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
