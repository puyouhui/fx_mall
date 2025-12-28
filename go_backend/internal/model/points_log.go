package model

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"

	"go_backend/internal/database"
)

// PointsLog 积分明细
type PointsLog struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	Points      int        `json:"points"`        // 积分变动数量（正数为增加，负数为减少）
	BalanceAfter int       `json:"balance_after"` // 变动后积分余额
	Type        string     `json:"type"`          // 积分类型
	RelatedID   *int       `json:"related_id,omitempty"`   // 关联ID
	RelatedType *string    `json:"related_type,omitempty"` // 关联类型
	Description string     `json:"description"`   // 积分变动说明
	CreatedAt   time.Time  `json:"created_at"`
}

// AddPointsForOrder 为订单添加积分（订单完成收款后调用）
// 积分规则：每消费1元奖励1积分，四舍五入，必须是整数
func AddPointsForOrder(userID, orderID int, orderNumber string, totalAmount float64) error {
	// 计算积分：每消费1元奖励1积分，四舍五入
	points := int(math.Round(totalAmount))
	if points <= 0 {
		// 订单金额为0或负数，不发放积分
		return nil
	}

	// 获取用户当前积分
	var currentPoints int
	query := `SELECT COALESCE(points, 0) FROM mini_app_users WHERE id = ?`
	err := database.DB.QueryRow(query, userID).Scan(&currentPoints)
	if err != nil {
		return fmt.Errorf("获取用户积分失败: %v", err)
	}

	// 计算变动后积分余额
	balanceAfter := currentPoints + points

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 更新用户积分
	updateQuery := `
		UPDATE mini_app_users
		SET points = points + ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err = tx.Exec(updateQuery, points, userID)
	if err != nil {
		return fmt.Errorf("更新用户积分失败: %v", err)
	}

	// 记录积分明细
	relatedType := "order"
	description := fmt.Sprintf("订单完成奖励（订单号：%s，消费金额：%.2f元）", orderNumber, totalAmount)
	insertQuery := `
		INSERT INTO points_logs (
			user_id, points, balance_after, type, related_id, related_type, description, created_at
		) VALUES (?, ?, ?, 'order_reward', ?, ?, ?, NOW())
	`
	_, err = tx.Exec(insertQuery, userID, points, balanceAfter, orderID, relatedType, description)
	if err != nil {
		return fmt.Errorf("记录积分明细失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	log.Printf("成功为用户 %d 添加积分 %d（订单：%s，消费金额：%.2f元）", userID, points, orderNumber, totalAmount)
	return nil
}

// AddPoints 添加积分（通用方法，用于推荐奖励等）
func AddPoints(userID, points int, pointsType string, relatedID *int, relatedType *string, description string) error {
	if points == 0 {
		return nil
	}

	// 获取用户当前积分
	var currentPoints int
	query := `SELECT COALESCE(points, 0) FROM mini_app_users WHERE id = ?`
	err := database.DB.QueryRow(query, userID).Scan(&currentPoints)
	if err != nil {
		return fmt.Errorf("获取用户积分失败: %v", err)
	}

	// 计算变动后积分余额
	balanceAfter := currentPoints + points

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 更新用户积分
	updateQuery := `
		UPDATE mini_app_users
		SET points = points + ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err = tx.Exec(updateQuery, points, userID)
	if err != nil {
		return fmt.Errorf("更新用户积分失败: %v", err)
	}

	// 记录积分明细
	var relatedIDValue, relatedTypeValue interface{}
	if relatedID != nil {
		relatedIDValue = *relatedID
	}
	if relatedType != nil {
		relatedTypeValue = *relatedType
	}

	insertQuery := `
		INSERT INTO points_logs (
			user_id, points, balance_after, type, related_id, related_type, description, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
	`
	_, err = tx.Exec(insertQuery, userID, points, balanceAfter, pointsType, relatedIDValue, relatedTypeValue, description)
	if err != nil {
		return fmt.Errorf("记录积分明细失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetPointsLogs 获取用户积分明细列表（分页）
func GetPointsLogs(userID, pageNum, pageSize int) ([]PointsLog, int, error) {
	offset := (pageNum - 1) * pageSize

	// 查询总数
	var total int
	countQuery := `SELECT COUNT(*) FROM points_logs WHERE user_id = ?`
	err := database.DB.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	query := `
		SELECT id, user_id, points, balance_after, type, related_id, related_type, description, created_at
		FROM points_logs
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := database.DB.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []PointsLog
	for rows.Next() {
		var pointsLog PointsLog
		var relatedID sql.NullInt64
		var relatedType sql.NullString

		err := rows.Scan(
			&pointsLog.ID,
			&pointsLog.UserID,
			&pointsLog.Points,
			&pointsLog.BalanceAfter,
			&pointsLog.Type,
			&relatedID,
			&relatedType,
			&pointsLog.Description,
			&pointsLog.CreatedAt,
		)
		if err != nil {
			log.Printf("扫描积分明细失败: %v", err)
			continue
		}

		if relatedID.Valid {
			relatedIDInt := int(relatedID.Int64)
			pointsLog.RelatedID = &relatedIDInt
		}
		if relatedType.Valid {
			pointsLog.RelatedType = &relatedType.String
		}

		logs = append(logs, pointsLog)
	}

	return logs, total, nil
}

