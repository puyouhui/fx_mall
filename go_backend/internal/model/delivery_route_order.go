package model

import (
	"database/sql"
	"fmt"
	"go_backend/internal/database"
	"time"
)

// DeliveryRouteOrder 配送路线排序记录
type DeliveryRouteOrder struct {
	ID                   int      `json:"id"`
	DeliveryEmployeeCode string   `json:"delivery_employee_code"`
	BatchID              string   `json:"batch_id"` // 批次ID（用于区分不同的趟）
	OrderID              int      `json:"order_id"`
	RouteSequence        int      `json:"route_sequence"`
	CalculatedDistance   *float64 `json:"calculated_distance"`
	CalculatedAt         string   `json:"calculated_at"`
	CreatedAt            string   `json:"created_at"`
	UpdatedAt            string   `json:"updated_at"`
}

// UpdateRouteSequence 更新或插入配送员的订单排序（使用批次ID）
func UpdateRouteSequence(employeeCode string, batchID string, orderSequences []struct {
	OrderID  int
	Sequence int
	Distance *float64
}) error {
	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 先删除该配送员当前批次的所有现有排序记录
	_, err = tx.Exec(`
		DELETE FROM delivery_route_orders 
		WHERE delivery_employee_code = ? AND batch_id = ?
	`, employeeCode, batchID)
	if err != nil {
		return fmt.Errorf("删除旧排序记录失败: %w", err)
	}

	// 插入新的排序记录
	stmt, err := tx.Prepare(`
		INSERT INTO delivery_route_orders 
		(delivery_employee_code, batch_id, order_id, route_sequence, calculated_distance, calculated_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`)
	if err != nil {
		return fmt.Errorf("准备插入语句失败: %w", err)
	}
	defer stmt.Close()

	for _, seq := range orderSequences {
		_, err = stmt.Exec(employeeCode, batchID, seq.OrderID, seq.Sequence, seq.Distance)
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

// GetRouteOrdersByEmployee 获取配送员的当前批次排序后订单列表
func GetRouteOrdersByEmployee(employeeCode string) ([]DeliveryRouteOrder, error) {
	fmt.Printf("[GetRouteOrdersByEmployee] 开始查询，配送员: %s\n", employeeCode)

	// 检查是否有 batch_id 字段
	var columnExists int
	checkColumnSQL := `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'delivery_route_orders' 
		AND COLUMN_NAME = 'batch_id'
	`
	hasBatchIDColumn := false
	if err := database.DB.QueryRow(checkColumnSQL).Scan(&columnExists); err == nil && columnExists > 0 {
		hasBatchIDColumn = true
		fmt.Printf("[GetRouteOrdersByEmployee] batch_id 字段存在\n")
	} else {
		fmt.Printf("[GetRouteOrdersByEmployee] batch_id 字段不存在，使用旧逻辑\n")
	}

	if !hasBatchIDColumn {
		// 没有 batch_id 字段，使用旧逻辑（返回所有订单，不区分批次）
		query := `
			SELECT id, delivery_employee_code, '' as batch_id, order_id, route_sequence, 
			       calculated_distance, calculated_at, created_at, updated_at
			FROM delivery_route_orders
			WHERE delivery_employee_code = ?
			ORDER BY route_sequence ASC
		`
		fmt.Printf("[GetRouteOrdersByEmployee] 执行旧逻辑查询\n")
		rows, err := database.DB.Query(query, employeeCode)
		if err != nil {
			fmt.Printf("[GetRouteOrdersByEmployee] 旧逻辑查询失败: %v\n", err)
			return nil, fmt.Errorf("查询排序记录失败: %w", err)
		}
		defer rows.Close()
		result, err := scanRouteOrders(rows)
		fmt.Printf("[GetRouteOrdersByEmployee] 旧逻辑查询成功，返回 %d 条记录\n", len(result))
		return result, err
	}

	// 有 batch_id 字段，使用批次逻辑
	// 获取当前最新的批次ID
	fmt.Printf("[GetRouteOrdersByEmployee] 获取当前批次ID\n")
	currentBatchID, err := GetCurrentBatchID(employeeCode)
	if err != nil {
		fmt.Printf("[GetRouteOrdersByEmployee] 获取当前批次ID失败: %v\n", err)
		return nil, fmt.Errorf("获取当前批次ID失败: %w", err)
	}
	fmt.Printf("[GetRouteOrdersByEmployee] 当前批次ID: %s\n", currentBatchID)
	if currentBatchID == "" {
		// 没有当前批次，返回空列表
		fmt.Printf("[GetRouteOrdersByEmployee] 没有当前批次，返回空列表\n")
		return []DeliveryRouteOrder{}, nil
	}

	// 查询当前批次的所有订单路线记录（包括已完成的）
	// 这样可以确保路线规划页面显示当前批次的所有订单
	query := `
		SELECT dro.id, dro.delivery_employee_code, dro.batch_id, dro.order_id, dro.route_sequence, 
		       dro.calculated_distance, dro.calculated_at, dro.created_at, dro.updated_at
		FROM delivery_route_orders dro
		INNER JOIN orders o ON dro.order_id = o.id
		WHERE dro.delivery_employee_code = ? 
		  AND dro.batch_id = ?
		ORDER BY dro.route_sequence ASC
	`

	fmt.Printf("[GetRouteOrdersByEmployee] 执行批次查询，批次ID: %s\n", currentBatchID)
	rows, err := database.DB.Query(query, employeeCode, currentBatchID)
	if err != nil {
		fmt.Printf("[GetRouteOrdersByEmployee] 批次查询失败: %v\n", err)
		return nil, fmt.Errorf("查询排序记录失败: %w", err)
	}
	defer rows.Close()

	result, err := scanRouteOrders(rows)
	if err != nil {
		fmt.Printf("[GetRouteOrdersByEmployee] 扫描记录失败: %v\n", err)
		return nil, err
	}
	fmt.Printf("[GetRouteOrdersByEmployee] 批次查询成功，返回 %d 条记录\n", len(result))
	return result, nil
}

// scanRouteOrders 扫描路线订单记录（辅助函数）
func scanRouteOrders(rows *sql.Rows) ([]DeliveryRouteOrder, error) {
	fmt.Printf("[scanRouteOrders] 开始扫描记录\n")
	var routeOrders []DeliveryRouteOrder
	count := 0
	for rows.Next() {
		count++
		var ro DeliveryRouteOrder
		var calculatedDistance sql.NullFloat64
		err := rows.Scan(
			&ro.ID,
			&ro.DeliveryEmployeeCode,
			&ro.BatchID,
			&ro.OrderID,
			&ro.RouteSequence,
			&calculatedDistance,
			&ro.CalculatedAt,
			&ro.CreatedAt,
			&ro.UpdatedAt,
		)
		if err != nil {
			fmt.Printf("[scanRouteOrders] 扫描第 %d 条记录失败: %v\n", count, err)
			return nil, fmt.Errorf("扫描排序记录失败: %w", err)
		}
		if calculatedDistance.Valid {
			ro.CalculatedDistance = &calculatedDistance.Float64
		}
		routeOrders = append(routeOrders, ro)
		fmt.Printf("[scanRouteOrders] 成功扫描第 %d 条记录，订单ID: %d\n", count, ro.OrderID)
	}

	// 检查是否有行扫描错误
	if err := rows.Err(); err != nil {
		fmt.Printf("[scanRouteOrders] 行扫描错误: %v\n", err)
		return nil, fmt.Errorf("行扫描错误: %w", err)
	}

	fmt.Printf("[scanRouteOrders] 扫描完成，共 %d 条记录\n", len(routeOrders))
	return routeOrders, nil
}

// DeleteRouteOrdersByEmployee 删除配送员当前批次的所有排序记录
func DeleteRouteOrdersByEmployee(employeeCode string) error {
	// 获取当前批次ID
	currentBatchID, err := GetCurrentBatchID(employeeCode)
	if err != nil {
		return fmt.Errorf("获取当前批次ID失败: %w", err)
	}
	if currentBatchID == "" {
		// 没有当前批次，无需删除
		return nil
	}

	_, err = database.DB.Exec(`
		DELETE FROM delivery_route_orders 
		WHERE delivery_employee_code = ? AND batch_id = ?
	`, employeeCode, currentBatchID)
	if err != nil {
		return fmt.Errorf("删除排序记录失败: %w", err)
	}
	return nil
}

// GetCurrentBatchID 获取配送员的当前批次ID（最新的未完成批次）
func GetCurrentBatchID(employeeCode string) (string, error) {
	fmt.Printf("[GetCurrentBatchID] 开始查询，配送员: %s\n", employeeCode)

	// 先检查是否有 batch_id 字段
	var columnExists int
	checkColumnSQL := `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'delivery_route_orders' 
		AND COLUMN_NAME = 'batch_id'
	`
	if err := database.DB.QueryRow(checkColumnSQL).Scan(&columnExists); err != nil {
		fmt.Printf("[GetCurrentBatchID] 检查字段失败: %v\n", err)
		// 检查失败，返回空字符串（使用旧逻辑）
		return "", nil
	}
	if columnExists == 0 {
		fmt.Printf("[GetCurrentBatchID] batch_id 字段不存在，返回空字符串\n")
		// 没有 batch_id 字段，返回空字符串（表示使用旧逻辑，不区分批次）
		return "", nil
	}

	fmt.Printf("[GetCurrentBatchID] batch_id 字段存在，开始查询批次\n")

	// 查询该配送员所有批次，找到最新的批次（包括已完成的）
	// 注意：使用 GROUP BY 而不是 DISTINCT，因为需要 ORDER BY created_at
	query := `
		SELECT dro.batch_id
		FROM delivery_route_orders dro
		WHERE dro.delivery_employee_code = ?
		GROUP BY dro.batch_id
		ORDER BY MIN(dro.created_at) DESC
		LIMIT 1
	`

	fmt.Printf("[GetCurrentBatchID] 执行查询: %s\n", query)
	var batchID string
	err := database.DB.QueryRow(query, employeeCode).Scan(&batchID)
	if err == sql.ErrNoRows {
		// 没有找到任何批次
		fmt.Printf("[GetCurrentBatchID] 没有找到批次记录\n")
		return "", nil
	}
	if err != nil {
		fmt.Printf("[GetCurrentBatchID] 查询失败: %v\n", err)
		return "", fmt.Errorf("查询当前批次ID失败: %w", err)
	}

	fmt.Printf("[GetCurrentBatchID] 找到批次ID: %s\n", batchID)

	// 检查该批次是否已完成（所有订单都是 delivered 或 shipped）
	checkQuery := `
		SELECT COUNT(*)
		FROM delivery_route_orders dro
		INNER JOIN orders o ON dro.order_id = o.id
		WHERE dro.delivery_employee_code = ? AND dro.batch_id = ?
		  AND o.status NOT IN ('delivered', 'shipped')
	`
	fmt.Printf("[GetCurrentBatchID] 检查批次状态\n")
	var incompleteCount int
	if err := database.DB.QueryRow(checkQuery, employeeCode, batchID).Scan(&incompleteCount); err != nil {
		fmt.Printf("[GetCurrentBatchID] 检查批次状态失败: %v\n", err)
		return "", fmt.Errorf("检查批次状态失败: %w", err)
	}

	fmt.Printf("[GetCurrentBatchID] 未完成订单数: %d\n", incompleteCount)

	// 如果所有订单都已完成，返回空字符串（表示没有当前批次，下次接单时会创建新批次）
	if incompleteCount == 0 {
		fmt.Printf("[GetCurrentBatchID] 所有订单已完成，返回空字符串\n")
		return "", nil
	}

	fmt.Printf("[GetCurrentBatchID] 返回批次ID: %s\n", batchID)
	return batchID, nil
}

// CreateNewBatch 为配送员创建新的批次ID
func CreateNewBatch(employeeCode string) (string, error) {
	// 生成批次ID：员工码_时间戳
	batchID := fmt.Sprintf("%s_%d", employeeCode, time.Now().Unix())
	return batchID, nil
}

// GetOrderIDsByEmployee 获取配送员的所有订单ID（按排序）
func GetOrderIDsByEmployee(employeeCode string) ([]int, error) {
	// 获取当前批次ID
	currentBatchID, err := GetCurrentBatchID(employeeCode)
	if err != nil {
		return nil, fmt.Errorf("获取当前批次ID失败: %w", err)
	}
	if currentBatchID == "" {
		// 没有当前批次，返回空列表
		return []int{}, nil
	}

	query := `
		SELECT order_id
		FROM delivery_route_orders
		WHERE delivery_employee_code = ? AND batch_id = ?
		ORDER BY route_sequence ASC
	`

	rows, err := database.DB.Query(query, employeeCode, currentBatchID)
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
