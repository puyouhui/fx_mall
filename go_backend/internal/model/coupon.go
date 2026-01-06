package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go_backend/internal/database"
)

// LocalTime 本地时间类型，序列化为 YYYY-MM-DD HH:mm:ss 格式
type LocalTime time.Time

// MarshalJSON 自定义序列化
func (t LocalTime) MarshalJSON() ([]byte, error) {
	tm := time.Time(t)
	// 转换为本地时区后再格式化
	localTime := tm.In(time.Local)
	formatted := localTime.Format("2006-01-02 15:04:05")
	return json.Marshal(formatted)
}

// UnmarshalJSON 自定义反序列化
func (t *LocalTime) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	// 尝试解析多种格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05Z07:00",
		time.RFC3339,
	}
	for _, format := range formats {
		if tm, err := time.Parse(format, str); err == nil {
			*t = LocalTime(tm)
			return nil
		}
	}
	return fmt.Errorf("无法解析时间格式: %s", str)
}

// ToTime 转换为 time.Time
func (t LocalTime) ToTime() time.Time {
	return time.Time(t)
}

// FromTime 从 time.Time 创建 LocalTime
func FromTime(t time.Time) LocalTime {
	return LocalTime(t)
}

// Coupon 优惠券模型
type Coupon struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"` // delivery_fee 或 amount
	DiscountValue float64   `json:"discount_value"`
	MinAmount     float64   `json:"min_amount"`
	CategoryIDs   []int     `json:"category_ids"`
	TotalCount    int       `json:"total_count"`
	UsedCount     int       `json:"used_count"`
	Status        int       `json:"status"`
	ValidFrom     LocalTime `json:"valid_from"`
	ValidTo       LocalTime `json:"valid_to"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserCoupon 用户优惠券关联
type UserCoupon struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	CouponID  int        `json:"coupon_id"`
	Status    string     `json:"status"` // unused, used, expired
	UsedAt    *time.Time `json:"used_at"`
	OrderID   *int       `json:"order_id"`
	ExpiresAt *time.Time `json:"expires_at"` // 有效期（发放时设置）
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Coupon    *Coupon    `json:"coupon,omitempty"` // 关联的优惠券信息
}

// CouponIssueLog 优惠券发放记录（管理员 / 员工）
type CouponIssueLog struct {
	ID           int        `json:"id"`
	UserID       int        `json:"user_id"`              // 客户ID
	UserName     string     `json:"user_name,omitempty"`  // 客户名称
	UserCode     string     `json:"user_code,omitempty"`  // 客户编号
	CouponID     int        `json:"coupon_id"`            // 优惠券ID
	CouponName   string     `json:"coupon_name"`          // 优惠券名称快照
	Quantity     int        `json:"quantity"`             // 发放数量
	Reason       string     `json:"reason"`               // 发放原因
	OperatorType string     `json:"operator_type"`        // admin / employee
	OperatorID   int        `json:"operator_id"`          // 操作人ID
	OperatorName string     `json:"operator_name"`        // 操作人名称
	CreatedAt    time.Time  `json:"created_at"`           // 发放时间
	ExpiresAt    *time.Time `json:"expires_at,omitempty"` // 到期时间（如有）
}

// CreateCouponIssueLog 创建一条优惠券发放记录
func CreateCouponIssueLog(log *CouponIssueLog) error {
	query := `
		INSERT INTO coupon_issue_logs
			(user_id, coupon_id, coupon_name, quantity, reason, operator_type, operator_id, operator_name, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`

	_, err := database.DB.Exec(
		query,
		log.UserID,
		log.CouponID,
		log.CouponName,
		log.Quantity,
		strings.TrimSpace(log.Reason),
		log.OperatorType,
		log.OperatorID,
		log.OperatorName,
		log.ExpiresAt,
	)
	return err
}

// GetCouponIssueLogs 获取优惠券发放记录（分页，可按优惠券/用户/操作人搜索）
func GetCouponIssueLogs(pageNum, pageSize int, keyword string, couponID int) ([]CouponIssueLog, int, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	where := "1=1"
	args := []interface{}{}

	if couponID > 0 {
		where += " AND coupon_id = ?"
		args = append(args, couponID)
	}

	if keyword != "" {
		kw := "%" + strings.TrimSpace(keyword) + "%"
		where += " AND (coupon_name LIKE ? OR operator_name LIKE ? OR reason LIKE ?)"
		args = append(args, kw, kw, kw)
	}

	// 统计总数
	countQuery := "SELECT COUNT(*) FROM coupon_issue_logs WHERE " + where
	var total int
	if err := database.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNum - 1) * pageSize
	query := `
		SELECT l.id, l.user_id, l.coupon_id, l.coupon_name, l.quantity, l.reason,
		       l.operator_type, l.operator_id, l.operator_name, l.expires_at, l.created_at,
		       u.name AS user_name, u.user_code
		FROM coupon_issue_logs l
		LEFT JOIN mini_app_users u ON l.user_id = u.id
		WHERE ` + where + `
		ORDER BY l.created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := make([]CouponIssueLog, 0)
	for rows.Next() {
		var l CouponIssueLog
		var expiresAt sql.NullTime
		var userName, userCode sql.NullString
		if err := rows.Scan(
			&l.ID,
			&l.UserID,
			&l.CouponID,
			&l.CouponName,
			&l.Quantity,
			&l.Reason,
			&l.OperatorType,
			&l.OperatorID,
			&l.OperatorName,
			&expiresAt,
			&l.CreatedAt,
			&userName,
			&userCode,
		); err != nil {
			return nil, 0, err
		}
		if expiresAt.Valid {
			t := expiresAt.Time
			l.ExpiresAt = &t
		}
		if userName.Valid {
			l.UserName = userName.String
		}
		if userCode.Valid {
			l.UserCode = userCode.String
		}
		logs = append(logs, l)
	}

	return logs, total, nil
}

