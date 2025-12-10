package model

import (
	"database/sql"
	"fmt"
	"go_backend/internal/database"
)

// DeliveryRouteOrder 配送路线排序记录
type DeliveryRouteOrder struct {
	ID                  int     `json:"id"`
	DeliveryEmployeeCode string `json:"delivery_employee_code"`
	OrderID             int     `json:"order_id"`
	RouteSequence       int     `json:"route_sequence"`
	CalculatedDistance  *float64 `json:"calculated_distance"`
	CalculatedAt        string  `json:"calculated_at"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

// UpdateRouteSequence 更新或插入配送员的订单排序
func UpdateRouteSequence(employeeCode string, orderSequences []struct {
	OrderID int
	Sequence int
	Distance *float64
}) error {
	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 先删除该配送员的所有现有排序记录
	_, err = tx.Exec(`
		DELETE FROM delivery_route_orders 
		WHERE delivery_employee_code = ?
	`, employeeCode)
	if err != nil {
		return fmt.Errorf("删除旧排序记录失败: %w", err)
	}

	// 插入新的排序记录
	stmt, err := tx.Prepare(`
		INSERT INTO delivery_route_orders 
		(delivery_employee_code, order_id, route_sequence, calculated_distance, calculated_at)
		VALUES (?, ?, ?, ?, NOW())
	`)
	if err != nil {
		return fmt.Errorf("准备插入语句失败: %w", err)
	}
	defer stmt.Close()

	for _, seq := range orderSequences {
		_, err = stmt.Exec(employeeCode, seq.OrderID, seq.Sequence, seq.Distance)
		if err != nil {
			return fmt.Errorf("插入排序记录失败: %w", err)
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// GetRouteOrdersByEmployee 获取配送员的排序后订单列表
func GetRouteOrdersByEmployee(employeeCode string) ([]DeliveryRouteOrder, error) {
	query := `
		SELECT id, delivery_employee_code, order_id, route_sequence, 
		       calculated_distance, calculated_at, created_at, updated_at
		FROM delivery_route_orders
		WHERE delivery_employee_code = ?
		ORDER BY route_sequence ASC
	`

	rows, err := database.DB.Query(query, employeeCode)
	if err != nil {
		return nil, fmt.Errorf("查询排序记录失败: %w", err)
	}
	defer rows.Close()

	var routeOrders []DeliveryRouteOrder
	for rows.Next() {
		var ro DeliveryRouteOrder
		var calculatedDistance sql.NullFloat64
		err := rows.Scan(
			&ro.ID,
			&ro.DeliveryEmployeeCode,
			&ro.OrderID,
			&ro.RouteSequence,
			&calculatedDistance,
			&ro.CalculatedAt,
			&ro.CreatedAt,
			&ro.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描排序记录失败: %w", err)
		}
		if calculatedDistance.Valid {
			ro.CalculatedDistance = &calculatedDistance.Float64
		}
		routeOrders = append(routeOrders, ro)
	}

	return routeOrders, nil
}

// DeleteRouteOrdersByEmployee 删除配送员的所有排序记录
func DeleteRouteOrdersByEmployee(employeeCode string) error {
	_, err := database.DB.Exec(`
		DELETE FROM delivery_route_orders 
		WHERE delivery_employee_code = ?
	`, employeeCode)
	if err != nil {
		return fmt.Errorf("删除排序记录失败: %w", err)
	}
	return nil
}

// GetOrderIDsByEmployee 获取配送员的所有订单ID（按排序）
func GetOrderIDsByEmployee(employeeCode string) ([]int, error) {
	query := `
		SELECT order_id
		FROM delivery_route_orders
		WHERE delivery_employee_code = ?
		ORDER BY route_sequence ASC
	`

	rows, err := database.DB.Query(query, employeeCode)
	if err != nil {
		return nil, fmt.Errorf("查询订单ID失败: %w", err)
	}
	defer rows.Close()

	var orderIDs []int
	for rows.Next() {
		var orderID int
		if err := rows.Scan(&orderID); err != nil {
			return nil, fmt.Errorf("扫描订单ID失败: %w", err)
		}
		orderIDs = append(orderIDs, orderID)
	}

	return orderIDs, nil
}

