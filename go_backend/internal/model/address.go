package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go_backend/internal/database"
	"go_backend/internal/utils"
)

// Address 表示用户地址
type Address struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Contact   string    `json:"contact"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Avatar    string    `json:"avatar,omitempty"`
	Latitude  *float64  `json:"latitude,omitempty"`
	Longitude *float64  `json:"longitude,omitempty"`
	StoreType string    `json:"store_type,omitempty"`
	SalesCode string    `json:"sales_code,omitempty"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateAddress 创建地址
func CreateAddress(userID int, addressData map[string]interface{}) (*Address, error) {
	// 如果设置为默认地址，先取消其他地址的默认状态
	isDefault := false
	if defaultVal, ok := addressData["is_default"].(bool); ok && defaultVal {
		isDefault = true
		// 取消该用户其他地址的默认状态
		_, err := database.DB.Exec(`
			UPDATE mini_app_addresses 
			SET is_default = 0 
			WHERE user_id = ?
		`, userID)
		if err != nil {
			return nil, err
		}
	}

	// 插入新地址（不再包含sales_code，销售员绑定到用户）
	query := `
		INSERT INTO mini_app_addresses (user_id, name, contact, phone, address, avatar, latitude, longitude, store_type, is_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	var latitude, longitude interface{}
	if lat, ok := addressData["latitude"].(float64); ok {
		latitude = lat
	}
	if lng, ok := addressData["longitude"].(float64); ok {
		longitude = lng
	}

	// 如果经纬度为空，尝试自动解析地址（兜底逻辑）
	if latitude == nil && longitude == nil {
		if addressStr, ok := addressData["address"].(string); ok && strings.TrimSpace(addressStr) != "" {
			// 获取地图API Key
			amapKey, _ := GetSystemSetting("map_amap_key")
			tencentKey, _ := GetSystemSetting("map_tencent_key")

			geocodeResult, err := utils.GeocodeAddress(strings.TrimSpace(addressStr), amapKey, tencentKey)
			if err == nil && geocodeResult.Success {
				latitude = geocodeResult.Latitude
				longitude = geocodeResult.Longitude
			}
			// 如果解析失败，不阻止保存，但记录日志
		}
	}

	result, err := database.DB.Exec(query,
		userID,
		addressData["name"],
		addressData["contact"],
		addressData["phone"],
		addressData["address"],
		addressData["avatar"],
		latitude,
		longitude,
		addressData["store_type"],
		isDefault,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	newAddress, err := GetAddressByID(int(id))
	if err != nil {
		return nil, err
	}

	// 如果设置为默认地址，且 store_type 不为空，则更新用户的 store_type
	if isDefault && newAddress != nil && newAddress.StoreType != "" {
		_, err = database.DB.Exec(`
			UPDATE mini_app_users 
			SET store_type = ?, updated_at = NOW()
			WHERE id = ?
		`, newAddress.StoreType, userID)
		if err != nil {
			// 记录错误但不影响地址创建
			_ = err
		}
	}

	return newAddress, nil
}

// GetAddressByID 根据ID获取地址
func GetAddressByID(id int) (*Address, error) {
	query := `
		SELECT id, user_id, name, contact, phone, address, avatar, latitude, longitude, store_type, sales_code, is_default, created_at, updated_at
		FROM mini_app_addresses
		WHERE id = ?
		LIMIT 1
	`

	var (
		address                      Address
		avatar, storeType, salesCode sql.NullString
		latitude, longitude          sql.NullFloat64
		isDefault                    sql.NullInt64
	)

	err := database.DB.QueryRow(query, id).Scan(
		&address.ID,
		&address.UserID,
		&address.Name,
		&address.Contact,
		&address.Phone,
		&address.Address,
		&avatar,
		&latitude,
		&longitude,
		&storeType,
		&salesCode,
		&isDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if latitude.Valid {
		val := latitude.Float64
		address.Latitude = &val
	}
	if longitude.Valid {
		val := longitude.Float64
		address.Longitude = &val
	}
	address.StoreType = nullString(storeType)
	address.SalesCode = nullString(salesCode)
	address.IsDefault = isDefault.Valid && isDefault.Int64 == 1

	return &address, nil
}

// GetAddressesByUserID 获取用户的所有地址
func GetAddressesByUserID(userID int) ([]Address, error) {
	query := `
		SELECT id, user_id, name, contact, phone, address, avatar, latitude, longitude, store_type, sales_code, is_default, created_at, updated_at
		FROM mini_app_addresses
		WHERE user_id = ?
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addresses := make([]Address, 0)
	for rows.Next() {
		var (
			address                      Address
			avatar, storeType, salesCode sql.NullString
			latitude, longitude          sql.NullFloat64
			isDefault                    sql.NullInt64
		)

		if err := rows.Scan(
			&address.ID,
			&address.UserID,
			&address.Name,
			&address.Contact,
			&address.Phone,
			&address.Address,
			&avatar,
			&latitude,
			&longitude,
			&storeType,
			&salesCode,
			&isDefault,
			&address.CreatedAt,
			&address.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if latitude.Valid {
			val := latitude.Float64
			address.Latitude = &val
		}
		if longitude.Valid {
			val := longitude.Float64
			address.Longitude = &val
		}
		address.Avatar = nullString(avatar)
		address.StoreType = nullString(storeType)
		address.SalesCode = nullString(salesCode)
		address.IsDefault = isDefault.Valid && isDefault.Int64 == 1

		addresses = append(addresses, address)
	}

	return addresses, nil
}

// GetDefaultAddressByUserID 获取用户的默认地址
func GetDefaultAddressByUserID(userID int) (*Address, error) {
	query := `
		SELECT id, user_id, name, contact, phone, address, avatar, latitude, longitude, store_type, sales_code, is_default, created_at, updated_at
		FROM mini_app_addresses
		WHERE user_id = ? AND is_default = 1
		LIMIT 1
	`

	var (
		address                      Address
		avatar, storeType, salesCode sql.NullString
		latitude, longitude          sql.NullFloat64
		isDefault                    sql.NullInt64
	)

	err := database.DB.QueryRow(query, userID).Scan(
		&address.ID,
		&address.UserID,
		&address.Name,
		&address.Contact,
		&address.Phone,
		&address.Address,
		&avatar,
		&latitude,
		&longitude,
		&storeType,
		&salesCode,
		&isDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if latitude.Valid {
		val := latitude.Float64
		address.Latitude = &val
	}
	if longitude.Valid {
		val := longitude.Float64
		address.Longitude = &val
	}
	address.Avatar = nullString(avatar)
	address.StoreType = nullString(storeType)
	address.SalesCode = nullString(salesCode)
	address.IsDefault = isDefault.Valid && isDefault.Int64 == 1

	return &address, nil
}

// UpdateAddress 更新地址
func UpdateAddress(id int, userID int, addressData map[string]interface{}) error {
	// 如果设置为默认地址，先取消其他地址的默认状态
	if defaultVal, ok := addressData["is_default"].(bool); ok && defaultVal {
		// 取消该用户其他地址的默认状态
		_, err := database.DB.Exec(`
			UPDATE mini_app_addresses 
			SET is_default = 0 
			WHERE user_id = ? AND id != ?
		`, userID, id)
		if err != nil {
			return err
		}
	}

	// 构建更新SQL
	updates := []string{}
	args := []interface{}{}

	if name, ok := addressData["name"].(string); ok && name != "" {
		updates = append(updates, "name = ?")
		args = append(args, name)
	}
	if contact, ok := addressData["contact"].(string); ok && contact != "" {
		updates = append(updates, "contact = ?")
		args = append(args, contact)
	}
	if phone, ok := addressData["phone"].(string); ok && phone != "" {
		updates = append(updates, "phone = ?")
		args = append(args, phone)
	}
	if address, ok := addressData["address"].(string); ok && address != "" {
		updates = append(updates, "address = ?")
		args = append(args, address)
	}
	if avatar, ok := addressData["avatar"].(string); ok {
		updates = append(updates, "avatar = ?")
		args = append(args, avatar)
	}
	if storeType, ok := addressData["store_type"].(string); ok {
		updates = append(updates, "store_type = ?")
		args = append(args, storeType)
	}
	// 不再更新sales_code，销售员绑定到用户而不是地址
	var hasLatitude, hasLongitude bool
	if latitude, ok := addressData["latitude"].(float64); ok {
		hasLatitude = true
		updates = append(updates, "latitude = ?")
		args = append(args, latitude)
	}
	if longitude, ok := addressData["longitude"].(float64); ok {
		hasLongitude = true
		updates = append(updates, "longitude = ?")
		args = append(args, longitude)
	}

	// 如果经纬度为空，尝试自动解析地址（兜底逻辑）
	if !hasLatitude && !hasLongitude {
		if addressStr, ok := addressData["address"].(string); ok && strings.TrimSpace(addressStr) != "" {
			// 获取地图API Key
			amapKey, _ := GetSystemSetting("map_amap_key")
			tencentKey, _ := GetSystemSetting("map_tencent_key")

			geocodeResult, err := utils.GeocodeAddress(strings.TrimSpace(addressStr), amapKey, tencentKey)
			if err == nil && geocodeResult.Success {
				updates = append(updates, "latitude = ?")
				args = append(args, geocodeResult.Latitude)
				updates = append(updates, "longitude = ?")
				args = append(args, geocodeResult.Longitude)
			}
			// 如果解析失败，不阻止保存，但记录日志
		}
	}
	if isDefault, ok := addressData["is_default"].(bool); ok {
		var defaultValue int
		if isDefault {
			defaultValue = 1
		} else {
			defaultValue = 0
		}
		updates = append(updates, "is_default = ?")
		args = append(args, defaultValue)
	}

	updates = append(updates, "updated_at = NOW()")

	if len(updates) == 0 {
		return nil
	}

	args = append(args, id, userID)
	query := "UPDATE mini_app_addresses SET " + strings.Join(updates, ", ") + " WHERE id = ? AND user_id = ?"
	_, err := database.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	// 如果设置为默认地址，获取更新后的地址信息，如果 store_type 不为空则更新用户的 store_type
	if defaultVal, ok := addressData["is_default"].(bool); ok && defaultVal {
		updatedAddress, err := GetAddressByID(id)
		if err == nil && updatedAddress != nil && updatedAddress.StoreType != "" {
			_, err = database.DB.Exec(`
				UPDATE mini_app_users 
				SET store_type = ?, updated_at = NOW()
				WHERE id = ?
			`, updatedAddress.StoreType, userID)
			if err != nil {
				// 记录错误但不影响地址更新
				_ = err
			}
		}
	}

	return nil
}

// DeleteAddress 删除地址
func DeleteAddress(id int, userID int) error {
	_, err := database.DB.Exec(`
		DELETE FROM mini_app_addresses 
		WHERE id = ? AND user_id = ?
	`, id, userID)
	return err
}

// SetDefaultAddress 设置默认地址
func SetDefaultAddress(id int, userID int) error {
	// 先取消该用户所有地址的默认状态
	_, err := database.DB.Exec(`
		UPDATE mini_app_addresses 
		SET is_default = 0 
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return err
	}

	// 设置指定地址为默认
	_, err = database.DB.Exec(`
		UPDATE mini_app_addresses 
		SET is_default = 1, updated_at = NOW()
		WHERE id = ? AND user_id = ?
	`, id, userID)
	if err != nil {
		return err
	}

	// 获取该地址的 store_type，如果不为空则更新用户的 store_type
	address, err := GetAddressByID(id)
	if err != nil {
		return err
	}
	if address != nil && address.StoreType != "" {
		_, err = database.DB.Exec(`
			UPDATE mini_app_users 
			SET store_type = ?, updated_at = NOW()
			WHERE id = ?
		`, address.StoreType, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

// CountAddressesByUserID 统计用户的地址数量
func CountAddressesByUserID(userID int) (int, error) {
	var count int
	err := database.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM mini_app_addresses 
		WHERE user_id = ?
	`, userID).Scan(&count)
	return count, err
}

// GetAddressesByIDs 批量获取地址信息
func GetAddressesByIDs(ids []int) (map[int]*Address, error) {
	if len(ids) == 0 {
		return make(map[int]*Address), nil
	}

	// 构建 IN 查询
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, name, contact, phone, address, avatar, latitude, longitude, store_type, sales_code, is_default, created_at, updated_at
		FROM mini_app_addresses
		WHERE id IN (%s)`, strings.Join(placeholders, ","))

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addresses := make(map[int]*Address)
	for rows.Next() {
		var (
			address                      Address
			avatar, storeType, salesCode sql.NullString
			latitude, longitude          sql.NullFloat64
			isDefault                    sql.NullInt64
		)

		err := rows.Scan(
			&address.ID,
			&address.UserID,
			&address.Name,
			&address.Contact,
			&address.Phone,
			&address.Address,
			&avatar,
			&latitude,
			&longitude,
			&storeType,
			&salesCode,
			&isDefault,
			&address.CreatedAt,
			&address.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if latitude.Valid {
			val := latitude.Float64
			address.Latitude = &val
		}
		if longitude.Valid {
			val := longitude.Float64
			address.Longitude = &val
		}
		address.StoreType = nullString(storeType)
		address.SalesCode = nullString(salesCode)
		address.IsDefault = isDefault.Valid && isDefault.Int64 == 1

		addresses[address.ID] = &address
	}

	return addresses, nil
}
