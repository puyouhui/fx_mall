package model

import (
	"database/sql"
	"time"

	"go_backend/internal/database"
)

// SupplierApplication 供应商合作申请
type SupplierApplication struct {
	ID                int       `json:"id"`
	UserID            *int      `json:"user_id"`
	CompanyName       string    `json:"company_name"`
	ContactName       string    `json:"contact_name"`
	ContactPhone      string    `json:"contact_phone"`
	Email             string    `json:"email"`
	Address           string    `json:"address"`
	MainCategory      string    `json:"main_category"`
	CompanyIntro      string    `json:"company_intro"`
	CooperationIntent string    `json:"cooperation_intent"`
	Status            string    `json:"status"` // pending, approved, rejected
	AdminRemark       string    `json:"admin_remark"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	// 关联信息
	UserName  string `json:"user_name,omitempty"`
	UserPhone string `json:"user_phone,omitempty"`
	UserCode  string `json:"user_code,omitempty"`
}

// CreateSupplierApplication 创建供应商合作申请
func CreateSupplierApplication(userID *int, companyName, contactName, contactPhone, email, address, mainCategory, companyIntro, cooperationIntent string) (*SupplierApplication, error) {
	query := `
		INSERT INTO supplier_applications (user_id, company_name, contact_name, contact_phone, email, address, main_category, company_intro, cooperation_intent)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := database.DB.Exec(query, userID, companyName, contactName, contactPhone, email, address, mainCategory, companyIntro, cooperationIntent)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetSupplierApplicationByID(int(id))
}

// GetSupplierApplicationByID 根据ID获取申请
func GetSupplierApplicationByID(id int) (*SupplierApplication, error) {
	query := `
		SELECT sa.id, sa.user_id, sa.company_name, sa.contact_name, sa.contact_phone, sa.email, 
		       sa.address, sa.main_category, sa.company_intro, sa.cooperation_intent, 
		       sa.status, sa.admin_remark, sa.created_at, sa.updated_at,
		       u.name, u.phone, u.user_code
		FROM supplier_applications sa
		LEFT JOIN mini_app_users u ON sa.user_id = u.id
		WHERE sa.id = ?
	`

	var sa SupplierApplication
	var userID sql.NullInt64
	var userName, userPhone, userCode, email, address, companyIntro, cooperationIntent, adminRemark sql.NullString

	err := database.DB.QueryRow(query, id).Scan(
		&sa.ID, &userID, &sa.CompanyName, &sa.ContactName, &sa.ContactPhone, &email,
		&address, &sa.MainCategory, &companyIntro, &cooperationIntent,
		&sa.Status, &adminRemark, &sa.CreatedAt, &sa.UpdatedAt,
		&userName, &userPhone, &userCode,
	)
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		userIDInt := int(userID.Int64)
		sa.UserID = &userIDInt
	}
	if email.Valid {
		sa.Email = email.String
	}
	if address.Valid {
		sa.Address = address.String
	}
	if companyIntro.Valid {
		sa.CompanyIntro = companyIntro.String
	}
	if cooperationIntent.Valid {
		sa.CooperationIntent = cooperationIntent.String
	}
	if adminRemark.Valid {
		sa.AdminRemark = adminRemark.String
	}
	if userName.Valid {
		sa.UserName = userName.String
	}
	if userPhone.Valid {
		sa.UserPhone = userPhone.String
	}
	if userCode.Valid {
		sa.UserCode = userCode.String
	}

	return &sa, nil
}

