package model

import (
	"database/sql"
	"go_backend/internal/database"
)

// SystemSetting 系统设置
type SystemSetting struct {
	ID          int    `json:"id"`
	SettingKey  string `json:"setting_key"`
	SettingValue string `json:"setting_value"`
	Description string `json:"description"`
}

// GetSystemSetting 获取系统设置
func GetSystemSetting(key string) (string, error) {
	var value sql.NullString
	err := database.DB.QueryRow(`
		SELECT setting_value 
		FROM system_settings 
		WHERE setting_key = ?
	`, key).Scan(&value)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	
	if value.Valid {
		return value.String, nil
	}
	return "", nil
}

// SetSystemSetting 设置系统设置
func SetSystemSetting(key, value, description string) error {
	_, err := database.DB.Exec(`
		INSERT INTO system_settings (setting_key, setting_value, description)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			setting_value = VALUES(setting_value),
			description = VALUES(description),
			updated_at = NOW()
	`, key, value, description)
	return err
}

// GetAllSystemSettings 获取所有系统设置
func GetAllSystemSettings() (map[string]string, error) {
	rows, err := database.DB.Query(`
		SELECT setting_key, setting_value 
		FROM system_settings
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key string
		var value sql.NullString
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		if value.Valid {
			settings[key] = value.String
		} else {
			settings[key] = ""
		}
	}

	return settings, nil
}

// GetMapSettings 获取地图相关设置
func GetMapSettings() (map[string]string, error) {
	settings, err := GetAllSystemSettings()
	if err != nil {
		return nil, err
	}

	mapSettings := make(map[string]string)
	if amapKey, ok := settings["map_amap_key"]; ok {
		mapSettings["amap_key"] = amapKey
	}
	if tencentKey, ok := settings["map_tencent_key"]; ok {
		mapSettings["tencent_key"] = tencentKey
	}

	return mapSettings, nil
}

