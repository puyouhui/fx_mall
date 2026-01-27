package model

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"go_backend/internal/database"
)

// RewardActivity 奖励活动配置
type RewardActivity struct {
	ID           int       `json:"id"`
	ActivityName string    `json:"activity_name"`        // 活动名称
	ActivityType string    `json:"activity_type"`        // 活动类型：referral-拉新活动，new_customer-新客奖励
	IsEnabled    bool      `json:"is_enabled"`           // 是否启用
	RewardType   string    `json:"reward_type"`          // 奖励类型：points/coupon/amount
	RewardValue  float64   `json:"reward_value"`         // 奖励值
	CouponIDs    []int     `json:"coupon_ids,omitempty"` // 多个优惠券ID（当reward_type为coupon时使用）
	Description  string    `json:"description"`          // 活动说明
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetRewardActivities 获取奖励活动列表（支持分页和筛选）
func GetRewardActivities(pageNum, pageSize int, activityType string) ([]RewardActivity, int, error) {
	// 计算偏移量
	offset := (pageNum - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	// 构建查询条件
	whereClause := "1=1"
	args := []interface{}{}

	if activityType != "" {
		whereClause += " AND activity_type = ?"
		args = append(args, activityType)
	}

	// 查询总数
	var total int
	countQuery := "SELECT COUNT(*) FROM reward_activities WHERE " + whereClause
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	query := `
		SELECT id, activity_name, activity_type, is_enabled, reward_type, reward_value, coupon_ids, description, created_at, updated_at
		FROM reward_activities
		WHERE ` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	activities := make([]RewardActivity, 0)
	for rows.Next() {
		var activity RewardActivity
		var couponIDsJSON sql.NullString

		if err := rows.Scan(
			&activity.ID,
			&activity.ActivityName,
			&activity.ActivityType,
			&activity.IsEnabled,
			&activity.RewardType,
			&activity.RewardValue,
			&couponIDsJSON,
			&activity.Description,
			&activity.CreatedAt,
			&activity.UpdatedAt,
		); err != nil {
			log.Printf("[GetRewardActivities] 扫描行数据失败: %v", err)
			continue
		}

		// 解析 coupon_ids JSON（兼容旧数据为空的情况）
		if couponIDsJSON.Valid && couponIDsJSON.String != "" {
			var ids []int
			if err := json.Unmarshal([]byte(couponIDsJSON.String), &ids); err != nil {
				log.Printf("[GetRewardActivities] 解析 coupon_ids 失败: %v, raw=%s", err, couponIDsJSON.String)
			} else {
				activity.CouponIDs = ids
			}
		}

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[GetRewardActivities] 遍历行时出错: %v", err)
		return nil, 0, err
	}

	return activities, total, nil
}

// GetRewardActivityByID 根据ID获取奖励活动
func GetRewardActivityByID(id int) (*RewardActivity, error) {
	var activity RewardActivity
	var couponIDsJSON sql.NullString

	query := `
		SELECT id, activity_name, activity_type, is_enabled, reward_type, reward_value, coupon_ids, description, created_at, updated_at
		FROM reward_activities
		WHERE id = ?
		LIMIT 1
	`

	err := database.DB.QueryRow(query, id).Scan(
		&activity.ID,
		&activity.ActivityName,
		&activity.ActivityType,
		&activity.IsEnabled,
		&activity.RewardType,
		&activity.RewardValue,
		&couponIDsJSON,
		&activity.Description,
		&activity.CreatedAt,
		&activity.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// 解析 coupon_ids JSON
	if couponIDsJSON.Valid && couponIDsJSON.String != "" {
		var ids []int
		if err := json.Unmarshal([]byte(couponIDsJSON.String), &ids); err != nil {
			log.Printf("[GetRewardActivityByID] 解析 coupon_ids 失败: %v, raw=%s", err, couponIDsJSON.String)
		} else {
			activity.CouponIDs = ids
		}
	}

	return &activity, nil
}

// CreateRewardActivity 创建奖励活动
func CreateRewardActivity(activity *RewardActivity) error {
	var couponIDsJSON interface{}
	if len(activity.CouponIDs) > 0 {
		if b, err := json.Marshal(activity.CouponIDs); err == nil {
			couponIDsJSON = string(b)
		} else {
			return err
		}
	} else {
		couponIDsJSON = nil
	}

	query := `
		INSERT INTO reward_activities (activity_name, activity_type, is_enabled, reward_type, reward_value, coupon_ids, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	result, err := database.DB.Exec(
		query,
		activity.ActivityName,
		activity.ActivityType,
		boolToTinyInt(activity.IsEnabled),
		activity.RewardType,
		activity.RewardValue,
		couponIDsJSON,
		activity.Description,
	)

	if err != nil {
		return err
	}

	// 获取插入的 ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	activity.ID = int(id)
	return nil
}

// UpdateRewardActivity 更新奖励活动
func UpdateRewardActivity(activity *RewardActivity) error {
	var couponIDsJSON interface{}
	if len(activity.CouponIDs) > 0 {
		if b, err := json.Marshal(activity.CouponIDs); err == nil {
			couponIDsJSON = string(b)
		} else {
			return err
		}
	} else {
		couponIDsJSON = nil
	}

	query := `
		UPDATE reward_activities
		SET activity_name = ?, activity_type = ?, is_enabled = ?, reward_type = ?, reward_value = ?, coupon_ids = ?, description = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := database.DB.Exec(
		query,
		activity.ActivityName,
		activity.ActivityType,
		boolToTinyInt(activity.IsEnabled),
		activity.RewardType,
		activity.RewardValue,
		couponIDsJSON,
		activity.Description,
		activity.ID,
	)

	return err
}

// DeleteRewardActivity 删除奖励活动
func DeleteRewardActivity(id int) error {
	query := `DELETE FROM reward_activities WHERE id = ?`
	_, err := database.DB.Exec(query, id)
	return err
}

// GetEnabledRewardActivityByType 根据活动类型获取启用的活动（用于业务逻辑）
func GetEnabledRewardActivityByType(activityType string) (*RewardActivity, error) {
	var activity RewardActivity
	var couponIDsJSON sql.NullString

	query := `
		SELECT id, activity_name, activity_type, is_enabled, reward_type, reward_value, coupon_ids, description, created_at, updated_at
		FROM reward_activities
		WHERE activity_type = ? AND is_enabled = 1
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := database.DB.QueryRow(query, activityType).Scan(
		&activity.ID,
		&activity.ActivityName,
		&activity.ActivityType,
		&activity.IsEnabled,
		&activity.RewardType,
		&activity.RewardValue,
		&couponIDsJSON,
		&activity.Description,
		&activity.CreatedAt,
		&activity.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if couponIDsJSON.Valid && couponIDsJSON.String != "" {
		var ids []int
		if err := json.Unmarshal([]byte(couponIDsJSON.String), &ids); err != nil {
			log.Printf("[GetEnabledRewardActivityByType] 解析 coupon_ids 失败: %v, raw=%s", err, couponIDsJSON.String)
		} else {
			activity.CouponIDs = ids
		}
	}

	return &activity, nil
}

// GrantNewCustomerLoginReward 新用户登录即送奖励（根据 activity_type = new_customer 的奖励活动配置）
func GrantNewCustomerLoginReward(userID int) error {
	activity, err := GetEnabledRewardActivityByType("new_customer")
	if err != nil {
		return err
	}
	if activity == nil || !activity.IsEnabled {
		// 未配置新客奖励活动，直接返回
		return nil
	}

	switch activity.RewardType {
	case "points":
		// 发放积分，并同步到奖励记录
		if activity.RewardValue <= 0 {
			return nil
		}
		points := int(activity.RewardValue)
		if points <= 0 {
			return nil
		}
		if err := AddPoints(userID, points, "new_customer_reward", nil, nil, "新客户登录奖励"); err != nil {
			return err
		}
		// 同步到奖励记录（受奖人为当前新客户）
		_ = LogActivityRewardForUser(userID, "points", float64(points), nil, "新客户登录奖励")
		return nil
	case "coupon":
		// 发放优惠券（支持多张），并记录到优惠券发放记录和奖励记录，原因：活动奖励
		if len(activity.CouponIDs) == 0 {
			return nil
		}
		for _, cid := range activity.CouponIDs {
			if cid <= 0 {
				continue
			}
			// 使用优惠券模板本身的有效期（通常配置为1年或更长），不单独缩短
			if err := IssueCouponToUser(userID, cid, 1, nil); err != nil {
				return err
			}

			// 查询优惠券信息，用于记录发放日志
			coupon, err := GetCouponByID(cid)
			if err != nil || coupon == nil {
				continue
			}

			logEntry := &CouponIssueLog{
				UserID:       userID,
				CouponID:     coupon.ID,
				CouponName:   coupon.Name,
				Quantity:     1,
				Reason:       "活动奖励",
				OperatorType: "system",
				OperatorID:   0,
				OperatorName: "系统",
				ExpiresAt:    nil,
			}
			if err := CreateCouponIssueLog(logEntry); err != nil {
				// 不影响主流程，只记录错误日志
				log.Printf("[GrantNewCustomerLoginReward] 记录优惠券发放日志失败(user_id=%d,coupon_id=%d): %v", userID, cid, err)
			}

			// 同步到奖励记录（受奖人为当前新客户）
			_ = LogActivityRewardForUser(userID, "coupon", 0, &coupon.ID, "新客户登录奖励")
		}
		return nil
	case "amount":
		// 金额奖励：暂时只做记录，不实际发放
		// 可以将来扩展为写入单独的金额奖励记录表
		return nil
	default:
		return nil
	}
}