// GetSupplierApplicationsByUserID 获取用户的申请列表
func GetSupplierApplicationsByUserID(userID int) ([]*SupplierApplication, error) {
	query := `
		SELECT sa.id, sa.user_id, sa.company_name, sa.contact_name, sa.contact_phone, sa.email, 
		       sa.address, sa.main_category, sa.company_intro, sa.cooperation_intent, 
		       sa.status, sa.admin_remark, sa.created_at, sa.updated_at
		FROM supplier_applications sa
		WHERE sa.user_id = ?
		ORDER BY sa.created_at DESC
	`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []*SupplierApplication
	for rows.Next() {
		var sa SupplierApplication
		var userID sql.NullInt64
		var email, address, companyIntro, cooperationIntent, adminRemark sql.NullString

		err := rows.Scan(
			&sa.ID, &userID, &sa.CompanyName, &sa.ContactName, &sa.ContactPhone, &email,
			&address, &sa.MainCategory, &companyIntro, &cooperationIntent,
			&sa.Status, &adminRemark, &sa.CreatedAt, &sa.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if userID.Valid {
			userIDInt := int(userID.Int64)
			sa.UserID = &userIDInt
		}
		if email.Valid {
			sa.Email = email.String
		}
		if address.Valid {
			sa.Address = address.String
		}
		if companyIntro.Valid {
			sa.CompanyIntro = companyIntro.String
		}
		if cooperationIntent.Valid {
			sa.CooperationIntent = cooperationIntent.String
		}
		if adminRemark.Valid {
			sa.AdminRemark = adminRemark.String
		}

		applications = append(applications, &sa)
	}

	return applications, nil
}

// GetAllSupplierApplications 获取所有申请（管理员）
func GetAllSupplierApplications(pageNum, pageSize int, status string) ([]*SupplierApplication, int, error) {
	offset := (pageNum - 1) * pageSize
	var whereClause string
	var args []interface{}

	if status != "" {
		whereClause = "WHERE sa.status = ?"
		args = append(args, status)
	}

	// 获取总数
	countQuery := `
		SELECT COUNT(*) FROM supplier_applications sa
		` + whereClause

	var total int
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query := `
		SELECT sa.id, sa.user_id, sa.company_name, sa.contact_name, sa.contact_phone, sa.email, 
		       sa.address, sa.main_category, sa.company_intro, sa.cooperation_intent, 
		       sa.status, sa.admin_remark, sa.created_at, sa.updated_at,
		       u.name, u.phone, u.user_code
		FROM supplier_applications sa
		LEFT JOIN mini_app_users u ON sa.user_id = u.id
		` + whereClause + `
		ORDER BY sa.created_at DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, pageSize, offset)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var applications []*SupplierApplication
	for rows.Next() {
		var sa SupplierApplication
		var userID sql.NullInt64
		var userName, userPhone, userCode, email, address, companyIntro, cooperationIntent, adminRemark sql.NullString

		err := rows.Scan(
			&sa.ID, &userID, &sa.CompanyName, &sa.ContactName, &sa.ContactPhone, &email,
			&address, &sa.MainCategory, &companyIntro, &cooperationIntent,
			&sa.Status, &adminRemark, &sa.CreatedAt, &sa.UpdatedAt,
			&userName, &userPhone, &userCode,
		)
		if err != nil {
			return nil, 0, err
		}

		if userID.Valid {
			userIDInt := int(userID.Int64)
			sa.UserID = &userIDInt
		}
		if email.Valid {
			sa.Email = email.String
		}
		if address.Valid {
			sa.Address = address.String
		}
		if companyIntro.Valid {
			sa.CompanyIntro = companyIntro.String
		}
		if cooperationIntent.Valid {
			sa.CooperationIntent = cooperationIntent.String
		}
		if adminRemark.Valid {
			sa.AdminRemark = adminRemark.String
		}
		if userName.Valid {
			sa.UserName = userName.String
		}
		if userPhone.Valid {
			sa.UserPhone = userPhone.String
		}
		if userCode.Valid {
			sa.UserCode = userCode.String
		}

		applications = append(applications, &sa)
	}

	return applications, total, nil
}

// UpdateSupplierApplicationStatus 更新申请状态
func UpdateSupplierApplicationStatus(id int, status, adminRemark string) error {
	query := `
		UPDATE supplier_applications 
		SET status = ?, admin_remark = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := database.DB.Exec(query, status, adminRemark, id)
	return err
}
