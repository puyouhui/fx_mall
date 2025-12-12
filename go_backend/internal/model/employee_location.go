package model

import (
	"database/sql"
	"time"

	"go_backend/internal/database"
)

// EmployeeLocationHistory 配送员位置历史
type EmployeeLocationHistory struct {
	ID           int       `json:"id"`
	EmployeeID   int       `json:"employee_id"`
	EmployeeCode string    `json:"employee_code"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Accuracy     *float64  `json:"accuracy,omitempty"` // 精度（米），可为空
	CreatedAt    time.Time `json:"created_at"`
}

// SaveEmployeeLocation 保存配送员位置到数据库
func SaveEmployeeLocation(employeeID int, employeeCode string, latitude, longitude, accuracy float64) error {
	var accuracyValue sql.NullFloat64
	if accuracy > 0 {
		accuracyValue = sql.NullFloat64{
			Float64: accuracy,
			Valid:   true,
		}
	}

	_, err := database.DB.Exec(`
		INSERT INTO employee_location_history (employee_id, employee_code, latitude, longitude, accuracy, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`, employeeID, employeeCode, latitude, longitude, accuracyValue)
	return err
}

// GetLatestEmployeeLocation 获取配送员最新位置（优先从内存，其次从数据库）
func GetLatestEmployeeLocation(employeeID int) (*EmployeeLocationHistory, error) {
	var location EmployeeLocationHistory
	var accuracy sql.NullFloat64

	err := database.DB.QueryRow(`
		SELECT id, employee_id, employee_code, latitude, longitude, accuracy, created_at
		FROM employee_location_history
		WHERE employee_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, employeeID).Scan(
		&location.ID,
		&location.EmployeeID,
		&location.EmployeeCode,
		&location.Latitude,
		&location.Longitude,
		&accuracy,
		&location.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有找到位置记录
		}
		return nil, err
	}

	if accuracy.Valid {
		location.Accuracy = &accuracy.Float64
	}

	return &location, nil
}

// GetLatestEmployeeLocationByCode 根据员工码获取配送员最新位置
func GetLatestEmployeeLocationByCode(employeeCode string) (*EmployeeLocationHistory, error) {
	var location EmployeeLocationHistory
	var accuracy sql.NullFloat64

	err := database.DB.QueryRow(`
		SELECT id, employee_id, employee_code, latitude, longitude, accuracy, created_at
		FROM employee_location_history
		WHERE employee_code = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, employeeCode).Scan(
		&location.ID,
		&location.EmployeeID,
		&location.EmployeeCode,
		&location.Latitude,
		&location.Longitude,
		&accuracy,
		&location.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有找到位置记录
		}
		return nil, err
	}

	if accuracy.Valid {
		location.Accuracy = &accuracy.Float64
	}

	return &location, nil
}

// CleanOldLocationHistory 清理旧的位置记录（保留最近N天的数据）
func CleanOldLocationHistory(days int) error {
	_, err := database.DB.Exec(`
		DELETE FROM employee_location_history
		WHERE created_at < DATE_SUB(NOW(), INTERVAL ? DAY)
	`, days)
	return err
}

