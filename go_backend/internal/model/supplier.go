package model

import (
	"database/sql"
	"time"
)

// Supplier 供应商结构体
type Supplier struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`      // 供应商名称
	Contact   string    `json:"contact"`   // 联系人
	Phone     string    `json:"phone"`     // 联系电话
	Email     string    `json:"email"`     // 邮箱
	Address   string    `json:"address"`   // 地址
	Latitude  *float64  `json:"latitude"`  // 纬度
	Longitude *float64  `json:"longitude"` // 经度
	Username  string    `json:"username"`  // 登录账号
	Password  string    `json:"password"`  // 密码（加密存储）
	Status    int       `json:"status"`    // 状态：1-启用，0-禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetSupplierByID 根据ID获取供应商
func GetSupplierByID(db *sql.DB, id int) (*Supplier, error) {
	var supplier Supplier
	var latitude, longitude sql.NullFloat64
	query := "SELECT id, name, contact, phone, email, address, latitude, longitude, username, password, status, created_at, updated_at FROM suppliers WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&supplier.ID, &supplier.Name, &supplier.Contact, &supplier.Phone, &supplier.Email, &supplier.Address, &latitude, &longitude, &supplier.Username, &supplier.Password, &supplier.Status, &supplier.CreatedAt, &supplier.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if latitude.Valid {
		supplier.Latitude = &latitude.Float64
	}
	if longitude.Valid {
		supplier.Longitude = &longitude.Float64
	}
	return &supplier, nil
}

// GetSupplierByUsername 根据用户名获取供应商
func GetSupplierByUsername(db *sql.DB, username string) (*Supplier, error) {
	var supplier Supplier
	var latitude, longitude sql.NullFloat64
	query := "SELECT id, name, contact, phone, email, address, latitude, longitude, username, password, status, created_at, updated_at FROM suppliers WHERE username = ?"
	err := db.QueryRow(query, username).Scan(&supplier.ID, &supplier.Name, &supplier.Contact, &supplier.Phone, &supplier.Email, &supplier.Address, &latitude, &longitude, &supplier.Username, &supplier.Password, &supplier.Status, &supplier.CreatedAt, &supplier.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if latitude.Valid {
		supplier.Latitude = &latitude.Float64
	}
	if longitude.Valid {
		supplier.Longitude = &longitude.Float64
	}
	return &supplier, nil
}

// GetAllSuppliers 获取所有供应商
func GetAllSuppliers(db *sql.DB) ([]Supplier, error) {
	var suppliers []Supplier
	query := "SELECT id, name, contact, phone, email, address, latitude, longitude, username, password, status, created_at, updated_at FROM suppliers ORDER BY id DESC"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var supplier Supplier
		var latitude, longitude sql.NullFloat64
		err := rows.Scan(&supplier.ID, &supplier.Name, &supplier.Contact, &supplier.Phone, &supplier.Email, &supplier.Address, &latitude, &longitude, &supplier.Username, &supplier.Password, &supplier.Status, &supplier.CreatedAt, &supplier.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if latitude.Valid {
			supplier.Latitude = &latitude.Float64
		}
		if longitude.Valid {
			supplier.Longitude = &longitude.Float64
		}
		suppliers = append(suppliers, supplier)
	}

	return suppliers, nil
}

// CreateSupplier 创建供应商
func CreateSupplier(db *sql.DB, supplier *Supplier) error {
	query := "INSERT INTO suppliers (name, contact, phone, email, address, latitude, longitude, username, password, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())"
	result, err := db.Exec(query, supplier.Name, supplier.Contact, supplier.Phone, supplier.Email, supplier.Address, supplier.Latitude, supplier.Longitude, supplier.Username, supplier.Password, supplier.Status)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	supplier.ID = int(id)
	return nil
}

// UpdateSupplier 更新供应商
func UpdateSupplier(db *sql.DB, supplier *Supplier) error {
	query := "UPDATE suppliers SET name = ?, contact = ?, phone = ?, email = ?, address = ?, latitude = ?, longitude = ?, username = ?, status = ?, updated_at = NOW() WHERE id = ?"
	_, err := db.Exec(query, supplier.Name, supplier.Contact, supplier.Phone, supplier.Email, supplier.Address, supplier.Latitude, supplier.Longitude, supplier.Username, supplier.Status, supplier.ID)
	return err
}

// UpdateSupplierPassword 更新供应商密码
func UpdateSupplierPassword(db *sql.DB, id int, hashedPassword string) error {
	query := "UPDATE suppliers SET password = ?, updated_at = NOW() WHERE id = ?"
	_, err := db.Exec(query, hashedPassword, id)
	return err
}

// DeleteSupplier 删除供应商（软删除，设置为禁用状态）
func DeleteSupplier(db *sql.DB, id int) error {
	query := "UPDATE suppliers SET status = 0, updated_at = NOW() WHERE id = ?"
	_, err := db.Exec(query, id)
	return err
}
