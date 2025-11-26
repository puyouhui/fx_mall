package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"` // delivery_fee 或 amount
	DiscountValue float64  `json:"discount_value"`
	MinAmount    float64   `json:"min_amount"`
	CategoryIDs  []int     `json:"category_ids"`
	TotalCount   int       `json:"total_count"`
	UsedCount    int       `json:"used_count"`
	Status       int       `json:"status"`
	ValidFrom    LocalTime `json:"valid_from"`
	ValidTo      LocalTime `json:"valid_to"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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

// CouponWithStats 带统计信息的优惠券
type CouponWithStats struct {
	Coupon
	IssuedCount int `json:"issued_count"` // 已发放数量
	UsedCount   int `json:"used_count"`   // 已使用数量
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

	// 检查用户是否已使用过该优惠券
	var existingStatus string
	err = tx.QueryRow("SELECT status FROM user_coupons WHERE user_id = ? AND coupon_id = ?", userID, couponID).Scan(&existingStatus)
	if err == nil && existingStatus == "used" {
		return fmt.Errorf("该优惠券已使用")
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


