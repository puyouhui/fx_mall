package model

import (
	"database/sql"
	"strconv"
	"time"

	"go_backend/internal/database"
)

// DeliveryRecord 配送记录
type DeliveryRecord struct {
	ID                  int       `json:"id"`
	OrderID             int       `json:"order_id"`
	DeliveryEmployeeCode string   `json:"delivery_employee_code"`
	ProductImageURL     *string   `json:"product_image_url,omitempty"`
	DoorplateImageURL   *string   `json:"doorplate_image_url,omitempty"`
	CompletedAt         time.Time `json:"completed_at"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// CreateDeliveryRecord 创建配送记录
func CreateDeliveryRecord(record *DeliveryRecord) error {
	query := `
		INSERT INTO delivery_records (
			order_id, delivery_employee_code, product_image_url, doorplate_image_url, completed_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`
	
	result, err := database.DB.Exec(
		query,
		record.OrderID,
		record.DeliveryEmployeeCode,
		record.ProductImageURL,
		record.DoorplateImageURL,
		record.CompletedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	record.ID = int(id)
	return nil
}

// GetDeliveryRecordByOrderID 根据订单ID获取配送记录
func GetDeliveryRecordByOrderID(orderID int) (*DeliveryRecord, error) {
	query := `
		SELECT id, order_id, delivery_employee_code, product_image_url, doorplate_image_url, 
		       completed_at, created_at, updated_at
		FROM delivery_records
		WHERE order_id = ?
	`
	
	record := &DeliveryRecord{}
	var productImageURL, doorplateImageURL sql.NullString
	
	err := database.DB.QueryRow(query, orderID).Scan(
		&record.ID,
		&record.OrderID,
		&record.DeliveryEmployeeCode,
		&productImageURL,
		&doorplateImageURL,
		&record.CompletedAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if productImageURL.Valid {
		record.ProductImageURL = &productImageURL.String
	}
	if doorplateImageURL.Valid {
		record.DoorplateImageURL = &doorplateImageURL.String
	}

	return record, nil
}

// GetAllDeliveryRecords 获取所有配送记录（分页）
// 只显示配送员已接单的订单（基于delivery_logs表中的accepted操作）
func GetAllDeliveryRecords(pageNum, pageSize int, keyword string, startDate, endDate string) ([]*DeliveryRecord, int, error) {
	offset := (pageNum - 1) * pageSize
	
	// 构建查询条件 - 基于delivery_logs表，只查询有"accepted"操作的订单
	whereClause := ""
	args := []interface{}{}
	
	if keyword != "" {
		whereClause += " AND (o.id = ? OR o.order_number LIKE ? OR dl.delivery_employee_code LIKE ?)"
		// 尝试将关键词解析为订单ID
		if orderID, err := strconv.Atoi(keyword); err == nil {
			args = append(args, orderID, "%"+keyword+"%", "%"+keyword+"%")
		} else {
			args = append(args, 0, "%"+keyword+"%", "%"+keyword+"%")
		}
	}
	
	if startDate != "" {
		whereClause += " AND DATE(dl.accept_time) >= ?"
		args = append(args, startDate)
	}
	
	if endDate != "" {
		whereClause += " AND DATE(dl.accept_time) <= ?"
		args = append(args, endDate)
	}
	
	// 获取总数 - 基于接单日志，每个订单只统计一次
	countQuery := `
		SELECT COUNT(DISTINCT dl.order_id)
		FROM (
			SELECT order_id, delivery_employee_code, MIN(action_time) as accept_time
			FROM delivery_logs
			WHERE action = 'accepted'
			GROUP BY order_id, delivery_employee_code
		) dl
		INNER JOIN orders o ON dl.order_id = o.id
		WHERE 1=1` + whereClause
	var total int
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// 获取列表 - 基于接单日志，关联订单表和配送记录表
	// 使用子查询获取每个订单的接单时间（最早的accepted操作）
	query := `
		SELECT 
			COALESCE(dr.id, 0) as id,
			o.id as order_id,
			COALESCE(dl.delivery_employee_code, o.delivery_employee_code, '') as delivery_employee_code,
			dr.product_image_url,
			dr.doorplate_image_url,
			COALESCE(dr.completed_at, o.updated_at) as completed_at,
			COALESCE(dr.created_at, o.created_at) as created_at,
			COALESCE(dr.updated_at, o.updated_at) as updated_at
		FROM (
			SELECT order_id, delivery_employee_code, MIN(action_time) as accept_time
			FROM delivery_logs
			WHERE action = 'accepted'
			GROUP BY order_id, delivery_employee_code
		) dl
		INNER JOIN orders o ON dl.order_id = o.id
		LEFT JOIN delivery_records dr ON o.id = dr.order_id
		WHERE 1=1` + whereClause + `
		ORDER BY dl.accept_time DESC
		LIMIT ? OFFSET ?
	`
	
	args = append(args, pageSize, offset)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	records := []*DeliveryRecord{}
	for rows.Next() {
		record := &DeliveryRecord{}
		var productImageURL, doorplateImageURL sql.NullString
		var deliveryEmployeeCode string
		
		err := rows.Scan(
			&record.ID,
			&record.OrderID,
			&deliveryEmployeeCode,
			&productImageURL,
			&doorplateImageURL,
			&record.CompletedAt,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		record.DeliveryEmployeeCode = deliveryEmployeeCode
		if productImageURL.Valid {
			record.ProductImageURL = &productImageURL.String
		}
		if doorplateImageURL.Valid {
			record.DoorplateImageURL = &doorplateImageURL.String
		}
		
		records = append(records, record)
	}
	
	return records, total, nil
}

