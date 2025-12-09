package model

import (
	"database/sql"
	"time"

	"go_backend/internal/database"
)

// DeliveryLogAction 配送日志操作类型
const (
	DeliveryLogActionCreated          = "created"           // 订单创建
	DeliveryLogActionAccepted         = "accepted"          // 接单
	DeliveryLogActionPickupStarted    = "pickup_started"    // 开始取货
	DeliveryLogActionPickupCompleted = "pickup_completed"  // 取货完成
	DeliveryLogActionDeliveringStarted = "delivering_started" // 开始配送
	DeliveryLogActionDeliveringCompleted = "delivering_completed" // 配送完成
)

// DeliveryLog 配送流程日志
type DeliveryLog struct {
	ID                  int       `json:"id"`
	OrderID             int       `json:"order_id"`
	Action              string    `json:"action"`                // 操作类型
	DeliveryEmployeeCode *string  `json:"delivery_employee_code,omitempty"` // 配送员员工码
	ActionTime          time.Time `json:"action_time"`           // 操作时间
	Remark              *string   `json:"remark,omitempty"`     // 备注信息
	CreatedAt           time.Time `json:"created_at"`
}

// CreateDeliveryLog 创建配送流程日志
func CreateDeliveryLog(log *DeliveryLog) error {
	query := `
		INSERT INTO delivery_logs (
			order_id, action, delivery_employee_code, action_time, remark, created_at
		) VALUES (?, ?, ?, ?, ?, NOW())
	`
	
	_, err := database.DB.Exec(
		query,
		log.OrderID,
		log.Action,
		log.DeliveryEmployeeCode,
		log.ActionTime,
		log.Remark,
	)
	
	return err
}

// GetDeliveryLogsByOrderID 根据订单ID获取配送流程日志（按时间排序）
func GetDeliveryLogsByOrderID(orderID int) ([]DeliveryLog, error) {
	query := `
		SELECT id, order_id, action, delivery_employee_code, action_time, remark, created_at
		FROM delivery_logs
		WHERE order_id = ?
		ORDER BY action_time ASC, created_at ASC
	`
	
	rows, err := database.DB.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]DeliveryLog, 0)
	for rows.Next() {
		var log DeliveryLog
		var deliveryEmployeeCode, remark sql.NullString
		
		err := rows.Scan(
			&log.ID,
			&log.OrderID,
			&log.Action,
			&deliveryEmployeeCode,
			&log.ActionTime,
			&remark,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if deliveryEmployeeCode.Valid {
			log.DeliveryEmployeeCode = &deliveryEmployeeCode.String
		}
		if remark.Valid {
			log.Remark = &remark.String
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetDeliveryLogsByEmployeeCode 根据配送员员工码获取配送流程日志
func GetDeliveryLogsByEmployeeCode(employeeCode string, pageNum, pageSize int) ([]DeliveryLog, int, error) {
	offset := (pageNum - 1) * pageSize

	// 获取总数
	var total int
	countQuery := `
		SELECT COUNT(*) FROM delivery_logs
		WHERE delivery_employee_code = ?
	`
	err := database.DB.QueryRow(countQuery, employeeCode).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query := `
		SELECT id, order_id, action, delivery_employee_code, action_time, remark, created_at
		FROM delivery_logs
		WHERE delivery_employee_code = ?
		ORDER BY action_time DESC, created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := database.DB.Query(query, employeeCode, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := make([]DeliveryLog, 0)
	for rows.Next() {
		var log DeliveryLog
		var deliveryEmployeeCode, remark sql.NullString
		
		err := rows.Scan(
			&log.ID,
			&log.OrderID,
			&log.Action,
			&deliveryEmployeeCode,
			&log.ActionTime,
			&remark,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if deliveryEmployeeCode.Valid {
			log.DeliveryEmployeeCode = &deliveryEmployeeCode.String
		}
		if remark.Valid {
			log.Remark = &remark.String
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

