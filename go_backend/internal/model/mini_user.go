package model

import (
	"database/sql"
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

func nullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
