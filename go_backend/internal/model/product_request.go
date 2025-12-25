package model

import (
	"database/sql"
	"time"

	"go_backend/internal/database"
)

// ProductRequest 新品需求模型
type ProductRequest struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	ProductName     string    `json:"product_name"`
	Brand           string    `json:"brand"`
	MonthlyQuantity int       `json:"monthly_quantity"`
	Description     string    `json:"description"`
	Status          string    `json:"status"` // pending, processing, completed, rejected
	AdminRemark     string    `json:"admin_remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	// 关联信息
	UserName  string `json:"user_name,omitempty"`
	UserPhone string `json:"user_phone,omitempty"`
	UserCode  string `json:"user_code,omitempty"`
}

// CreateProductRequest 创建新品需求
func CreateProductRequest(userID int, productName, brand string, monthlyQuantity int, description string) (*ProductRequest, error) {
	query := `
		INSERT INTO product_requests (user_id, product_name, brand, monthly_quantity, description)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := database.DB.Exec(query, userID, productName, brand, monthlyQuantity, description)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetProductRequestByID(int(id))
}

// GetProductRequestByID 根据ID获取新品需求
func GetProductRequestByID(id int) (*ProductRequest, error) {
	query := `
		SELECT pr.id, pr.user_id, pr.product_name, pr.brand, pr.monthly_quantity, 
		       pr.description, pr.status, pr.admin_remark, pr.created_at, pr.updated_at,
		       u.name, u.phone, u.user_code
		FROM product_requests pr
		LEFT JOIN mini_app_users u ON pr.user_id = u.id
		WHERE pr.id = ?
	`

	var pr ProductRequest
	var userName, userPhone, userCode, brand, description, adminRemark sql.NullString

	err := database.DB.QueryRow(query, id).Scan(
		&pr.ID, &pr.UserID, &pr.ProductName, &brand, &pr.MonthlyQuantity,
		&description, &pr.Status, &adminRemark, &pr.CreatedAt, &pr.UpdatedAt,
		&userName, &userPhone, &userCode,
	)
	if err != nil {
		return nil, err
	}

	if brand.Valid {
		pr.Brand = brand.String
	}
	if description.Valid {
		pr.Description = description.String
	}
	if adminRemark.Valid {
		pr.AdminRemark = adminRemark.String
	}
	if userName.Valid {
		pr.UserName = userName.String
	}
	if userPhone.Valid {
		pr.UserPhone = userPhone.String
	}
	if userCode.Valid {
		pr.UserCode = userCode.String
	}

	return &pr, nil
}

// GetProductRequestsByUserID 获取用户的新品需求列表
func GetProductRequestsByUserID(userID int) ([]*ProductRequest, error) {
	query := `
		SELECT pr.id, pr.user_id, pr.product_name, pr.brand, pr.monthly_quantity, 
		       pr.description, pr.status, pr.admin_remark, pr.created_at, pr.updated_at
		FROM product_requests pr
		WHERE pr.user_id = ?
		ORDER BY pr.created_at DESC
	`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*ProductRequest
	for rows.Next() {
		var pr ProductRequest
		var brand, description, adminRemark sql.NullString

		err := rows.Scan(
			&pr.ID, &pr.UserID, &pr.ProductName, &brand, &pr.MonthlyQuantity,
			&description, &pr.Status, &adminRemark, &pr.CreatedAt, &pr.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if brand.Valid {
			pr.Brand = brand.String
		}
		if description.Valid {
			pr.Description = description.String
		}
		if adminRemark.Valid {
			pr.AdminRemark = adminRemark.String
		}

		requests = append(requests, &pr)
	}

	return requests, nil
}

// GetAllProductRequests 获取所有新品需求（管理员）
func GetAllProductRequests(pageNum, pageSize int, status string) ([]*ProductRequest, int, error) {
	offset := (pageNum - 1) * pageSize
	var whereClause string
	var args []interface{}

	if status != "" {
		whereClause = "WHERE pr.status = ?"
		args = append(args, status)
	}

	// 获取总数
	countQuery := `
		SELECT COUNT(*) FROM product_requests pr
		` + whereClause

	var total int
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query := `
		SELECT pr.id, pr.user_id, pr.product_name, pr.brand, pr.monthly_quantity, 
		       pr.description, pr.status, pr.admin_remark, pr.created_at, pr.updated_at,
		       u.name, u.phone, u.user_code
		FROM product_requests pr
		LEFT JOIN mini_app_users u ON pr.user_id = u.id
		` + whereClause + `
		ORDER BY pr.created_at DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, pageSize, offset)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []*ProductRequest
	for rows.Next() {
		var pr ProductRequest
		var userName, userPhone, userCode, brand, description, adminRemark sql.NullString

		err := rows.Scan(
			&pr.ID, &pr.UserID, &pr.ProductName, &brand, &pr.MonthlyQuantity,
			&description, &pr.Status, &adminRemark, &pr.CreatedAt, &pr.UpdatedAt,
			&userName, &userPhone, &userCode,
		)
		if err != nil {
			return nil, 0, err
		}

		if brand.Valid {
			pr.Brand = brand.String
		}
		if description.Valid {
			pr.Description = description.String
		}
		if adminRemark.Valid {
			pr.AdminRemark = adminRemark.String
		}
		if userName.Valid {
			pr.UserName = userName.String
		}
		if userPhone.Valid {
			pr.UserPhone = userPhone.String
		}
		if userCode.Valid {
			pr.UserCode = userCode.String
		}

		requests = append(requests, &pr)
	}

	return requests, total, nil
}

// UpdateProductRequestStatus 更新新品需求状态
func UpdateProductRequestStatus(id int, status, adminRemark string) error {
	query := `
		UPDATE product_requests 
		SET status = ?, admin_remark = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := database.DB.Exec(query, status, adminRemark, id)
	return err
}