// CouponUsageLog 优惠券使用记录
type CouponUsageLog struct {
	ID            int       `json:"id"`
	UserCouponID  int       `json:"user_coupon_id"`
	UserID        int       `json:"user_id"`
	UserName      string    `json:"user_name"`
	UserPhone     string    `json:"user_phone"`
	CouponID      int       `json:"coupon_id"`
	CouponName    string    `json:"coupon_name"`
	CouponType    string    `json:"coupon_type"`
	DiscountValue float64   `json:"discount_value"`
	OrderID       int       `json:"order_id"`
	OrderNumber   string    `json:"order_number"`
	UsedAt        time.Time `json:"used_at"`
	CreatedAt     time.Time `json:"created_at"` // 发放时间
}

// GetCouponUsageLogs 获取优惠券使用记录（分页，可按优惠券/用户/订单搜索）
func GetCouponUsageLogs(pageNum, pageSize int, keyword string, couponID int) ([]CouponUsageLog, int, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (pageNum - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	// 构建WHERE条件
	where := "uc.status = 'used' AND uc.used_at IS NOT NULL"
	args := []interface{}{}

	if couponID > 0 {
		where += " AND uc.coupon_id = ?"
		args = append(args, couponID)
	}

	if keyword != "" {
		where += " AND (u.name LIKE ? OR u.phone LIKE ? OR o.order_number LIKE ?)"
		keywordPattern := "%" + keyword + "%"
		args = append(args, keywordPattern, keywordPattern, keywordPattern)
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM user_coupons uc " +
		"LEFT JOIN mini_app_users u ON uc.user_id = u.id " +
		"LEFT JOIN orders o ON uc.order_id = o.id " +
		"WHERE " + where

	var total int
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 查询列表
	query := `
		SELECT 
			uc.id, uc.user_id, uc.coupon_id, uc.order_id, uc.used_at, uc.created_at,
			u.name, u.phone,
			c.name, c.type, c.discount_value,
			o.order_number
		FROM user_coupons uc
		LEFT JOIN mini_app_users u ON uc.user_id = u.id
		LEFT JOIN coupons c ON uc.coupon_id = c.id
		LEFT JOIN orders o ON uc.order_id = o.id
		WHERE ` + where + `
		ORDER BY uc.used_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询使用记录失败: %w", err)
	}
	defer rows.Close()

	logs := make([]CouponUsageLog, 0)
	for rows.Next() {
		var l CouponUsageLog
		var userName, userPhone, couponName, couponType, orderNumber sql.NullString
		var discountValue sql.NullFloat64
		var orderID sql.NullInt64

		err := rows.Scan(
			&l.UserCouponID, &l.UserID, &l.CouponID, &orderID, &l.UsedAt, &l.CreatedAt,
			&userName, &userPhone,
			&couponName, &couponType, &discountValue,
			&orderNumber,
		)
		if err != nil {
			continue
		}

		l.ID = l.UserCouponID
		if userName.Valid {
			l.UserName = userName.String
		}
		if userPhone.Valid {
			l.UserPhone = userPhone.String
		}
		if couponName.Valid {
			l.CouponName = couponName.String
		}
		if couponType.Valid {
			l.CouponType = couponType.String
		}
		if discountValue.Valid {
			l.DiscountValue = discountValue.Float64
		}
		if orderID.Valid {
			l.OrderID = int(orderID.Int64)
		}
		if orderNumber.Valid {
			l.OrderNumber = orderNumber.String
		}

		logs = append(logs, l)
	}

	return logs, total, nil
}

// CouponWithStats 带统计信息的优惠券
type CouponWithStats struct {
	Coupon
	IssuedCount int  `json:"issued_count"` // 已发放数量
	UsedCount   int  `json:"used_count"`   // 已使用数量
	IsValid     bool `json:"is_valid"`     // 是否在有效期内（当前时间在valid_from和valid_to之间）
	IsExpired   bool `json:"is_expired"`   // 是否已过期（当前时间晚于valid_to）
}

// GetAllCoupons 获取所有优惠券（后台管理，包含统计信息）
func GetAllCoupons() ([]CouponWithStats, error) {
	var coupons []CouponWithStats
	query := `
		SELECT c.id, c.name, c.type, c.discount_value, c.min_amount, c.category_ids, c.total_count, c.used_count, c.status, c.valid_from, c.valid_to, c.description, c.created_at, c.updated_at,
		       COALESCE(COUNT(DISTINCT uc.id), 0) as issued_count,
		       COALESCE(SUM(CASE WHEN uc.status = 'used' THEN 1 ELSE 0 END), 0) as actual_used_count
		FROM coupons c
		LEFT JOIN user_coupons uc ON c.id = uc.coupon_id
		GROUP BY c.id
		ORDER BY c.created_at DESC
	`
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var coupon CouponWithStats
		var categoryIDsJSON sql.NullString
		var validFrom, validTo time.Time
		var actualUsedCount int
		if err := rows.Scan(&coupon.ID, &coupon.Name, &coupon.Type, &coupon.DiscountValue, &coupon.MinAmount, &categoryIDsJSON, &coupon.TotalCount, &coupon.UsedCount, &coupon.Status, &validFrom, &validTo, &coupon.Description, &coupon.CreatedAt, &coupon.UpdatedAt, &coupon.IssuedCount, &actualUsedCount); err != nil {
			return nil, err
		}

		// 使用实际统计的已使用数量
		coupon.UsedCount = actualUsedCount

		// 确保时间使用本地时区
		coupon.ValidFrom = LocalTime(validFrom.In(time.Local))
		coupon.ValidTo = LocalTime(validTo.In(time.Local))

		// 解析分类ID JSON
		if categoryIDsJSON.Valid && categoryIDsJSON.String != "" {
			if err := json.Unmarshal([]byte(categoryIDsJSON.String), &coupon.CategoryIDs); err != nil {
				coupon.CategoryIDs = []int{}
			}
		} else {
			coupon.CategoryIDs = []int{}
		}

		// 计算有效期状态
		now := time.Now()
		validFromTime := coupon.ValidFrom.ToTime()
		validToTime := coupon.ValidTo.ToTime()
		coupon.IsValid = now.After(validFromTime) && now.Before(validToTime)
		coupon.IsExpired = now.After(validToTime)

		coupons = append(coupons, coupon)
	}

	return coupons, nil
}

// GetCouponByID 根据ID获取优惠券
func GetCouponByID(id int) (*Coupon, error) {
	var coupon Coupon
	var categoryIDsJSON sql.NullString
	var validFrom, validTo time.Time
	query := "SELECT id, name, type, discount_value, min_amount, category_ids, total_count, used_count, status, valid_from, valid_to, description, created_at, updated_at FROM coupons WHERE id = ?"
	err := database.DB.QueryRow(query, id).Scan(&coupon.ID, &coupon.Name, &coupon.Type, &coupon.DiscountValue, &coupon.MinAmount, &categoryIDsJSON, &coupon.TotalCount, &coupon.UsedCount, &coupon.Status, &validFrom, &validTo, &coupon.Description, &coupon.CreatedAt, &coupon.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	coupon.ValidFrom = LocalTime(validFrom)
	coupon.ValidTo = LocalTime(validTo)

	// 解析分类ID JSON
	if categoryIDsJSON.Valid && categoryIDsJSON.String != "" {
		if err := json.Unmarshal([]byte(categoryIDsJSON.String), &coupon.CategoryIDs); err != nil {
			coupon.CategoryIDs = []int{}
		}
	} else {
		coupon.CategoryIDs = []int{}
	}

	return &coupon, nil
}

// CreateCoupon 创建优惠券
func CreateCoupon(coupon *Coupon) error {
	categoryIDsJSON, err := json.Marshal(coupon.CategoryIDs)
	if err != nil {
		return err
	}

	query := "INSERT INTO coupons (name, type, discount_value, min_amount, category_ids, total_count, used_count, status, valid_from, valid_to, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?, NOW(), NOW())"
	result, err := database.DB.Exec(query, coupon.Name, coupon.Type, coupon.DiscountValue, coupon.MinAmount, string(categoryIDsJSON), coupon.TotalCount, coupon.Status, coupon.ValidFrom.ToTime(), coupon.ValidTo.ToTime(), coupon.Description)
	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	coupon.ID = int(lastID)
	return nil
}

// UpdateCoupon 更新优惠券
func UpdateCoupon(coupon *Coupon) error {
	categoryIDsJSON, err := json.Marshal(coupon.CategoryIDs)
	if err != nil {
		return err
	}

	query := "UPDATE coupons SET name = ?, type = ?, discount_value = ?, min_amount = ?, category_ids = ?, total_count = ?, status = ?, valid_from = ?, valid_to = ?, description = ?, updated_at = NOW() WHERE id = ?"
	_, err = database.DB.Exec(query, coupon.Name, coupon.Type, coupon.DiscountValue, coupon.MinAmount, string(categoryIDsJSON), coupon.TotalCount, coupon.Status, coupon.ValidFrom.ToTime(), coupon.ValidTo.ToTime(), coupon.Description, coupon.ID)
	return err
}

// DeleteCoupon 删除优惠券
func DeleteCoupon(id int) error {
	query := "DELETE FROM coupons WHERE id = ?"
	_, err := database.DB.Exec(query, id)
	return err
}

// GetAvailableCouponsForUser 获取用户可用的优惠券列表
func GetAvailableCouponsForUser(userID int, orderAmount float64, categoryIDs []int) ([]Coupon, error) {
	now := time.Now()
	query := `
		SELECT c.id, c.name, c.type, c.discount_value, c.min_amount, c.category_ids, c.total_count, c.used_count, c.status, c.valid_from, c.valid_to, c.description, c.created_at, c.updated_at
		FROM coupons c
		WHERE c.status = 1
		  AND c.valid_from <= ?
		  AND c.valid_to >= ?
		  AND (c.total_count = 0 OR c.used_count < c.total_count)
		  AND c.min_amount <= ?
		  AND NOT EXISTS (
		      SELECT 1 FROM user_coupons uc
		      WHERE uc.user_id = ? AND uc.coupon_id = c.id AND uc.status = 'used'
		  )
		ORDER BY c.discount_value DESC, c.created_at DESC
	`
	rows, err := database.DB.Query(query, now, now, orderAmount, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coupons []Coupon
	for rows.Next() {
		var coupon Coupon
		var categoryIDsJSON sql.NullString
		var validFrom, validTo time.Time
		if err := rows.Scan(&coupon.ID, &coupon.Name, &coupon.Type, &coupon.DiscountValue, &coupon.MinAmount, &categoryIDsJSON, &coupon.TotalCount, &coupon.UsedCount, &coupon.Status, &validFrom, &validTo, &coupon.Description, &coupon.CreatedAt, &coupon.UpdatedAt); err != nil {
			return nil, err
		}

		// 确保时间使用本地时区
		coupon.ValidFrom = LocalTime(validFrom.In(time.Local))
		coupon.ValidTo = LocalTime(validTo.In(time.Local))

		// 解析分类ID JSON
		if categoryIDsJSON.Valid && categoryIDsJSON.String != "" {
			if err := json.Unmarshal([]byte(categoryIDsJSON.String), &coupon.CategoryIDs); err != nil {
				coupon.CategoryIDs = []int{}
			}
		} else {
			coupon.CategoryIDs = []int{}
		}

		// 检查分类限制
		if len(coupon.CategoryIDs) > 0 {
			// 如果优惠券指定了分类，检查订单中是否有该分类的商品
			hasMatchingCategory := false
			for _, orderCatID := range categoryIDs {
				for _, couponCatID := range coupon.CategoryIDs {
					if orderCatID == couponCatID {
						hasMatchingCategory = true
						break
					}
				}
				if hasMatchingCategory {
					break
				}
			}
			if !hasMatchingCategory {
				continue // 跳过不符合分类条件的优惠券
			}
		}

		coupons = append(coupons, coupon)
	}

	return coupons, nil
}

// GetUserCoupons 获取用户的优惠券列表
func GetUserCoupons(userID int) ([]UserCoupon, error) {
	query := `
		SELECT uc.id, uc.user_id, uc.coupon_id, uc.status, uc.used_at, uc.order_id, uc.expires_at, uc.created_at, uc.updated_at,
		       c.id, c.name, c.type, c.discount_value, c.min_amount, c.category_ids, c.total_count, c.used_count, c.status, c.valid_from, c.valid_to, c.description, c.created_at, c.updated_at
		FROM user_coupons uc
		LEFT JOIN coupons c ON uc.coupon_id = c.id
		WHERE uc.user_id = ?
		ORDER BY uc.created_at DESC
	`
	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userCoupons []UserCoupon
	for rows.Next() {
		var uc UserCoupon
		var usedAt sql.NullTime
		var orderID sql.NullInt64
		var expiresAt sql.NullTime
		var categoryIDsJSON sql.NullString
		var coupon Coupon

		var validFrom, validTo time.Time
		err := rows.Scan(
			&uc.ID, &uc.UserID, &uc.CouponID, &uc.Status, &usedAt, &orderID, &expiresAt, &uc.CreatedAt, &uc.UpdatedAt,
			&coupon.ID, &coupon.Name, &coupon.Type, &coupon.DiscountValue, &coupon.MinAmount, &categoryIDsJSON, &coupon.TotalCount, &coupon.UsedCount, &coupon.Status, &validFrom, &validTo, &coupon.Description, &coupon.CreatedAt, &coupon.UpdatedAt,
		)
		if err == nil {
			coupon.ValidFrom = LocalTime(validFrom)
			coupon.ValidTo = LocalTime(validTo)
		}
		if err != nil {
			return nil, err
		}

		if usedAt.Valid {
			uc.UsedAt = &usedAt.Time
		}
		if orderID.Valid {
			orderIDInt := int(orderID.Int64)
			uc.OrderID = &orderIDInt
		}
		if expiresAt.Valid {
			uc.ExpiresAt = &expiresAt.Time
		}

		// 解析分类ID JSON
		if categoryIDsJSON.Valid && categoryIDsJSON.String != "" {
			if err := json.Unmarshal([]byte(categoryIDsJSON.String), &coupon.CategoryIDs); err != nil {
				coupon.CategoryIDs = []int{}
			}
		} else {
			coupon.CategoryIDs = []int{}
		}

		// 自动检查并更新过期状态（仅对未使用的优惠券）
		now := time.Now()
		if uc.Status == "unused" {
			isExpired := false
			// 优先检查用户优惠券的 expires_at
			if uc.ExpiresAt != nil {
				if now.After(*uc.ExpiresAt) {
					isExpired = true
				}
			} else if coupon.ID > 0 {
				// 如果没有 expires_at，检查优惠券模板的 valid_to
				if now.After(coupon.ValidTo.ToTime()) {
					isExpired = true
				}
			}

			// 如果已过期，更新状态
			if isExpired {
				_, _ = database.DB.Exec(`
					UPDATE user_coupons 
					SET status = 'expired', updated_at = NOW() 
					WHERE id = ? AND status = 'unused'
				`, uc.ID)
				uc.Status = "expired"
			}
		}

		uc.Coupon = &coupon
		userCoupons = append(userCoupons, uc)
	}

	return userCoupons, nil
}

// UseCoupon 使用优惠券
func UseCoupon(userID, couponID, orderID int) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 检查优惠券是否可用
	var coupon Coupon
	err = tx.QueryRow("SELECT id, used_count, total_count FROM coupons WHERE id = ? AND status = 1", couponID).Scan(&coupon.ID, &coupon.UsedCount, &coupon.TotalCount)
	if err != nil {
		return fmt.Errorf("优惠券不存在或已禁用")
	}

	if coupon.TotalCount > 0 && coupon.UsedCount >= coupon.TotalCount {
		return fmt.Errorf("优惠券已用完")
	}

	// 检查优惠券模板是否在有效期内
	now := time.Now()
	var validFrom, validTo time.Time
	err = tx.QueryRow("SELECT valid_from, valid_to FROM coupons WHERE id = ?", couponID).Scan(&validFrom, &validTo)
	if err != nil {
		return fmt.Errorf("获取优惠券信息失败: %w", err)
	}
	if now.Before(validFrom) || now.After(validTo) {
		return fmt.Errorf("优惠券不在有效期内")
	}

	// 检查用户优惠券记录，包括状态和有效期
	var existingStatus string
	var expiresAt sql.NullTime
	err = tx.QueryRow("SELECT status, expires_at FROM user_coupons WHERE user_id = ? AND coupon_id = ?", userID, couponID).Scan(&existingStatus, &expiresAt)
	if err == nil {
		if existingStatus == "used" {
			return fmt.Errorf("该优惠券已使用")
		}
		if existingStatus == "expired" {
			return fmt.Errorf("该优惠券已过期")
		}
		// 检查用户优惠券的有效期（expires_at）
		if expiresAt.Valid {
			if now.After(expiresAt.Time) {
				return fmt.Errorf("该优惠券已过期")
			}
		}
	}

	// 更新或插入用户优惠券记录
	if err == sql.ErrNoRows {
		// 插入新记录
		_, err = tx.Exec("INSERT INTO user_coupons (user_id, coupon_id, status, used_at, order_id, created_at, updated_at) VALUES (?, ?, 'used', NOW(), ?, NOW(), NOW())", userID, couponID, orderID)
	} else {
		// 更新现有记录
		_, err = tx.Exec("UPDATE user_coupons SET status = 'used', used_at = NOW(), order_id = ?, updated_at = NOW() WHERE user_id = ? AND coupon_id = ?", orderID, userID, couponID)
	}
	if err != nil {
		return err
	}

	// 更新优惠券使用计数
	_, err = tx.Exec("UPDATE coupons SET used_count = used_count + 1, updated_at = NOW() WHERE id = ?", couponID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// UseCouponByUserCouponIDInTx 在事务内使用优惠券（通过用户优惠券ID，更精确）
func UseCouponByUserCouponIDInTx(tx *sql.Tx, userCouponID, orderID int) error {
	// 获取用户优惠券信息
	var userID, couponID int
	var status string
	var expiresAt sql.NullTime
	err := tx.QueryRow(`
		SELECT user_id, coupon_id, status, expires_at 
		FROM user_coupons 
		WHERE id = ?
	`, userCouponID).Scan(&userID, &couponID, &status, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("用户优惠券不存在")
		}
		return err
	}

	// 检查是否已使用
	if status == "used" {
		return fmt.Errorf("该优惠券已使用")
	}

	// 检查是否过期
	now := time.Now()
	if expiresAt.Valid {
		if now.After(expiresAt.Time) {
			return fmt.Errorf("该优惠券已过期")
		}
	}

	// 检查优惠券模板是否可用
	var coupon Coupon
	var validFrom, validTo time.Time
	err = tx.QueryRow("SELECT id, used_count, total_count, valid_from, valid_to FROM coupons WHERE id = ? AND status = 1", couponID).Scan(&coupon.ID, &coupon.UsedCount, &coupon.TotalCount, &validFrom, &validTo)
	if err != nil {
		return fmt.Errorf("优惠券不存在或已禁用")
	}

	// 检查优惠券模板是否在有效期内
	if now.Before(validFrom) || now.After(validTo) {
		return fmt.Errorf("优惠券不在有效期内")
	}

	if coupon.TotalCount > 0 && coupon.UsedCount >= coupon.TotalCount {
		return fmt.Errorf("优惠券已用完")
	}

	// 更新用户优惠券记录（使用 user_coupon_id 精确更新）
	_, err = tx.Exec(`
		UPDATE user_coupons 
		SET status = 'used', used_at = NOW(), order_id = ?, updated_at = NOW() 
		WHERE id = ?
	`, orderID, userCouponID)
	if err != nil {
		return err
	}

	// 更新优惠券使用计数
	_, err = tx.Exec("UPDATE coupons SET used_count = used_count + 1, updated_at = NOW() WHERE id = ?", couponID)
	if err != nil {
		return err
	}

	return nil
}

// UseCouponByUserCouponID 使用优惠券（通过用户优惠券ID，更精确）
// 注意：此函数会创建新事务，如果需要在订单创建事务内处理，请使用 UseCouponByUserCouponIDInTx
func UseCouponByUserCouponID(userCouponID, orderID int) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := UseCouponByUserCouponIDInTx(tx, userCouponID, orderID); err != nil {
		return err
	}

	return tx.Commit()
}

// IssueCouponToUser 发放优惠券给用户（管理员操作，支持数量和有效期）
func IssueCouponToUser(userID, couponID, quantity int, expiresAt *time.Time) error {
	if quantity <= 0 {
		return fmt.Errorf("发放数量必须大于0")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 检查优惠券是否存在且启用
	var coupon Coupon
	var categoryIDsJSON sql.NullString
	var validFrom, validTo time.Time
	err = tx.QueryRow(`
		SELECT id, name, type, discount_value, min_amount, category_ids, total_count, used_count, status, valid_from, valid_to, description
		FROM coupons WHERE id = ? AND status = 1
	`, couponID).Scan(
		&coupon.ID, &coupon.Name, &coupon.Type, &coupon.DiscountValue, &coupon.MinAmount, &categoryIDsJSON,
		&coupon.TotalCount, &coupon.UsedCount, &coupon.Status, &validFrom, &validTo, &coupon.Description,
	)
	if err == nil {
		// 确保时间使用本地时区
		coupon.ValidFrom = LocalTime(validFrom.In(time.Local))
		coupon.ValidTo = LocalTime(validTo.In(time.Local))
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("优惠券不存在或已禁用")
		}
		return fmt.Errorf("查询优惠券失败: %w", err)
	}

	// 检查优惠券是否在有效期内
	now := time.Now()
	if now.Before(coupon.ValidFrom.ToTime()) || now.After(coupon.ValidTo.ToTime()) {
		return fmt.Errorf("优惠券不在有效期内")
	}

	// 检查是否还有剩余数量（通过统计已发放数量）
	if coupon.TotalCount > 0 {
		var issuedCount int
		err = tx.QueryRow("SELECT COUNT(*) FROM user_coupons WHERE coupon_id = ?", couponID).Scan(&issuedCount)
		if err != nil {
			return fmt.Errorf("统计已发放数量失败: %w", err)
		}
		if issuedCount+quantity > coupon.TotalCount {
			return fmt.Errorf("优惠券剩余数量不足，当前剩余：%d 张", coupon.TotalCount-issuedCount)
		}
	}

	// 检查用户已拥有该优惠券的数量
	var existingCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM user_coupons WHERE user_id = ? AND coupon_id = ?", userID, couponID).Scan(&existingCount)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("查询用户优惠券失败: %w", err)
	}

	// 如果指定了过期时间，验证过期时间
	if expiresAt != nil {
		// 过期时间不能早于当前时间
		if expiresAt.Before(now) {
			return fmt.Errorf("过期时间不能早于当前时间")
		}
		// 过期时间不能晚于优惠券模板的有效期结束时间
		if expiresAt.After(coupon.ValidTo.ToTime()) {
			return fmt.Errorf("过期时间不能晚于优惠券模板的有效期结束时间（%s）", coupon.ValidTo.ToTime().Format("2006-01-02 15:04:05"))
		}
		// 过期时间不能早于优惠券模板的有效期开始时间
		if expiresAt.Before(coupon.ValidFrom.ToTime()) {
			return fmt.Errorf("过期时间不能早于优惠券模板的有效期开始时间（%s）", coupon.ValidFrom.ToTime().Format("2006-01-02 15:04:05"))
		}
	}

	// 插入用户优惠券记录（支持多张）
	for i := 0; i < quantity; i++ {
		if expiresAt != nil {
			_, err = tx.Exec(`
				INSERT INTO user_coupons (user_id, coupon_id, status, expires_at, created_at, updated_at)
				VALUES (?, ?, 'unused', ?, NOW(), NOW())
			`, userID, couponID, expiresAt)
		} else {
			_, err = tx.Exec(`
				INSERT INTO user_coupons (user_id, coupon_id, status, created_at, updated_at)
				VALUES (?, ?, 'unused', NOW(), NOW())
			`, userID, couponID)
		}
		if err != nil {
			return fmt.Errorf("发放优惠券失败: %w", err)
		}
	}

	return tx.Commit()
}
