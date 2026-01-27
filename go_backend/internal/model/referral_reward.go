package model

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"go_backend/internal/database"
)

// ReferralRewardConfig 推荐奖励活动配置
type ReferralRewardConfig struct {
	ID          int       `json:"id"`
	IsEnabled   bool      `json:"is_enabled"`   // 是否启用
	RewardType  string    `json:"reward_type"`   // 奖励类型：points/coupon/amount
	RewardValue float64   `json:"reward_value"` // 奖励值
	CouponID    *int      `json:"coupon_id,omitempty"` // 优惠券ID（当reward_type为coupon时使用）
	Description string    `json:"description"`   // 活动说明
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ReferralReward 推荐奖励记录
type ReferralReward struct {
	ID          int        `json:"id"`
	ReferrerID  int        `json:"referrer_id"`  // 推荐人用户ID（老用户）
	NewUserID   int        `json:"new_user_id"`  // 新用户ID
	OrderID     int        `json:"order_id"`     // 订单ID
	OrderNumber string     `json:"order_number"`  // 订单编号
	RewardType  string     `json:"reward_type"`  // 奖励类型
	RewardValue float64    `json:"reward_value"` // 奖励值
	CouponID    *int       `json:"coupon_id,omitempty"` // 优惠券ID
	Status      string     `json:"status"`        // pending/completed/failed
	RewardAt    *time.Time `json:"reward_at,omitempty"` // 奖励发放时间
	Remark      string     `json:"remark"`       // 备注
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// GetReferralRewardConfig 获取推荐奖励活动配置
func GetReferralRewardConfig() (*ReferralRewardConfig, error) {
	var config ReferralRewardConfig
	var couponID sql.NullInt64

	query := `
		SELECT id, is_enabled, reward_type, reward_value, coupon_id, description, created_at, updated_at
		FROM referral_reward_config
		ORDER BY id DESC
		LIMIT 1
	`

	err := database.DB.QueryRow(query).Scan(
		&config.ID,
		&config.IsEnabled,
		&config.RewardType,
		&config.RewardValue,
		&couponID,
		&config.Description,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if couponID.Valid {
		couponIDInt := int(couponID.Int64)
		config.CouponID = &couponIDInt
	}

	return &config, nil
}

// CreateReferralRewardConfig 创建推荐奖励活动配置
func CreateReferralRewardConfig(config *ReferralRewardConfig) error {
	var couponID interface{}
	if config.CouponID != nil {
		couponID = *config.CouponID
	} else {
		couponID = nil
	}

	query := `
		INSERT INTO referral_reward_config (is_enabled, reward_type, reward_value, coupon_id, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`

	result, err := database.DB.Exec(
		query,
		boolToTinyInt(config.IsEnabled),
		config.RewardType,
		config.RewardValue,
		couponID,
		config.Description,
	)

	if err != nil {
		return err
	}

	// 获取插入的 ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	config.ID = int(id)
	return nil
}

// UpdateReferralRewardConfig 更新推荐奖励活动配置
func UpdateReferralRewardConfig(config *ReferralRewardConfig) error {
	var couponID interface{}
	if config.CouponID != nil {
		couponID = *config.CouponID
	} else {
		couponID = nil
	}

	query := `
		UPDATE referral_reward_config
		SET is_enabled = ?, reward_type = ?, reward_value = ?, coupon_id = ?, description = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := database.DB.Exec(
		query,
		boolToTinyInt(config.IsEnabled),
		config.RewardType,
		config.RewardValue,
		couponID,
		config.Description,
		config.ID,
	)

	return err
}

// CreateReferralReward 创建推荐奖励记录（待发放状态）
func CreateReferralReward(referrerID, newUserID, orderID int, orderNumber string) error {
	// 检查推荐人是否是销售员，如果是销售员则不给予推荐奖励
	referrer, err := GetMiniAppUserByID(referrerID)
	if err != nil || referrer == nil {
		// 获取推荐人信息失败，记录日志但不影响主流程
		log.Printf("获取推荐人信息失败 (referrerID: %d): %v", referrerID, err)
		return nil
	}
	if referrer.IsSalesEmployee {
		// 推荐人是销售员，不给予推荐奖励（销售员已有销售提成）
		log.Printf("推荐人是销售员 (referrerID: %d)，不给予推荐奖励", referrerID)
		return nil
	}

	// 检查是否已存在该订单的奖励记录
	var count int
	checkQuery := `
		SELECT COUNT(*) FROM referral_rewards
		WHERE order_id = ? AND referrer_id = ?
	`
	if err := database.DB.QueryRow(checkQuery, orderID, referrerID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		// 已存在，不重复创建
		return nil
	}

	// 优先使用新奖励活动配置表 reward_activities 中的“拉新活动”（activity_type = referral）
	if activity, err := GetEnabledRewardActivityByType("referral"); err != nil {
		return err
	} else if activity != nil {
		// 根据奖励类型创建推荐奖励记录
		switch activity.RewardType {
		case "points":
			// 积分：一条记录，reward_value 为积分数量
			query := `
				INSERT INTO referral_rewards (
					referrer_id, new_user_id, order_id, order_number,
					reward_type, reward_value, coupon_id, status, created_at, updated_at
				) VALUES (?, ?, ?, ?, 'points', ?, NULL, 'pending', NOW(), NOW())
			`
			_, err = database.DB.Exec(
				query,
				referrerID,
				newUserID,
				orderID,
				orderNumber,
				activity.RewardValue,
			)
			return err
		case "coupon":
			// 优惠券：支持多张券，为每张券各创建一条记录
			if len(activity.CouponIDs) == 0 {
				// 未配置具体券ID，则不创建记录
				return nil
			}
			query := `
				INSERT INTO referral_rewards (
					referrer_id, new_user_id, order_id, order_number,
					reward_type, reward_value, coupon_id, status, created_at, updated_at
				) VALUES (?, ?, ?, ?, 'coupon', 0, ?, 'pending', NOW(), NOW())
			`
			for _, cid := range activity.CouponIDs {
				if cid <= 0 {
					continue
				}
				if _, err = database.DB.Exec(
					query,
					referrerID,
					newUserID,
					orderID,
					orderNumber,
					cid,
				); err != nil {
					return err
				}
			}
			return nil
		case "amount":
			// 金额：只做记录，不真正发放
			query := `
				INSERT INTO referral_rewards (
					referrer_id, new_user_id, order_id, order_number,
					reward_type, reward_value, coupon_id, status, created_at, updated_at
				) VALUES (?, ?, ?, ?, 'amount', ?, NULL, 'pending', NOW(), NOW())
			`
			_, err = database.DB.Exec(
				query,
				referrerID,
				newUserID,
				orderID,
				orderNumber,
				activity.RewardValue,
			)
			return err
		default:
			// 未知类型，直接忽略
			return nil
		}
	}

	// 兼容旧系统：如果没有配置新的 reward_activities，则继续使用 referral_reward_config
	config, err := GetReferralRewardConfig()
	if err != nil {
		return err
	}
	if config == nil || !config.IsEnabled {
		// 活动未启用，不创建奖励记录
		return nil
	}

	var couponID interface{}
	if config.CouponID != nil {
		couponID = *config.CouponID
	} else {
		couponID = nil
	}

	query := `
		INSERT INTO referral_rewards (
			referrer_id, new_user_id, order_id, order_number,
			reward_type, reward_value, coupon_id, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', NOW(), NOW())
	`

	_, err = database.DB.Exec(
		query,
		referrerID,
		newUserID,
		orderID,
		orderNumber,
		config.RewardType,
		config.RewardValue,
		couponID,
	)

	return err
}

// ProcessReferralReward 处理推荐奖励（订单完成付款后调用）
func ProcessReferralReward(orderID int) error {
	// 获取该订单的所有待发放奖励记录
	query := `
		SELECT id, referrer_id, new_user_id, order_number, reward_type, reward_value, coupon_id
		FROM referral_rewards
		WHERE order_id = ? AND status = 'pending'
	`

	rows, err := database.DB.Query(query, orderID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var rewards []ReferralReward
	for rows.Next() {
		var reward ReferralReward
		var couponID sql.NullInt64

		err := rows.Scan(
			&reward.ID,
			&reward.ReferrerID,
			&reward.NewUserID,
			&reward.OrderNumber,
			&reward.RewardType,
			&reward.RewardValue,
			&couponID,
		)
		if err != nil {
			log.Printf("扫描推荐奖励记录失败: %v", err)
			continue
		}

		if couponID.Valid {
			couponIDInt := int(couponID.Int64)
			reward.CouponID = &couponIDInt
		}

		rewards = append(rewards, reward)
	}

	// 处理每个奖励
	for _, reward := range rewards {
		if err := issueReferralReward(&reward); err != nil {
			log.Printf("发放推荐奖励失败 (ID: %d): %v", reward.ID, err)
			// 更新状态为失败
			updateQuery := `
				UPDATE referral_rewards
				SET status = 'failed', remark = ?, updated_at = NOW()
				WHERE id = ?
			`
			_, _ = database.DB.Exec(updateQuery, err.Error(), reward.ID)
		} else {
			// 更新状态为已完成
			updateQuery := `
				UPDATE referral_rewards
				SET status = 'completed', reward_at = NOW(), updated_at = NOW()
				WHERE id = ?
			`
			_, _ = database.DB.Exec(updateQuery, reward.ID)
			log.Printf("成功发放推荐奖励 (ID: %d, 推荐人: %d, 奖励类型: %s, 奖励值: %.2f)",
				reward.ID, reward.ReferrerID, reward.RewardType, reward.RewardValue)
		}
	}

	return nil
}

// issueReferralReward 发放推荐奖励
func issueReferralReward(reward *ReferralReward) error {
	switch reward.RewardType {
	case "points":
		// 发放积分
		return addUserPoints(reward.ReferrerID, int(reward.RewardValue), fmt.Sprintf("推荐新用户首次下单奖励（订单：%s）", reward.OrderNumber))
	case "coupon":
		// 发放优惠券
		if reward.CouponID == nil {
			return fmt.Errorf("优惠券ID为空")
		}
		// 记录为活动奖励，显示在优惠券管理-发放记录中
		return issueCouponToUser(reward.ReferrerID, *reward.CouponID, 1, "活动奖励")
	case "amount":
		// 发放金额（可以记录到用户余额或通过其他方式发放）
		// 这里暂时记录到备注中，实际发放逻辑可以根据业务需求实现
		return nil
	default:
		return fmt.Errorf("未知的奖励类型: %s", reward.RewardType)
	}
}

// addUserPoints 给用户添加积分
func addUserPoints(userID int, points int, reason string) error {
	// 使用统一的积分添加方法，会自动记录积分明细
	pointsType := "referral_reward"
	return AddPoints(userID, points, pointsType, nil, nil, reason)
}

// LogActivityRewardForUser 将某个用户获得的活动奖励同步到推荐奖励记录表（referral_rewards）
// 这里统一作为“受奖人 = referrer_id”，new_user_id / order 相关字段为0或空，用 remark 标识来源
func LogActivityRewardForUser(userID int, rewardType string, rewardValue float64, couponID *int, remark string) error {
	var couponIDValue interface{}
	if couponID != nil {
		couponIDValue = *couponID
	} else {
		couponIDValue = nil
	}

	query := `
		INSERT INTO referral_rewards (
			referrer_id, new_user_id, order_id, order_number,
			reward_type, reward_value, coupon_id, status, reward_at, remark, created_at, updated_at
		) VALUES (?, 0, 0, '', ?, ?, ?, 'completed', NOW(), ?, NOW(), NOW())
	`

	_, err := database.DB.Exec(
		query,
		userID,
		rewardType,
		rewardValue,
		couponIDValue,
		remark,
	)
	return err
}

// issueCouponToUser 给用户发放优惠券
func issueCouponToUser(userID, couponID, quantity int, reason string) error {
	// 检查用户是否存在
	user, err := GetMiniAppUserByID(userID)
	if err != nil || user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查优惠券是否存在
	coupon, err := GetCouponByID(couponID)
	if err != nil || coupon == nil {
		return fmt.Errorf("优惠券不存在")
	}

	// 计算过期时间
	// 拉新奖励场景下，优惠券默认有效期为发放起一年
	expireTime := time.Now().AddDate(1, 0, 0)

	// 发放优惠券
	for i := 0; i < quantity; i++ {
		query := `
			INSERT INTO user_coupons (user_id, coupon_id, status, expires_at, created_at, updated_at)
			VALUES (?, ?, 'unused', ?, NOW(), NOW())
		`
		_, err = database.DB.Exec(query, userID, couponID, expireTime)
		if err != nil {
			return err
		}
	}

	// 记录发放日志
	issueLogQuery := `
		INSERT INTO coupon_issue_logs (
			user_id, coupon_id, coupon_name, quantity, reason,
			operator_type, operator_id, operator_name, expires_at, created_at
		) VALUES (?, ?, ?, ?, ?, 'system', 0, '系统', ?, NOW())
	`
	_, err = database.DB.Exec(issueLogQuery, userID, couponID, coupon.Name, quantity, reason, expireTime)
	if err != nil {
		log.Printf("记录优惠券发放日志失败: %v", err)
		// 不影响主流程，只记录日志
	}

	return nil
}

// CheckIsFirstOrder 检查用户是否是首次下单
func CheckIsFirstOrder(userID int) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM orders
		WHERE user_id = ? AND status != 'cancelled'
	`
	err := database.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// GetReferralRewards 获取推荐奖励记录列表（分页）
func GetReferralRewards(pageNum, pageSize int, referrerID, newUserID *int, status string) ([]map[string]interface{}, int, error) {
	offset := (pageNum - 1) * pageSize

	// 构建查询条件
	whereClause := "1=1"
	args := []interface{}{}

	if referrerID != nil {
		whereClause += " AND referrer_id = ?"
		args = append(args, *referrerID)
	}

	if newUserID != nil {
		whereClause += " AND new_user_id = ?"
		args = append(args, *newUserID)
	}

	if status != "" {
		whereClause += " AND status = ?"
		args = append(args, status)
	}

	// 查询总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM referral_rewards WHERE %s", whereClause)
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	query := fmt.Sprintf(`
		SELECT 
			r.id, r.referrer_id, r.new_user_id, r.order_id, r.order_number,
			r.reward_type, r.reward_value, r.coupon_id, r.status, r.reward_at, r.remark,
			r.created_at, r.updated_at,
			u1.name as referrer_name, u1.user_code as referrer_code,
			u2.name as new_user_name, u2.user_code as new_user_code
		FROM referral_rewards r
		LEFT JOIN mini_app_users u1 ON r.referrer_id = u1.id
		LEFT JOIN mini_app_users u2 ON r.new_user_id = u2.id
		WHERE %s
		ORDER BY r.created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, pageSize, offset)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rewards []map[string]interface{}
	for rows.Next() {
		var reward ReferralReward
		var couponID sql.NullInt64
		var rewardAt sql.NullTime
		var referrerName, referrerCode, newUserName, newUserCode sql.NullString

		err := rows.Scan(
			&reward.ID,
			&reward.ReferrerID,
			&reward.NewUserID,
			&reward.OrderID,
			&reward.OrderNumber,
			&reward.RewardType,
			&reward.RewardValue,
			&couponID,
			&reward.Status,
			&rewardAt,
			&reward.Remark,
			&reward.CreatedAt,
			&reward.UpdatedAt,
			&referrerName,
			&referrerCode,
			&newUserName,
			&newUserCode,
		)
		if err != nil {
			log.Printf("扫描推荐奖励记录失败: %v", err)
			continue
		}

		rewardData := map[string]interface{}{
			"id":           reward.ID,
			"referrer_id":  reward.ReferrerID,
			"new_user_id":  reward.NewUserID,
			"order_id":     reward.OrderID,
			"order_number": reward.OrderNumber,
			"reward_type":  reward.RewardType,
			"reward_value": reward.RewardValue,
			"status":       reward.Status,
			"remark":       reward.Remark,
			"created_at":   reward.CreatedAt,
			"updated_at":   reward.UpdatedAt,
		}

		if couponID.Valid {
			rewardData["coupon_id"] = couponID.Int64
		}

		if rewardAt.Valid {
			rewardData["reward_at"] = rewardAt.Time
		}

		if referrerName.Valid {
			rewardData["referrer_name"] = referrerName.String
		}
		if referrerCode.Valid {
			rewardData["referrer_code"] = referrerCode.String
		}
		if newUserName.Valid {
			rewardData["new_user_name"] = newUserName.String
		}
		if newUserCode.Valid {
			rewardData["new_user_code"] = newUserCode.String
		}

		rewards = append(rewards, rewardData)
	}

	return rewards, total, nil
}

