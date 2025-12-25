package model

import (
	"database/sql"
	"fmt"
	"time"

	"go_backend/internal/database"
)

// Invoice 表示发票抬头
type Invoice struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	InvoiceType   string    `json:"invoice_type"`   // personal, company
	Title         string    `json:"title"`          // 发票抬头
	TaxNumber     string    `json:"tax_number"`     // 纳税人识别号
	CompanyAddress string   `json:"company_address"` // 公司地址
	CompanyPhone  string    `json:"company_phone"`   // 公司电话
	BankName      string    `json:"bank_name"`       // 开户银行
	BankAccount   string    `json:"bank_account"`    // 银行账号
	IsDefault     bool      `json:"is_default"`      // 是否默认
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GetInvoiceByUserID 获取用户的发票抬头
func GetInvoiceByUserID(userID int) (*Invoice, error) {
	query := `
		SELECT id, user_id, invoice_type, title, tax_number, company_address, company_phone, 
		       bank_name, bank_account, is_default, created_at, updated_at
		FROM mini_app_invoices
		WHERE user_id = ?
		ORDER BY is_default DESC, created_at DESC
		LIMIT 1
	`

	var invoice Invoice
	err := database.DB.QueryRow(query, userID).Scan(
		&invoice.ID,
		&invoice.UserID,
		&invoice.InvoiceType,
		&invoice.Title,
		&invoice.TaxNumber,
		&invoice.CompanyAddress,
		&invoice.CompanyPhone,
		&invoice.BankName,
		&invoice.BankAccount,
		&invoice.IsDefault,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &invoice, nil
}

// CreateOrUpdateInvoice 创建或更新发票抬头
func CreateOrUpdateInvoice(userID int, invoiceData map[string]interface{}) (*Invoice, error) {
	// 如果设置为默认，先取消其他发票的默认状态
	isDefault := false
	if defaultVal, ok := invoiceData["is_default"].(bool); ok && defaultVal {
		isDefault = true
		_, err := database.DB.Exec(`
			UPDATE mini_app_invoices 
			SET is_default = 0 
			WHERE user_id = ?
		`, userID)
		if err != nil {
			return nil, err
		}
	}

	// 检查是否已存在发票抬头
	existingInvoice, _ := GetInvoiceByUserID(userID)

	if existingInvoice != nil {
		// 更新现有发票
		query := `
			UPDATE mini_app_invoices 
			SET invoice_type = ?, title = ?, tax_number = ?, company_address = ?, 
			    company_phone = ?, bank_name = ?, bank_account = ?, is_default = ?, updated_at = NOW()
			WHERE id = ? AND user_id = ?
		`

		invoiceType := getInvoiceStringValue(invoiceData, "invoice_type", "personal")
		title := getInvoiceStringValue(invoiceData, "title", "")
		taxNumber := getInvoiceStringValue(invoiceData, "tax_number", "")
		companyAddress := getInvoiceStringValue(invoiceData, "company_address", "")
		companyPhone := getInvoiceStringValue(invoiceData, "company_phone", "")
		bankName := getInvoiceStringValue(invoiceData, "bank_name", "")
		bankAccount := getInvoiceStringValue(invoiceData, "bank_account", "")

		_, err := database.DB.Exec(query,
			invoiceType, title, taxNumber, companyAddress,
			companyPhone, bankName, bankAccount, isDefault,
			existingInvoice.ID, userID,
		)
		if err != nil {
			return nil, fmt.Errorf("更新发票抬头失败: %w", err)
		}

		// 返回更新后的发票
		return GetInvoiceByUserID(userID)
	} else {
		// 创建新发票
		query := `
			INSERT INTO mini_app_invoices 
			(user_id, invoice_type, title, tax_number, company_address, company_phone, 
			 bank_name, bank_account, is_default, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		`

		invoiceType := getInvoiceStringValue(invoiceData, "invoice_type", "personal")
		title := getInvoiceStringValue(invoiceData, "title", "")
		taxNumber := getInvoiceStringValue(invoiceData, "tax_number", "")
		companyAddress := getInvoiceStringValue(invoiceData, "company_address", "")
		companyPhone := getInvoiceStringValue(invoiceData, "company_phone", "")
		bankName := getInvoiceStringValue(invoiceData, "bank_name", "")
		bankAccount := getInvoiceStringValue(invoiceData, "bank_account", "")

		result, err := database.DB.Exec(query,
			userID, invoiceType, title, taxNumber, companyAddress,
			companyPhone, bankName, bankAccount, isDefault,
		)
		if err != nil {
			return nil, fmt.Errorf("创建发票抬头失败: %w", err)
		}

		invoiceID, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("获取发票ID失败: %w", err)
		}

		// 查询并返回创建的发票
		var invoice Invoice
		err = database.DB.QueryRow(`
			SELECT id, user_id, invoice_type, title, tax_number, company_address, company_phone, 
			       bank_name, bank_account, is_default, created_at, updated_at
			FROM mini_app_invoices
			WHERE id = ?
		`, invoiceID).Scan(
			&invoice.ID,
			&invoice.UserID,
			&invoice.InvoiceType,
			&invoice.Title,
			&invoice.TaxNumber,
			&invoice.CompanyAddress,
			&invoice.CompanyPhone,
			&invoice.BankName,
			&invoice.BankAccount,
			&invoice.IsDefault,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("查询发票失败: %w", err)
		}

		return &invoice, nil
	}
}

// getInvoiceStringValue 从map中获取字符串值
func getInvoiceStringValue(data map[string]interface{}, key string, defaultValue string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return defaultValue
}

