package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go_backend/internal/database"
)

// SalesCommissionConfig 销售分成配置
type SalesCommissionConfig struct {
	ID                    int       `json:"id"`
	EmployeeCode          string    `json:"employee_code"`
	BaseCommissionRate   float64   `json:"base_commission_rate"`   // 基础提成比例（默认45%）
	NewCustomerBonusRate  float64   `json:"new_customer_bonus_rate"` // 新客开发激励比例（默认20%）
	Tier1Threshold        float64   `json:"tier1_threshold"`       // 阶梯1阈值（默认50000元）
	Tier1Rate             float64   `json:"tier1_rate"`            // 阶梯1提成比例（默认5%）
	Tier2Threshold        float64   `json:"tier2_threshold"`       // 阶梯2阈值（默认100000元）
	Tier2Rate             float64   `json:"tier2_rate"`            // 阶梯2提成比例（默认10%）
	Tier3Threshold        float64   `json:"tier3_threshold"`       // 阶梯3阈值（默认200000元）
	Tier3Rate             float64   `json:"tier3_rate"`            // 阶梯3提成比例（默认20%）
	MinProfitThreshold    float64   `json:"min_profit_threshold"`   // 最小利润阈值（默认5元）
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// SalesCommission 销售分成记录
type SalesCommission struct {
	ID                   int        `json:"id"`
	OrderID              int         `json:"order_id"`
	EmployeeCode         string      `json:"employee_code"`
	UserID               int         `json:"user_id"`
	OrderNumber          string      `json:"order_number"`
	OrderDate            time.Time   `json:"order_date"`
	SettlementDate       *time.Time  `json:"settlement_date,omitempty"`
	IsValidOrder         bool        `json:"is_valid_order"`
	IsNewCustomerOrder   bool        `json:"is_new_customer_order"`
	OrderAmount          float64     `json:"order_amount"`          // 平台总收入
	GoodsCost            float64     `json:"goods_cost"`            // 商品总成本
	DeliveryCost         float64     `json:"delivery_cost"`         // 配送成本
	OrderProfit          float64     `json:"order_profit"`          // 订单利润
	BaseCommission       float64     `json:"base_commission"`       // 基础提成
	NewCustomerBonus     float64     `json:"new_customer_bonus"`    // 新客开发激励
	TierCommission       float64     `json:"tier_commission"`       // 阶梯提成
	TotalCommission      float64     `json:"total_commission"`     // 总分成
	TierLevel            int         `json:"tier_level"`            // 达到的阶梯等级
	CalculationMonth     string      `json:"calculation_month"`     // 计算月份（YYYY-MM）
	IsAccounted          bool        `json:"is_accounted"`          // 是否已计入（平台承认了销售员这个分润收入）
	AccountedAt          *time.Time  `json:"accounted_at,omitempty"` // 计入时间
	IsSettled            bool        `json:"is_settled"`            // 是否已结算（平台已经将该费用结算给销售员）
	SettledAt            *time.Time  `json:"settled_at,omitempty"`   // 结算时间
	IsAccountedCancelled bool        `json:"is_accounted_cancelled"` // 计入是否已取消
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
}

// SalesCommissionMonthlyStats 销售分成月统计
type SalesCommissionMonthlyStats struct {
	ID                   int       `json:"id"`
	EmployeeCode         string    `json:"employee_code"`
	StatMonth           string    `json:"stat_month"`            // YYYY-MM格式
	TotalSalesAmount     float64   `json:"total_sales_amount"`  // 总销售额
	TotalValidOrders     int       `json:"total_valid_orders"`  // 有效订单数
	TotalNewCustomers    int       `json:"total_new_customers"` // 新客户数
	TotalProfit          float64   `json:"total_profit"`        // 总利润
	TotalBaseCommission  float64   `json:"total_base_commission"`  // 总基础提成
	TotalNewCustomerBonus float64  `json:"total_new_customer_bonus"` // 总新客激励
	TotalTierCommission  float64   `json:"total_tier_commission"`  // 总阶梯提成
	TotalCommission      float64   `json:"total_commission"`       // 总分成
	TierLevel            int       `json:"tier_level"`              // 达到的阶梯等级
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// CommissionCalculationResult 分成计算结果
type CommissionCalculationResult struct {
	OrderProfit          float64 `json:"order_profit"`          // 订单利润
	BaseCommission       float64 `json:"base_commission"`       // 基础提成
	NewCustomerBonus     float64 `json:"new_customer_bonus"`    // 新客开发激励
	TierCommission       float64 `json:"tier_commission"`       // 阶梯提成
	TotalCommission      float64 `json:"total_commission"`      // 总分成
	TierLevel            int     `json:"tier_level"`            // 达到的阶梯等级
	IsValidOrder         bool    `json:"is_valid_order"`        // 是否有效订单
	IsNewCustomerOrder   bool    `json:"is_new_customer_order"` // 是否新客户首单
}

// GetSalesCommissionConfig 获取销售员的分成配置（如果不存在则创建默认配置）
func GetSalesCommissionConfig(employeeCode string) (*SalesCommissionConfig, error) {
	var config SalesCommissionConfig
	query := `
		SELECT id, employee_code, base_commission_rate, new_customer_bonus_rate,
		       tier1_threshold, tier1_rate, tier2_threshold, tier2_rate,
		       tier3_threshold, tier3_rate, min_profit_threshold,
		       created_at, updated_at
		FROM sales_commission_config
		WHERE employee_code = ?
	`
	err := database.DB.QueryRow(query, employeeCode).Scan(
		&config.ID, &config.EmployeeCode, &config.BaseCommissionRate, &config.NewCustomerBonusRate,
		&config.Tier1Threshold, &config.Tier1Rate, &config.Tier2Threshold, &config.Tier2Rate,
		&config.Tier3Threshold, &config.Tier3Rate, &config.MinProfitThreshold,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// 不存在则创建默认配置
			return CreateDefaultSalesCommissionConfig(employeeCode)
		}
		return nil, err
	}
	return &config, nil
}

// CreateDefaultSalesCommissionConfig 创建默认的分成配置
func CreateDefaultSalesCommissionConfig(employeeCode string) (*SalesCommissionConfig, error) {
	config := &SalesCommissionConfig{
		EmployeeCode:         employeeCode,
		BaseCommissionRate:  0.45,  // 45%
		NewCustomerBonusRate: 0.20, // 20%
		Tier1Threshold:       50000.00,
		Tier1Rate:           0.05,  // 5%
		Tier2Threshold:       100000.00,
		Tier2Rate:           0.10,  // 10%
		Tier3Threshold:       200000.00,
		Tier3Rate:           0.20,  // 20%
		MinProfitThreshold:   5.00,
	}

	query := `
		INSERT INTO sales_commission_config (
			employee_code, base_commission_rate, new_customer_bonus_rate,
			tier1_threshold, tier1_rate, tier2_threshold, tier2_rate,
			tier3_threshold, tier3_rate, min_profit_threshold
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := database.DB.Exec(query,
		config.EmployeeCode, config.BaseCommissionRate, config.NewCustomerBonusRate,
		config.Tier1Threshold, config.Tier1Rate, config.Tier2Threshold, config.Tier2Rate,
		config.Tier3Threshold, config.Tier3Rate, config.MinProfitThreshold,
	)
	if err != nil {
		return nil, err
	}

	// 重新查询获取ID和时间戳
	return GetSalesCommissionConfig(employeeCode)
}

// UpdateSalesCommissionConfig 更新销售员的分成配置
func UpdateSalesCommissionConfig(employeeCode string, config *SalesCommissionConfig) error {
	query := `
		UPDATE sales_commission_config
		SET base_commission_rate = ?, new_customer_bonus_rate = ?,
		    tier1_threshold = ?, tier1_rate = ?,
		    tier2_threshold = ?, tier2_rate = ?,
		    tier3_threshold = ?, tier3_rate = ?,
		    min_profit_threshold = ?,
		    updated_at = NOW()
		WHERE employee_code = ?
	`
	_, err := database.DB.Exec(query,
		config.BaseCommissionRate, config.NewCustomerBonusRate,
		config.Tier1Threshold, config.Tier1Rate,
		config.Tier2Threshold, config.Tier2Rate,
		config.Tier3Threshold, config.Tier3Rate,
		config.MinProfitThreshold,
		employeeCode,
	)
	return err
}

// HasSettledOrder 判断用户是否有已结算的有效订单
func HasSettledOrder(userID int) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM orders o
		WHERE o.user_id = ? 
		  AND o.status = 'paid'
		  AND o.settlement_date IS NOT NULL
		  AND (o.total_amount - (o.goods_amount - COALESCE(o.order_profit, 0)) - 
		       COALESCE(JSON_EXTRACT(o.delivery_fee_calculation, '$.total_platform_cost'), 0)) > 5
	`
	var count int
	err := database.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsNewCustomerOrder 判断订单是否为新客户首单
// 规则：
// 1. 该用户在此订单之前没有已结算的有效订单
// 2. 该用户没有已计入分成的新客订单（即使未结算，但已经标记为新客订单的）
// 3. 如果订单被取消（cancelled），不影响新客判断
func IsNewCustomerOrder(userID int, orderID int) (bool, error) {
	// 1. 查询该用户在此订单之前是否有已结算的有效订单
	query1 := `
		SELECT COUNT(*) 
		FROM orders o
		WHERE o.user_id = ? 
		  AND o.id < ?
		  AND o.status = 'paid'
		  AND o.settlement_date IS NOT NULL
		  AND (o.total_amount - (o.goods_amount - COALESCE(o.order_profit, 0)) - 
		       COALESCE(JSON_EXTRACT(o.delivery_fee_calculation, '$.total_platform_cost'), 0)) > 5
	`
	var count1 int
	err := database.DB.QueryRow(query1, userID, orderID).Scan(&count1)
	if err != nil {
		return false, err
	}
	if count1 > 0 {
		// 已有已结算的有效订单，不是新客户
		return false, nil
	}

	// 2. 查询该用户是否有已计入分成的新客订单（订单状态不是 cancelled）
	// 这样可以避免同一用户多个未结算订单都被标记为新客订单
	query2 := `
		SELECT COUNT(*) 
		FROM sales_commissions sc
		INNER JOIN orders o ON sc.order_id = o.id
		WHERE sc.user_id = ?
		  AND sc.order_id != ?
		  AND sc.is_new_customer_order = 1
		  AND sc.is_accounted = 1
		  AND sc.is_accounted_cancelled = 0
		  AND o.status != 'cancelled'
	`
	var count2 int
	err = database.DB.QueryRow(query2, userID, orderID).Scan(&count2)
	if err != nil {
		return false, err
	}
	if count2 > 0 {
		// 已有已计入分成的新客订单，不是新客户
		return false, nil
	}

	// 3. 检查是否有其他未结算但已创建分成记录的新客订单（订单状态不是 cancelled）
	// 这样可以避免同一用户多个未结算订单都被标记为新客订单
	query3 := `
		SELECT COUNT(*) 
		FROM sales_commissions sc
		INNER JOIN orders o ON sc.order_id = o.id
		WHERE sc.user_id = ?
		  AND sc.order_id != ?
		  AND sc.is_new_customer_order = 1
		  AND o.status != 'cancelled'
		  AND o.status != 'paid'
	`
	var count3 int
	err = database.DB.QueryRow(query3, userID, orderID).Scan(&count3)
	if err != nil {
		return false, err
	}
	if count3 > 0 {
		// 已有未结算的新客订单记录，不是新客户
		return false, nil
	}

	// 满足所有条件，是新客户首单
	return true, nil
}

// CancelOrderCommissions 取消订单的分成记录（订单取消时调用）
// 如果该订单是新客订单，取消后其他订单可以重新计算新客激励
func CancelOrderCommissions(orderID int) error {
	// 删除或标记该订单的分成记录
	// 注意：如果已经计入或结算，需要谨慎处理
	// 这里我们只删除未计入且未结算的记录，已计入或已结算的需要通过其他方式处理
	query := `
		DELETE FROM sales_commissions
		WHERE order_id = ?
		  AND is_accounted = 0
		  AND is_settled = 0
	`
	_, err := database.DB.Exec(query, orderID)
	if err != nil {
		log.Printf("取消订单 %d 的分成记录失败: %v", orderID, err)
		return err
	}
	
	// 对于已计入但未结算的新客订单记录，标记为取消计入
	// 这样其他订单就可以重新计算新客激励了
	updateQuery := `
		UPDATE sales_commissions
		SET is_accounted_cancelled = 1,
		    is_accounted = 0,
		    accounted_at = NULL,
		    updated_at = NOW()
		WHERE order_id = ?
		  AND is_new_customer_order = 1
		  AND is_accounted = 1
		  AND is_settled = 0
	`
	_, err = database.DB.Exec(updateQuery, orderID)
	if err != nil {
		log.Printf("取消订单 %d 的新客分成记录失败: %v", orderID, err)
		return err
	}
	
	return nil
}

// CalculateSalesCommission 计算销售分成
// orderAmount: 平台总收入（total_amount）
// goodsCost: 商品总成本（goods_amount - order_profit）
// deliveryCost: 配送成本（从delivery_fee_calculation中获取total_platform_cost）
// isNewCustomer: 是否新客户首单
// monthTotalSales: 当月有效订单总金额（用于计算阶梯提成）
func CalculateSalesCommission(employeeCode string, orderAmount, goodsCost, deliveryCost float64, isNewCustomer bool, monthTotalSales float64) (*CommissionCalculationResult, error) {
	// 获取配置
	config, err := GetSalesCommissionConfig(employeeCode)
	if err != nil {
		return nil, fmt.Errorf("获取分成配置失败: %v", err)
	}

	result := &CommissionCalculationResult{}

	// 计算订单利润
	result.OrderProfit = orderAmount - goodsCost - deliveryCost

	// 判断是否有效订单
	result.IsValidOrder = result.OrderProfit > config.MinProfitThreshold

	// 如果不是有效订单，直接返回
	if !result.IsValidOrder {
		return result, nil
	}

	result.IsNewCustomerOrder = isNewCustomer

	// 1. 计算基础提成
	result.BaseCommission = result.OrderProfit * config.BaseCommissionRate

	// 2. 计算新客开发激励（仅针对新客户首单）
	if isNewCustomer {
		result.NewCustomerBonus = result.OrderProfit * config.NewCustomerBonusRate
	}

	// 3. 计算业绩阶梯提成（全量补差方式）
	// 基于当月有效订单总金额计算
	tierLevel := 0
	tierRate := 0.0

	if monthTotalSales > config.Tier3Threshold {
		tierLevel = 3
		tierRate = config.Tier3Rate
	} else if monthTotalSales > config.Tier2Threshold {
		tierLevel = 2
		tierRate = config.Tier2Rate
	} else if monthTotalSales > config.Tier1Threshold {
		tierLevel = 1
		tierRate = config.Tier1Rate
	}

	result.TierLevel = tierLevel
	if tierLevel > 0 {
		// 阶梯提成 = 订单利润 × 阶梯比例
		result.TierCommission = result.OrderProfit * tierRate
	}

	// 4. 计算总分成
	result.TotalCommission = result.BaseCommission + result.NewCustomerBonus + result.TierCommission

	return result, nil
}

// SaveSalesCommission 保存销售分成记录
func SaveSalesCommission(commission *SalesCommission) error {
	query := `
		INSERT INTO sales_commissions (
			order_id, employee_code, user_id, order_number, order_date,
			settlement_date, is_valid_order, is_new_customer_order,
			order_amount, goods_cost, delivery_cost, order_profit,
			base_commission, new_customer_bonus, tier_commission,
			total_commission, tier_level, calculation_month,
			is_accounted, accounted_at, is_settled, settled_at, is_accounted_cancelled
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			settlement_date = VALUES(settlement_date),
			is_valid_order = VALUES(is_valid_order),
			is_new_customer_order = VALUES(is_new_customer_order),
			order_amount = VALUES(order_amount),
			goods_cost = VALUES(goods_cost),
			delivery_cost = VALUES(delivery_cost),
			order_profit = VALUES(order_profit),
			base_commission = VALUES(base_commission),
			new_customer_bonus = VALUES(new_customer_bonus),
			tier_commission = VALUES(tier_commission),
			total_commission = VALUES(total_commission),
			tier_level = VALUES(tier_level),
			calculation_month = VALUES(calculation_month),
			-- 如果已取消计入，不更新计入相关字段
			is_accounted = IF(is_accounted_cancelled = 1, is_accounted, VALUES(is_accounted)),
			accounted_at = IF(is_accounted_cancelled = 1, accounted_at, VALUES(accounted_at)),
			updated_at = NOW()
	`
	var settlementDate interface{}
	if commission.SettlementDate != nil {
		settlementDate = commission.SettlementDate
	} else {
		settlementDate = nil
	}

	isValidOrder := 0
	if commission.IsValidOrder {
		isValidOrder = 1
	}
	isNewCustomerOrder := 0
	if commission.IsNewCustomerOrder {
		isNewCustomerOrder = 1
	}

	isAccounted := 0
	if commission.IsAccounted {
		isAccounted = 1
	}
	var accountedAt interface{}
	if commission.AccountedAt != nil {
		accountedAt = commission.AccountedAt
	} else {
		accountedAt = nil
	}

	isSettled := 0
	if commission.IsSettled {
		isSettled = 1
	}
	var settledAt interface{}
	if commission.SettledAt != nil {
		settledAt = commission.SettledAt
	} else {
		settledAt = nil
	}

	isAccountedCancelled := 0
	if commission.IsAccountedCancelled {
		isAccountedCancelled = 1
	}

	_, err := database.DB.Exec(query,
		commission.OrderID, commission.EmployeeCode, commission.UserID, commission.OrderNumber,
		commission.OrderDate, settlementDate, isValidOrder, isNewCustomerOrder,
		commission.OrderAmount, commission.GoodsCost, commission.DeliveryCost, commission.OrderProfit,
		commission.BaseCommission, commission.NewCustomerBonus, commission.TierCommission,
		commission.TotalCommission, commission.TierLevel, commission.CalculationMonth,
		isAccounted, accountedAt, isSettled, settledAt, isAccountedCancelled,
	)
	return err
}

// GetSalesCommissionsByEmployee 获取销售员的分成记录列表
// status: "all" - 全部, "accounted" - 已计入, "settled" - 已结算, "unaccounted" - 未计入, "unsettled" - 未结算
func GetSalesCommissionsByEmployee(employeeCode string, month string, status string, startDate, endDate *time.Time, pageNum, pageSize int) ([]SalesCommission, int, error) {
	offset := (pageNum - 1) * pageSize

	// 构建查询条件
	where := "employee_code = ?"
	args := []interface{}{employeeCode}

	if month != "" {
		where += " AND calculation_month = ?"
		args = append(args, month)
	}

	// 状态筛选（已取消计入的记录需要特殊处理）
	if status == "accounted" {
		// 已计入：排除已取消的记录
		where += " AND is_accounted = 1 AND is_accounted_cancelled = 0"
	} else if status == "settled" {
		// 已结算：排除已取消的记录
		where += " AND is_settled = 1 AND is_accounted_cancelled = 0"
	} else if status == "unaccounted" {
		// 未计入：包括已取消的记录（因为取消后 is_accounted = 0）
		where += " AND is_accounted = 0"
	} else if status == "unsettled" {
		// 未结算：排除已取消的记录（因为已取消的记录不能结算）
		where += " AND is_settled = 0 AND is_accounted_cancelled = 0"
	} else if status == "invalid" {
		// 无效订单
		where += " AND is_valid_order = 0"
	}

	// 日期范围筛选
	if startDate != nil {
		where += " AND order_date >= ?"
		args = append(args, startDate.Format("2006-01-02"))
	}
	if endDate != nil {
		where += " AND order_date <= ?"
		args = append(args, endDate.Format("2006-01-02"))
	}

	// 查询总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sales_commissions WHERE %s", where)
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表（关联查询地址名称）
	query := fmt.Sprintf(`
		SELECT sc.id, sc.order_id, sc.employee_code, sc.user_id, sc.order_number, sc.order_date,
		       sc.settlement_date, sc.is_valid_order, sc.is_new_customer_order,
		       sc.order_amount, sc.goods_cost, sc.delivery_cost, sc.order_profit,
		       sc.base_commission, sc.new_customer_bonus, sc.tier_commission,
		       sc.total_commission, sc.tier_level, sc.calculation_month,
		       sc.is_accounted, sc.accounted_at, sc.is_settled, sc.settled_at, sc.is_accounted_cancelled,
		       sc.created_at, sc.updated_at,
		       a.name AS address_name
		FROM sales_commissions sc
		LEFT JOIN orders o ON sc.order_id = o.id
		LEFT JOIN mini_app_addresses a ON o.address_id = a.id
		WHERE %s
		ORDER BY sc.order_date DESC, sc.id DESC
		LIMIT ? OFFSET ?
	`, where)
	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	commissions := make([]SalesCommission, 0)
	for rows.Next() {
		var commission SalesCommission
		var settlementDate, accountedAt, settledAt sql.NullTime
		var isValidOrder, isNewCustomerOrder, isAccounted, isSettled, isAccountedCancelled int
		var addressName sql.NullString

		err := rows.Scan(
			&commission.ID, &commission.OrderID, &commission.EmployeeCode, &commission.UserID,
			&commission.OrderNumber, &commission.OrderDate, &settlementDate,
			&isValidOrder, &isNewCustomerOrder,
			&commission.OrderAmount, &commission.GoodsCost, &commission.DeliveryCost,
			&commission.OrderProfit, &commission.BaseCommission, &commission.NewCustomerBonus,
			&commission.TierCommission, &commission.TotalCommission, &commission.TierLevel,
			&commission.CalculationMonth,
			&isAccounted, &accountedAt, &isSettled, &settledAt, &isAccountedCancelled,
			&commission.CreatedAt, &commission.UpdatedAt,
			&addressName, // 读取地址名称，但不在结构体中存储，在API层处理
		)
		if err != nil {
			return nil, 0, err
		}

		commission.IsValidOrder = isValidOrder == 1
		commission.IsNewCustomerOrder = isNewCustomerOrder == 1
		commission.IsAccounted = isAccounted == 1
		commission.IsSettled = isSettled == 1
		commission.IsAccountedCancelled = isAccountedCancelled == 1
		if settlementDate.Valid {
			t := settlementDate.Time
			commission.SettlementDate = &t
		}
		if accountedAt.Valid {
			t := accountedAt.Time
			commission.AccountedAt = &t
		}
		if settledAt.Valid {
			t := settledAt.Time
			commission.SettledAt = &t
		}

		// 将地址名称添加到结构体的JSON标签中（通过反射或手动构建）
		// 由于结构体没有AddressName字段，我们需要在API层处理
		commissions = append(commissions, commission)
	}

	return commissions, total, nil
}

// GetSalesCommissionsByOrderIDs 批量获取订单的销售分成记录
func GetSalesCommissionsByOrderIDs(orderIDs []int) ([]*SalesCommission, error) {
	if len(orderIDs) == 0 {
		return []*SalesCommission{}, nil
	}

	placeholders := ""
	args := make([]interface{}, len(orderIDs))
	for i, id := range orderIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, order_id, employee_code, user_id, order_number, order_date,
		       settlement_date, is_valid_order, is_new_customer_order,
		       order_amount, goods_cost, delivery_cost, order_profit,
		       base_commission, new_customer_bonus, tier_commission,
		       total_commission, tier_level, calculation_month,
		       is_accounted, accounted_at, is_settled, settled_at, is_accounted_cancelled,
		       created_at, updated_at
		FROM sales_commissions
		WHERE order_id IN (%s)
	`, placeholders)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	commissions := make([]*SalesCommission, 0)
	for rows.Next() {
		var commission SalesCommission
		var settlementDate, accountedAt, settledAt sql.NullTime
		var isValidOrder, isNewCustomerOrder, isAccounted, isSettled, isAccountedCancelled int

		err := rows.Scan(
			&commission.ID, &commission.OrderID, &commission.EmployeeCode, &commission.UserID,
			&commission.OrderNumber, &commission.OrderDate, &settlementDate,
			&isValidOrder, &isNewCustomerOrder,
			&commission.OrderAmount, &commission.GoodsCost, &commission.DeliveryCost,
			&commission.OrderProfit, &commission.BaseCommission, &commission.NewCustomerBonus,
			&commission.TierCommission, &commission.TotalCommission, &commission.TierLevel,
			&commission.CalculationMonth,
			&isAccounted, &accountedAt, &isSettled, &settledAt, &isAccountedCancelled,
			&commission.CreatedAt, &commission.UpdatedAt,
		)
		if err != nil {
			continue
		}

		commission.IsValidOrder = isValidOrder == 1
		commission.IsNewCustomerOrder = isNewCustomerOrder == 1
		commission.IsAccounted = isAccounted == 1
		commission.IsSettled = isSettled == 1
		commission.IsAccountedCancelled = isAccountedCancelled == 1
		if settlementDate.Valid {
			t := settlementDate.Time
			commission.SettlementDate = &t
		}
		if accountedAt.Valid {
			t := accountedAt.Time
			commission.AccountedAt = &t
		}
		if settledAt.Valid {
			t := settledAt.Time
			commission.SettledAt = &t
		}

		commissions = append(commissions, &commission)
	}

	return commissions, nil
}

// GetMonthlyTotalSales 获取销售员指定月份的有效订单总金额（排除已取消计入的记录）
func GetMonthlyTotalSales(employeeCode string, month string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(order_amount), 0)
		FROM sales_commissions
		WHERE employee_code = ?
		  AND calculation_month = ?
		  AND is_valid_order = 1
		  AND is_accounted_cancelled = 0
	`
	var totalSales float64
	err := database.DB.QueryRow(query, employeeCode, month).Scan(&totalSales)
	if err != nil {
		return 0, err
	}
	return totalSales, nil
}

// GetSalesCommissionMonthlyStats 获取销售员的分成月统计
func GetSalesCommissionMonthlyStats(employeeCode string, month string) (*SalesCommissionMonthlyStats, error) {
	var stats SalesCommissionMonthlyStats
	query := `
		SELECT id, employee_code, stat_month,
		       total_sales_amount, total_valid_orders, total_new_customers,
		       total_profit, total_base_commission, total_new_customer_bonus,
		       total_tier_commission, total_commission, tier_level,
		       created_at, updated_at
		FROM sales_commission_monthly_stats
		WHERE employee_code = ? AND stat_month = ?
	`
	err := database.DB.QueryRow(query, employeeCode, month).Scan(
		&stats.ID, &stats.EmployeeCode, &stats.StatMonth,
		&stats.TotalSalesAmount, &stats.TotalValidOrders, &stats.TotalNewCustomers,
		&stats.TotalProfit, &stats.TotalBaseCommission, &stats.TotalNewCustomerBonus,
		&stats.TotalTierCommission, &stats.TotalCommission, &stats.TierLevel,
		&stats.CreatedAt, &stats.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &stats, nil
}

// CalculateAndSaveMonthlyStats 计算并保存月统计（排除已取消计入的记录）
func CalculateAndSaveMonthlyStats(employeeCode string, month string) error {
	// 查询该月的所有有效订单分成记录（排除已取消计入的记录）
	query := `
		SELECT 
			COALESCE(SUM(order_amount), 0) as total_sales_amount,
			COUNT(*) as total_valid_orders,
			SUM(CASE WHEN is_new_customer_order = 1 THEN 1 ELSE 0 END) as total_new_customers,
			COALESCE(SUM(order_profit), 0) as total_profit,
			COALESCE(SUM(base_commission), 0) as total_base_commission,
			COALESCE(SUM(new_customer_bonus), 0) as total_new_customer_bonus,
			COALESCE(SUM(tier_commission), 0) as total_tier_commission,
			COALESCE(SUM(total_commission), 0) as total_commission,
			MAX(tier_level) as tier_level
		FROM sales_commissions
		WHERE employee_code = ? 
		  AND calculation_month = ? 
		  AND is_valid_order = 1
		  AND is_accounted_cancelled = 0
	`

	var stats SalesCommissionMonthlyStats
	stats.EmployeeCode = employeeCode
	stats.StatMonth = month

	err := database.DB.QueryRow(query, employeeCode, month).Scan(
		&stats.TotalSalesAmount, &stats.TotalValidOrders, &stats.TotalNewCustomers,
		&stats.TotalProfit, &stats.TotalBaseCommission, &stats.TotalNewCustomerBonus,
		&stats.TotalTierCommission, &stats.TotalCommission, &stats.TierLevel,
	)
	if err != nil {
		return err
	}

	// 保存或更新统计
	saveQuery := `
		INSERT INTO sales_commission_monthly_stats (
			employee_code, stat_month,
			total_sales_amount, total_valid_orders, total_new_customers,
			total_profit, total_base_commission, total_new_customer_bonus,
			total_tier_commission, total_commission, tier_level
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			total_sales_amount = VALUES(total_sales_amount),
			total_valid_orders = VALUES(total_valid_orders),
			total_new_customers = VALUES(total_new_customers),
			total_profit = VALUES(total_profit),
			total_base_commission = VALUES(total_base_commission),
			total_new_customer_bonus = VALUES(total_new_customer_bonus),
			total_tier_commission = VALUES(total_tier_commission),
			total_commission = VALUES(total_commission),
			tier_level = VALUES(tier_level),
			updated_at = NOW()
	`
	_, err = database.DB.Exec(saveQuery,
		stats.EmployeeCode, stats.StatMonth,
		stats.TotalSalesAmount, stats.TotalValidOrders, stats.TotalNewCustomers,
		stats.TotalProfit, stats.TotalBaseCommission, stats.TotalNewCustomerBonus,
		stats.TotalTierCommission, stats.TotalCommission, stats.TierLevel,
	)
	return err
}

// ProcessOrderSettlement 处理订单结算时的分成计算
func ProcessOrderSettlement(orderID int) error {
	// 获取订单信息
	order, err := GetOrderByID(orderID)
	if err != nil {
		return fmt.Errorf("获取订单失败: %v", err)
	}
	if order == nil {
		return fmt.Errorf("订单不存在")
	}

	// 订单必须是已结算状态
	if order.Status != "paid" || order.SettlementDate == nil {
		return fmt.Errorf("订单未结算，无法计算分成")
	}

	// 获取订单对应的销售员（通过用户）
	user, err := GetMiniAppUserByID(order.UserID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %v", err)
	}
	if user == nil || user.SalesCode == "" {
		// 没有销售员，不需要计算分成
		return nil
	}

	// 查询订单的配送费计算结果（JSON字段）
	var deliveryFeeCalcJSON sql.NullString
	err = database.DB.QueryRow("SELECT delivery_fee_calculation FROM orders WHERE id = ?", orderID).Scan(&deliveryFeeCalcJSON)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("查询delivery_fee_calculation失败: %v", err)
	}

	// 获取订单的利润和成本信息
	// 平台总收入 = total_amount
	orderAmount := order.TotalAmount

	// 商品总成本 = goods_amount - order_profit
	goodsCost := order.GoodsAmount
	if order.OrderProfit != nil && *order.OrderProfit > 0 {
		goodsCost = order.GoodsAmount - *order.OrderProfit
	}

	// 配送成本（从delivery_fee_calculation中获取total_platform_cost）
	deliveryCost := 0.0
	if deliveryFeeCalcJSON.Valid && deliveryFeeCalcJSON.String != "" {
		deliveryCost = extractDeliveryCostFromJSON(deliveryFeeCalcJSON.String)
	}

	// 判断是否新客户首单
	isNewCustomer, err := IsNewCustomerOrder(order.UserID, orderID)
	if err != nil {
		log.Printf("判断是否新客户失败: %v", err)
		isNewCustomer = false
	}

	// 计算月份（YYYY-MM格式）
	settlementMonth := order.SettlementDate.Format("2006-01")

	// 获取当月有效订单总金额（用于计算阶梯提成）
	monthTotalSales, err := GetMonthlyTotalSales(user.SalesCode, settlementMonth)
	if err != nil {
		log.Printf("获取当月总销售额失败: %v", err)
		monthTotalSales = 0
	}

	// 计算分成
	calcResult, err := CalculateSalesCommission(
		user.SalesCode, orderAmount, goodsCost, deliveryCost,
		isNewCustomer, monthTotalSales,
	)
	if err != nil {
		return fmt.Errorf("计算分成失败: %v", err)
	}

	// 检查是否已存在记录且已取消计入
	var existingCommission *SalesCommission
	existingCommissions, err := GetSalesCommissionsByOrderIDs([]int{orderID})
	if err == nil && len(existingCommissions) > 0 {
		for _, c := range existingCommissions {
			if c.EmployeeCode == user.SalesCode {
				existingCommission = c
				break
			}
		}
	}

	// 如果已取消计入，不再自动计入
	if existingCommission != nil && existingCommission.IsAccountedCancelled {
		log.Printf("订单 %d 的销售分成已取消计入，不再自动计入", orderID)
		return nil
	}

	// 保存分成记录
	// 订单标记为已收款后自动计入
	now := time.Now()
	accountedAt := &now
	commission := &SalesCommission{
		OrderID:            orderID,
		EmployeeCode:       user.SalesCode,
		UserID:             order.UserID,
		OrderNumber:        order.OrderNumber,
		OrderDate:          order.CreatedAt,
		SettlementDate:     order.SettlementDate,
		IsValidOrder:       calcResult.IsValidOrder,
		IsNewCustomerOrder: calcResult.IsNewCustomerOrder,
		OrderAmount:        orderAmount,
		GoodsCost:          goodsCost,
		DeliveryCost:       deliveryCost,
		OrderProfit:        calcResult.OrderProfit,
		BaseCommission:     calcResult.BaseCommission,
		NewCustomerBonus:   calcResult.NewCustomerBonus,
		TierCommission:     calcResult.TierCommission,
		TotalCommission:   calcResult.TotalCommission,
		TierLevel:          calcResult.TierLevel,
		CalculationMonth:   settlementMonth,
		IsAccounted:        true,  // 自动计入
		AccountedAt:        accountedAt,
		IsSettled:          false,
		IsAccountedCancelled: false,
	}

	err = SaveSalesCommission(commission)
	if err != nil {
		return fmt.Errorf("保存分成记录失败: %v", err)
	}

	// 重新计算当月总销售额（因为新增了订单）
	monthTotalSales, err = GetMonthlyTotalSales(user.SalesCode, settlementMonth)
	if err != nil {
		log.Printf("重新获取当月总销售额失败: %v", err)
	} else {
		// 如果当月总销售额发生变化，需要重新计算所有订单的阶梯提成
		// 因为阶梯提成是基于当月总销售额的
		err = RecalculateTierCommissionsForMonth(user.SalesCode, settlementMonth, monthTotalSales)
		if err != nil {
			log.Printf("重新计算阶梯提成失败: %v", err)
		}
	}

	// 更新月统计
	err = CalculateAndSaveMonthlyStats(user.SalesCode, settlementMonth)
	if err != nil {
		log.Printf("更新月统计失败: %v", err)
	}

	return nil
}

// RecalculateTierCommissionsForMonth 重新计算指定月份的阶梯提成
func RecalculateTierCommissionsForMonth(employeeCode string, month string, monthTotalSales float64) error {
	// 获取配置
	config, err := GetSalesCommissionConfig(employeeCode)
	if err != nil {
		return err
	}

	// 确定阶梯等级和比例
	tierLevel := 0
	tierRate := 0.0

	if monthTotalSales > config.Tier3Threshold {
		tierLevel = 3
		tierRate = config.Tier3Rate
	} else if monthTotalSales > config.Tier2Threshold {
		tierLevel = 2
		tierRate = config.Tier2Rate
	} else if monthTotalSales > config.Tier1Threshold {
		tierLevel = 1
		tierRate = config.Tier1Rate
	}

	// 更新该月所有有效订单的阶梯提成（排除已取消计入的记录）
	updateQuery := `
		UPDATE sales_commissions
		SET tier_commission = order_profit * ?,
		    tier_level = ?,
		    total_commission = base_commission + new_customer_bonus + (order_profit * ?),
		    updated_at = NOW()
		WHERE employee_code = ?
		  AND calculation_month = ?
		  AND is_valid_order = 1
		  AND is_accounted_cancelled = 0
	`
	_, err = database.DB.Exec(updateQuery, tierRate, tierLevel, tierRate, employeeCode, month)
	if err != nil {
		return err
	}

	return nil
}

// extractDeliveryCostFromJSON 从JSON中提取配送成本
func extractDeliveryCostFromJSON(jsonStr string) float64 {
	if jsonStr == "" {
		return 0.0
	}

	var result DeliveryFeeCalculationResult
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		log.Printf("解析delivery_fee_calculation JSON失败: %v", err)
		return 0.0
	}

	return result.TotalPlatformCost
}

// GetAllSalesCommissionMonthlyStats 获取所有销售员的分成月统计
func GetAllSalesCommissionMonthlyStats(month string) ([]SalesCommissionMonthlyStats, error) {
	query := `
		SELECT id, employee_code, stat_month,
		       total_sales_amount, total_valid_orders, total_new_customers,
		       total_profit, total_base_commission, total_new_customer_bonus,
		       total_tier_commission, total_commission, tier_level,
		       created_at, updated_at
		FROM sales_commission_monthly_stats
		WHERE stat_month = ?
		ORDER BY total_commission DESC
	`
	rows, err := database.DB.Query(query, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statsList := make([]SalesCommissionMonthlyStats, 0)
	for rows.Next() {
		var stats SalesCommissionMonthlyStats
		err := rows.Scan(
			&stats.ID, &stats.EmployeeCode, &stats.StatMonth,
			&stats.TotalSalesAmount, &stats.TotalValidOrders, &stats.TotalNewCustomers,
			&stats.TotalProfit, &stats.TotalBaseCommission, &stats.TotalNewCustomerBonus,
			&stats.TotalTierCommission, &stats.TotalCommission, &stats.TierLevel,
			&stats.CreatedAt, &stats.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		statsList = append(statsList, stats)
	}

	return statsList, nil
}

// AccountSalesCommissions 批量计入销售分成（标记为已计入）
// commissionIDs: 分成记录ID列表，如果为空则根据其他条件批量计入
// employeeCode: 销售员员工码（可选）
// startDate, endDate: 日期范围（可选）
func AccountSalesCommissions(commissionIDs []int, employeeCode string, startDate, endDate *time.Time) (int64, error) {
	now := time.Now()
	
	if len(commissionIDs) > 0 {
		// 按ID批量计入
		placeholders := ""
		args := make([]interface{}, len(commissionIDs))
		for i, id := range commissionIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args[i] = id
		}
		
		query := fmt.Sprintf(`
			UPDATE sales_commissions
			SET is_accounted = 1, accounted_at = ?, is_accounted_cancelled = 0, updated_at = NOW()
			WHERE id IN (%s) AND is_accounted = 0 AND is_accounted_cancelled = 0
		`, placeholders)
		args = append([]interface{}{now}, args...)
		
		result, err := database.DB.Exec(query, args...)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
	
	// 按条件批量计入
	where := "is_accounted = 0 AND is_accounted_cancelled = 0"
	args := []interface{}{now}
	
	if employeeCode != "" {
		where += " AND employee_code = ?"
		args = append(args, employeeCode)
	}
	if startDate != nil {
		where += " AND order_date >= ?"
		args = append(args, startDate.Format("2006-01-02"))
	}
	if endDate != nil {
		where += " AND order_date <= ?"
		args = append(args, endDate.Format("2006-01-02"))
	}
	
		query := fmt.Sprintf(`
			UPDATE sales_commissions
			SET is_accounted = 1, accounted_at = ?, is_accounted_cancelled = 0, updated_at = NOW()
			WHERE %s
		`, where)
	
	result, err := database.DB.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// SettleSalesCommissions 批量结算销售分成（标记为已结算）
// commissionIDs: 分成记录ID列表，如果为空则根据其他条件批量结算
// employeeCode: 销售员员工码（可选）
// startDate, endDate: 日期范围（可选）
// 注意：只有已计入的记录才能被结算
func SettleSalesCommissions(commissionIDs []int, employeeCode string, startDate, endDate *time.Time) (int64, error) {
	now := time.Now()
	
	if len(commissionIDs) > 0 {
		// 按ID批量结算
		placeholders := ""
		args := make([]interface{}, len(commissionIDs))
		for i, id := range commissionIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args[i] = id
		}
		
		query := fmt.Sprintf(`
			UPDATE sales_commissions
			SET is_settled = 1, settled_at = ?, updated_at = NOW()
			WHERE id IN (%s) AND is_accounted = 1 AND is_settled = 0 AND is_accounted_cancelled = 0
		`, placeholders)
		args = append([]interface{}{now}, args...)
		
		result, err := database.DB.Exec(query, args...)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
	
	// 按条件批量结算
	where := "is_accounted = 1 AND is_settled = 0 AND is_accounted_cancelled = 0"
	args := []interface{}{now}
	
	if employeeCode != "" {
		where += " AND employee_code = ?"
		args = append(args, employeeCode)
	}
	if startDate != nil {
		where += " AND order_date >= ?"
		args = append(args, startDate.Format("2006-01-02"))
	}
	if endDate != nil {
		where += " AND order_date <= ?"
		args = append(args, endDate.Format("2006-01-02"))
	}
	
	query := fmt.Sprintf(`
		UPDATE sales_commissions
		SET is_settled = 1, settled_at = ?, updated_at = NOW()
		WHERE %s
	`, where)
	
	result, err := database.DB.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetSalesCommissionByID 根据ID获取销售分成记录
func GetSalesCommissionByID(id int) (*SalesCommission, error) {
	var commission SalesCommission
	var settlementDate, accountedAt, settledAt sql.NullTime
	var isValidOrder, isNewCustomerOrder, isAccounted, isSettled, isAccountedCancelled int
	
	query := `
		SELECT id, order_id, employee_code, user_id, order_number, order_date,
		       settlement_date, is_valid_order, is_new_customer_order,
		       order_amount, goods_cost, delivery_cost, order_profit,
		       base_commission, new_customer_bonus, tier_commission,
		       total_commission, tier_level, calculation_month,
		       is_accounted, accounted_at, is_settled, settled_at, is_accounted_cancelled,
		       created_at, updated_at
		FROM sales_commissions
		WHERE id = ?
	`
	
	err := database.DB.QueryRow(query, id).Scan(
		&commission.ID, &commission.OrderID, &commission.EmployeeCode, &commission.UserID,
		&commission.OrderNumber, &commission.OrderDate, &settlementDate,
		&isValidOrder, &isNewCustomerOrder,
		&commission.OrderAmount, &commission.GoodsCost, &commission.DeliveryCost,
		&commission.OrderProfit, &commission.BaseCommission, &commission.NewCustomerBonus,
		&commission.TierCommission, &commission.TotalCommission, &commission.TierLevel,
		&commission.CalculationMonth,
		&isAccounted, &accountedAt, &isSettled, &settledAt, &isAccountedCancelled,
		&commission.CreatedAt, &commission.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	commission.IsValidOrder = isValidOrder == 1
	commission.IsNewCustomerOrder = isNewCustomerOrder == 1
	commission.IsAccounted = isAccounted == 1
	commission.IsSettled = isSettled == 1
	commission.IsAccountedCancelled = isAccountedCancelled == 1
	if settlementDate.Valid {
		t := settlementDate.Time
		commission.SettlementDate = &t
	}
	if accountedAt.Valid {
		t := accountedAt.Time
		commission.AccountedAt = &t
	}
	if settledAt.Valid {
		t := settledAt.Time
		commission.SettledAt = &t
	}
	
	return &commission, nil
}

// CancelAccountSalesCommissions 取消计入销售分成（仅支持已计入未结算的记录）
// commissionIDs: 分成记录ID列表
func CancelAccountSalesCommissions(commissionIDs []int) (int64, error) {
	if len(commissionIDs) == 0 {
		return 0, fmt.Errorf("请提供要取消计入的记录ID")
	}

	placeholders := ""
	args := make([]interface{}, len(commissionIDs))
	for i, id := range commissionIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	// 先查询要取消的记录，获取月份信息，以便后续重新计算统计
	queryRecords := fmt.Sprintf(`
		SELECT DISTINCT employee_code, calculation_month
		FROM sales_commissions
		WHERE id IN (%s)
		  AND is_accounted = 1
		  AND is_settled = 0
	`, placeholders)

	rows, err := database.DB.Query(queryRecords, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	// 收集需要重新计算统计的月份
	monthsToRecalc := make(map[string]string) // key: employeeCode, value: month
	for rows.Next() {
		var employeeCode, month string
		if err := rows.Scan(&employeeCode, &month); err == nil {
			monthsToRecalc[employeeCode] = month
		}
	}

	// 只允许取消已计入但未结算的记录
	query := fmt.Sprintf(`
		UPDATE sales_commissions
		SET is_accounted = 0,
		    accounted_at = NULL,
		    is_accounted_cancelled = 1,
		    updated_at = NOW()
		WHERE id IN (%s)
		  AND is_accounted = 1
		  AND is_settled = 0
	`, placeholders)

	result, err := database.DB.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	// 重新计算受影响月份的统计（因为取消计入会影响总分成）
	for employeeCode, month := range monthsToRecalc {
		// 重新计算月销售额（用于阶梯提成）
		monthTotalSales, err := GetMonthlyTotalSales(employeeCode, month)
		if err == nil {
			// 重新计算阶梯提成
			_ = RecalculateTierCommissionsForMonth(employeeCode, month, monthTotalSales)
		}
		// 重新计算月统计
		_ = CalculateAndSaveMonthlyStats(employeeCode, month)
	}

	return rowsAffected, nil
}

// ResetAccountSalesCommissions 重新计入销售分成（重置分成，仅支持已取消计入的记录）
// commissionIDs: 分成记录ID列表
func ResetAccountSalesCommissions(commissionIDs []int) (int64, error) {
	if len(commissionIDs) == 0 {
		return 0, fmt.Errorf("请提供要重新计入的记录ID")
	}

	placeholders := ""
	args := make([]interface{}, len(commissionIDs))
	for i, id := range commissionIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	// 先查询要重新计入的记录，获取月份信息，以便后续重新计算统计
	queryRecords := fmt.Sprintf(`
		SELECT DISTINCT employee_code, calculation_month
		FROM sales_commissions
		WHERE id IN (%s)
		  AND is_accounted_cancelled = 1
	`, placeholders)

	rows, err := database.DB.Query(queryRecords, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	// 收集需要重新计算统计的月份
	monthsToRecalc := make(map[string]string) // key: employeeCode, value: month
	for rows.Next() {
		var employeeCode, month string
		if err := rows.Scan(&employeeCode, &month); err == nil {
			monthsToRecalc[employeeCode] = month
		}
	}

	// 只允许重新计入已取消的记录
	now := time.Now()
	query := fmt.Sprintf(`
		UPDATE sales_commissions
		SET is_accounted = 1,
		    accounted_at = ?,
		    is_accounted_cancelled = 0,
		    updated_at = NOW()
		WHERE id IN (%s)
		  AND is_accounted_cancelled = 1
	`, placeholders)

	args = append([]interface{}{now}, args...)
	result, err := database.DB.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	// 重新计算受影响月份的统计（因为重新计入会影响总分成）
	for employeeCode, month := range monthsToRecalc {
		// 重新计算月销售额（用于阶梯提成）
		monthTotalSales, err := GetMonthlyTotalSales(employeeCode, month)
		if err == nil {
			// 重新计算阶梯提成
			_ = RecalculateTierCommissionsForMonth(employeeCode, month, monthTotalSales)
		}
		// 重新计算月统计
		_ = CalculateAndSaveMonthlyStats(employeeCode, month)
	}

	return rowsAffected, nil
}

// SalesCommissionOverview 销售分成总览统计
type SalesCommissionOverview struct {
	TotalAmount        float64 `json:"total_amount"`         // 总金额（所有有效订单的总分成）
	UnaccountedAmount  float64 `json:"unaccounted_amount"`   // 未计入金额
	AccountedAmount    float64 `json:"accounted_amount"`     // 已计入金额
	SettledAmount      float64 `json:"settled_amount"`        // 已结算金额
	CancelledAmount    float64 `json:"cancelled_amount"`     // 取消计入金额
	UnaccountedCount   int     `json:"unaccounted_count"`    // 未计入数量
	AccountedCount     int     `json:"accounted_count"`       // 已计入数量
	SettledCount       int     `json:"settled_count"`        // 已结算数量
	CancelledCount     int     `json:"cancelled_count"`       // 取消计入数量
	InvalidOrderCount  int     `json:"invalid_order_count"`  // 无效订单数量
}

// GetSalesCommissionOverview 获取销售员的分成总览统计
func GetSalesCommissionOverview(employeeCode string, startDate, endDate *time.Time) (*SalesCommissionOverview, error) {
	// 构建查询条件
	where := "employee_code = ?"
	args := []interface{}{employeeCode}

	// 日期范围筛选
	if startDate != nil {
		where += " AND order_date >= ?"
		args = append(args, startDate.Format("2006-01-02"))
	}
	if endDate != nil {
		where += " AND order_date <= ?"
		args = append(args, endDate.Format("2006-01-02"))
	}

	// 1. 从sales_commissions表获取已收款订单的统计
	query := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(CASE WHEN is_valid_order = 1 AND is_accounted_cancelled = 0 THEN total_commission ELSE 0 END), 0) as total_amount,
			COALESCE(SUM(CASE WHEN is_valid_order = 1 AND is_accounted = 1 AND is_settled = 0 AND is_accounted_cancelled = 0 THEN total_commission ELSE 0 END), 0) as accounted_amount,
			COALESCE(SUM(CASE WHEN is_valid_order = 1 AND is_settled = 1 AND is_accounted_cancelled = 0 THEN total_commission ELSE 0 END), 0) as settled_amount,
			COALESCE(SUM(CASE WHEN is_valid_order = 1 AND is_accounted_cancelled = 1 THEN total_commission ELSE 0 END), 0) as cancelled_amount,
			COUNT(CASE WHEN is_valid_order = 1 AND is_accounted = 1 AND is_settled = 0 AND is_accounted_cancelled = 0 THEN 1 END) as accounted_count,
			COUNT(CASE WHEN is_valid_order = 1 AND is_settled = 1 AND is_accounted_cancelled = 0 THEN 1 END) as settled_count,
			COUNT(CASE WHEN is_valid_order = 1 AND is_accounted_cancelled = 1 THEN 1 END) as cancelled_count,
			COUNT(CASE WHEN is_valid_order = 0 THEN 1 END) as invalid_order_count
		FROM sales_commissions
		WHERE %s
	`, where)

	var overview SalesCommissionOverview
	err := database.DB.QueryRow(query, args...).Scan(
		&overview.TotalAmount,
		&overview.AccountedAmount,
		&overview.SettledAmount,
		&overview.CancelledAmount,
		&overview.AccountedCount,
		&overview.SettledCount,
		&overview.CancelledCount,
		&overview.InvalidOrderCount,
	)
	if err != nil {
		return nil, err
	}

	// 2. 查询所有未收款且未取消的订单，计算分润预览总和作为未计入金额
	unpaidWhere := "u.sales_code = ? AND o.status != 'paid' AND o.status != 'cancelled'"
	unpaidArgs := []interface{}{employeeCode}

	// 日期范围筛选（对于未收款订单，使用created_at）
	if startDate != nil {
		unpaidWhere += " AND DATE(o.created_at) >= ?"
		unpaidArgs = append(unpaidArgs, startDate.Format("2006-01-02"))
	}
	if endDate != nil {
		unpaidWhere += " AND DATE(o.created_at) <= ?"
		unpaidArgs = append(unpaidArgs, endDate.Format("2006-01-02"))
	}

	unpaidOrdersQuery := fmt.Sprintf(`
		SELECT
			o.id,
			o.order_number,
			o.status,
			o.total_amount,
			o.goods_amount,
			o.delivery_fee,
			o.order_profit,
			o.delivery_fee_calculation,
			o.created_at,
			o.user_id
		FROM orders o
		JOIN mini_app_users u ON o.user_id = u.id
		WHERE %s
		ORDER BY o.created_at DESC
	`, unpaidWhere)

	rows, err := database.DB.Query(unpaidOrdersQuery, unpaidArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var unaccountedAmount float64
	unaccountedCount := 0

	for rows.Next() {
		var id, userID int
		var orderNumber, status string
		var totalAmount, goodsAmount, deliveryFee float64
		var orderProfit sql.NullFloat64
		var deliveryFeeCalculation sql.NullString
		var createdAt time.Time

		err := rows.Scan(
			&id, &orderNumber, &status, &totalAmount, &goodsAmount,
			&deliveryFee, &orderProfit, &deliveryFeeCalculation, &createdAt, &userID,
		)
		if err != nil {
			continue
		}

		// 获取订单对象
		order, err := GetOrderByID(id)
		if err != nil || order == nil {
			continue
		}

		// 计算订单利润
		var profit float64
		if orderProfit.Valid {
			profit = orderProfit.Float64
		} else if order.OrderProfit != nil {
			profit = *order.OrderProfit
		}

		// 获取配送费计算结果
		var deliveryFeeResult *DeliveryFeeCalculationResult
		if deliveryFeeCalculation.Valid && deliveryFeeCalculation.String != "" {
			var calcResult DeliveryFeeCalculationResult
			if err := json.Unmarshal([]byte(deliveryFeeCalculation.String), &calcResult); err == nil {
				deliveryFeeResult = &calcResult
			}
		}

		// 计算分润预览（需要调用API层的函数，但这里在model层，所以直接计算）
		// 获取用户信息
		user, err := GetMiniAppUserByID(userID)
		if err != nil || user == nil || user.SalesCode != employeeCode {
			continue
		}

		// 计算订单金额、商品成本、配送成本
		orderAmount := totalAmount
		goodsCost := goodsAmount - profit
		deliveryCost := 0.0
		if deliveryFeeResult != nil {
			deliveryCost = deliveryFeeResult.TotalPlatformCost
		}

		// 判断是否新客户
		isNewCustomer, _ := IsNewCustomerOrder(userID, id)

		// 获取当月有效订单总金额（用于计算阶梯提成）
		currentMonth := time.Now().Format("2006-01")
		monthTotalSales, _ := GetMonthlyTotalSales(employeeCode, currentMonth)

		// 计算分成
		calcResult, err := CalculateSalesCommission(
			employeeCode,
			orderAmount,
			goodsCost,
			deliveryCost,
			isNewCustomer,
			monthTotalSales,
		)
		if err == nil && calcResult != nil {
			unaccountedAmount += calcResult.TotalCommission
			unaccountedCount++
		}
	}

	overview.UnaccountedAmount = unaccountedAmount
	overview.UnaccountedCount = unaccountedCount

	// 总金额 = 已收款订单的总分成 + 未收款订单的分润预览总和
	overview.TotalAmount += unaccountedAmount

	return &overview, nil
}

