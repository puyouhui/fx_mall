package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go_backend/internal/database"
)

// PriceFeedback 价格反馈模型
type PriceFeedback struct {
	ID              int       `json:"id" db:"id"`
	UserID          *int      `json:"user_id" db:"user_id"`
	ProductID       int       `json:"product_id" db:"product_id"`
	ProductName     string    `json:"product_name" db:"product_name"`
	PlatformPriceMin float64  `json:"platform_price_min" db:"platform_price_min"`
	PlatformPriceMax float64  `json:"platform_price_max" db:"platform_price_max"`
	CompetitorPrice float64   `json:"competitor_price" db:"competitor_price"`
	Images          []string  `json:"images" db:"images"`
	Remark          string    `json:"remark" db:"remark"`
	Status          string    `json:"status" db:"status"`
	AdminRemark     string    `json:"admin_remark" db:"admin_remark"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	// 关联信息
	UserName        string    `json:"user_name,omitempty"`
	UserPhone       string    `json:"user_phone,omitempty"`
	UserCode        string    `json:"user_code,omitempty"`
}

// CreatePriceFeedback 创建价格反馈
func CreatePriceFeedback(userID *int, productID int, productName string, platformPriceMin, platformPriceMax, competitorPrice float64, images []string, remark string) (*PriceFeedback, error) {
	// 将图片数组转换为JSON
	imagesJSON, err := json.Marshal(images)
	if err != nil {
		return nil, fmt.Errorf("图片数据序列化失败: %w", err)
	}

	query := `
		INSERT INTO price_feedback (user_id, product_id, product_name, platform_price_min, platform_price_max, competitor_price, images, remark, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending')
	`
	
	result, err := database.DB.Exec(query, userID, productID, productName, platformPriceMin, platformPriceMax, competitorPrice, string(imagesJSON), remark)
	if err != nil {
		return nil, fmt.Errorf("创建价格反馈失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取插入ID失败: %w", err)
	}

	return GetPriceFeedbackByID(int(id))
}

// GetPriceFeedbackByID 根据ID获取价格反馈
func GetPriceFeedbackByID(id int) (*PriceFeedback, error) {
	query := `
		SELECT 
			pf.id, pf.user_id, pf.product_id, pf.product_name, pf.platform_price_min, pf.platform_price_max,
			pf.competitor_price, pf.images, pf.remark, pf.status, pf.admin_remark,
			pf.created_at, pf.updated_at,
			COALESCE(u.name, '') as user_name,
			COALESCE(u.phone, '') as user_phone,
			COALESCE(u.user_code, '') as user_code
		FROM price_feedback pf
		LEFT JOIN mini_app_users u ON pf.user_id = u.id
		WHERE pf.id = ?
	`
	
	var feedback PriceFeedback
	var imagesJSON sql.NullString
	var adminRemark sql.NullString
	var userName, userPhone, userCode sql.NullString
	
	err := database.DB.QueryRow(query, id).Scan(
		&feedback.ID, &feedback.UserID, &feedback.ProductID, &feedback.ProductName,
		&feedback.PlatformPriceMin, &feedback.PlatformPriceMax, &feedback.CompetitorPrice, &imagesJSON,
		&feedback.Remark, &feedback.Status, &adminRemark,
		&feedback.CreatedAt, &feedback.UpdatedAt,
		&userName, &userPhone, &userCode,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("价格反馈不存在")
		}
		return nil, fmt.Errorf("查询价格反馈失败: %w", err)
	}

	// 解析图片JSON
	if imagesJSON.Valid && imagesJSON.String != "" {
		if err := json.Unmarshal([]byte(imagesJSON.String), &feedback.Images); err != nil {
			feedback.Images = []string{}
		}
	} else {
		feedback.Images = []string{}
	}

	// 处理可空字段
	if adminRemark.Valid {
		feedback.AdminRemark = adminRemark.String
	}
	if userName.Valid {
		feedback.UserName = userName.String
	}
	if userPhone.Valid {
		feedback.UserPhone = userPhone.String
	}
	if userCode.Valid {
		feedback.UserCode = userCode.String
	}

	return &feedback, nil
}

// GetAllPriceFeedbacks 获取所有价格反馈（管理员）
func GetAllPriceFeedbacks(pageNum, pageSize int, status string) ([]*PriceFeedback, int, error) {
	// 构建查询条件
	whereClause := "1=1"
	args := []interface{}{}
	
	if status != "" {
		whereClause += " AND pf.status = ?"
		args = append(args, status)
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM price_feedback pf WHERE %s", whereClause)
	var total int
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 查询列表
	offset := (pageNum - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT 
			pf.id, pf.user_id, pf.product_id, pf.product_name, pf.platform_price_min, pf.platform_price_max,
			pf.competitor_price, pf.images, pf.remark, pf.status, pf.admin_remark,
			pf.created_at, pf.updated_at,
			COALESCE(u.name, '') as user_name,
			COALESCE(u.phone, '') as user_phone,
			COALESCE(u.user_code, '') as user_code
		FROM price_feedback pf
		LEFT JOIN mini_app_users u ON pf.user_id = u.id
		WHERE %s
		ORDER BY pf.created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, pageSize, offset)
	
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询价格反馈列表失败: %w", err)
	}
	defer rows.Close()

	var feedbacks []*PriceFeedback
	for rows.Next() {
		var feedback PriceFeedback
		var imagesJSON sql.NullString
		var adminRemark sql.NullString
		var userName, userPhone, userCode sql.NullString

		err := rows.Scan(
			&feedback.ID, &feedback.UserID, &feedback.ProductID, &feedback.ProductName,
			&feedback.PlatformPriceMin, &feedback.PlatformPriceMax, &feedback.CompetitorPrice, &imagesJSON,
			&feedback.Remark, &feedback.Status, &adminRemark,
			&feedback.CreatedAt, &feedback.UpdatedAt,
			&userName, &userPhone, &userCode,
		)
		if err != nil {
			continue
		}

		// 解析图片JSON
		if imagesJSON.Valid && imagesJSON.String != "" {
			if err := json.Unmarshal([]byte(imagesJSON.String), &feedback.Images); err != nil {
				feedback.Images = []string{}
			}
		} else {
			feedback.Images = []string{}
		}

		// 处理可空字段
		if adminRemark.Valid {
			feedback.AdminRemark = adminRemark.String
		}
		if userName.Valid {
			feedback.UserName = userName.String
		}
		if userPhone.Valid {
			feedback.UserPhone = userPhone.String
		}
		if userCode.Valid {
			feedback.UserCode = userCode.String
		}

		feedbacks = append(feedbacks, &feedback)
	}

	return feedbacks, total, nil
}

// UpdatePriceFeedbackStatus 更新价格反馈状态
func UpdatePriceFeedbackStatus(id int, status, adminRemark string) error {
	query := `
		UPDATE price_feedback 
		SET status = ?, admin_remark = ?, updated_at = NOW()
		WHERE id = ?
	`
	
	_, err := database.DB.Exec(query, status, adminRemark, id)
	if err != nil {
		return fmt.Errorf("更新价格反馈状态失败: %w", err)
	}

	return nil
}

