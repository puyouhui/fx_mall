package model

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go_backend/internal/database"
)

// Employee 表示员工
type Employee struct {
	ID           int       `json:"id"`
	EmployeeCode string    `json:"employee_code"`  // 员工码（5位数）
	Phone        string    `json:"phone"`          // 手机号（登录账号）
	Password     string    `json:"-"`              // 密码（不返回给前端）
	Name         string    `json:"name,omitempty"` // 员工姓名
	IsDelivery   bool      `json:"is_delivery"`    // 是否是配送员
	IsSales      bool      `json:"is_sales"`       // 是否是销售员
	Status       bool      `json:"status"`         // 状态：true-启用，false-禁用
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GenerateEmployeeCode 生成员工码（5位数）
func GenerateEmployeeCode() (string, error) {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())

	// 生成5位数（10000-99999）
	for i := 0; i < 100; i++ {
		code := fmt.Sprintf("%05d", 10000+rand.Intn(90000))
		var exists int
		err := database.DB.QueryRow(`
			SELECT COUNT(*) FROM employees WHERE employee_code = ?
		`, code).Scan(&exists)
		if err != nil {
			return "", err
		}
		if exists == 0 {
			return code, nil
		}
	}

	return "", fmt.Errorf("无法生成唯一的员工码")
}

// CreateEmployee 创建员工
func CreateEmployee(phone, password, name string, isDelivery, isSales bool) (*Employee, error) {
	// 检查手机号是否已存在
	var exists int
	err := database.DB.QueryRow(`
		SELECT COUNT(*) FROM employees WHERE phone = ?
	`, phone).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, fmt.Errorf("手机号已存在")
	}

	// 生成员工码
	employeeCode, err := GenerateEmployeeCode()
	if err != nil {
		return nil, err
	}

	// 插入新员工
	result, err := database.DB.Exec(`
		INSERT INTO employees (employee_code, phone, password, name, is_delivery, is_sales, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 1, NOW(), NOW())
	`, employeeCode, phone, password, name, isDelivery, isSales)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return GetEmployeeByID(int(id))
}

// GetEmployeeByID 根据ID获取员工
func GetEmployeeByID(id int) (*Employee, error) {
	var employee Employee
	var isDelivery, isSales, status sql.NullInt64

	err := database.DB.QueryRow(`
		SELECT id, employee_code, phone, password, name, is_delivery, is_sales, status, created_at, updated_at
		FROM employees
		WHERE id = ?
		LIMIT 1
	`, id).Scan(
		&employee.ID,
		&employee.EmployeeCode,
		&employee.Phone,
		&employee.Password,
		&employee.Name,
		&isDelivery,
		&isSales,
		&status,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	employee.IsDelivery = isDelivery.Valid && isDelivery.Int64 == 1
	employee.IsSales = isSales.Valid && isSales.Int64 == 1
	employee.Status = status.Valid && status.Int64 == 1

	return &employee, nil
}

// GetEmployeeByPhone 根据手机号获取员工
func GetEmployeeByPhone(phone string) (*Employee, error) {
	var employee Employee
	var isDelivery, isSales, status sql.NullInt64

	err := database.DB.QueryRow(`
		SELECT id, employee_code, phone, password, name, is_delivery, is_sales, status, created_at, updated_at
		FROM employees
		WHERE phone = ?
		LIMIT 1
	`, phone).Scan(
		&employee.ID,
		&employee.EmployeeCode,
		&employee.Phone,
		&employee.Password,
		&employee.Name,
		&isDelivery,
		&isSales,
		&status,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	employee.IsDelivery = isDelivery.Valid && isDelivery.Int64 == 1
	employee.IsSales = isSales.Valid && isSales.Int64 == 1
	employee.Status = status.Valid && status.Int64 == 1

	return &employee, nil
}

// GetEmployeeByEmployeeCode 根据员工码获取员工
func GetEmployeeByEmployeeCode(employeeCode string) (*Employee, error) {
	var employee Employee
	var isDelivery, isSales, status sql.NullInt64

	err := database.DB.QueryRow(`
		SELECT id, employee_code, phone, password, name, is_delivery, is_sales, status, created_at, updated_at
		FROM employees
		WHERE employee_code = ?
		LIMIT 1
	`, employeeCode).Scan(
		&employee.ID,
		&employee.EmployeeCode,
		&employee.Phone,
		&employee.Password,
		&employee.Name,
		&isDelivery,
		&isSales,
		&status,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	employee.IsDelivery = isDelivery.Valid && isDelivery.Int64 == 1
	employee.IsSales = isSales.Valid && isSales.Int64 == 1
	employee.Status = status.Valid && status.Int64 == 1

	return &employee, nil
}

// GetEmployees 获取员工列表
func GetEmployees(pageNum, pageSize int, keyword string) ([]Employee, int, error) {
	offset := (pageNum - 1) * pageSize
	var whereClause string
	var args []interface{}

	if keyword != "" {
		whereClause = "WHERE (employee_code LIKE ? OR phone LIKE ? OR name LIKE ?)"
		keywordPattern := "%" + keyword + "%"
		args = append(args, keywordPattern, keywordPattern, keywordPattern)
	}

	// 获取总数
	var total int
	countQuery := "SELECT COUNT(*) FROM employees " + whereClause
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query := `
		SELECT id, employee_code, phone, password, name, is_delivery, is_sales, status, created_at, updated_at
		FROM employees
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	employees := make([]Employee, 0)
	for rows.Next() {
		var employee Employee
		var isDelivery, isSales, status sql.NullInt64

		if err := rows.Scan(
			&employee.ID,
			&employee.EmployeeCode,
			&employee.Phone,
			&employee.Password,
			&employee.Name,
			&isDelivery,
			&isSales,
			&status,
			&employee.CreatedAt,
			&employee.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		employee.IsDelivery = isDelivery.Valid && isDelivery.Int64 == 1
		employee.IsSales = isSales.Valid && isSales.Int64 == 1
		employee.Status = status.Valid && status.Int64 == 1

		employees = append(employees, employee)
	}

	return employees, total, nil
}

// GetSalesEmployees 获取所有销售员列表（用于下拉选择）
func GetSalesEmployees() ([]Employee, error) {
	query := `
		SELECT id, employee_code, phone, password, name, is_delivery, is_sales, status, created_at, updated_at
		FROM employees
		WHERE is_sales = 1 AND status = 1
		ORDER BY created_at DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	employees := make([]Employee, 0)
	for rows.Next() {
		var employee Employee
		var isDelivery, isSales, status sql.NullInt64

		if err := rows.Scan(
			&employee.ID,
			&employee.EmployeeCode,
			&employee.Phone,
			&employee.Password,
			&employee.Name,
			&isDelivery,
			&isSales,
			&status,
			&employee.CreatedAt,
			&employee.UpdatedAt,
		); err != nil {
			return nil, err
		}

		employee.IsDelivery = isDelivery.Valid && isDelivery.Int64 == 1
		employee.IsSales = isSales.Valid && isSales.Int64 == 1
		employee.Status = status.Valid && status.Int64 == 1

		employees = append(employees, employee)
	}

	return employees, nil
}

// GetCustomersByEmployeeCode 根据员工码获取绑定的客户列表
func GetCustomersByEmployeeCode(employeeCode string) ([]map[string]interface{}, error) {
	query := `
		SELECT id, user_code, name, phone, created_at
		FROM mini_app_users
		WHERE sales_code = ?
		ORDER BY created_at DESC
	`

	rows, err := database.DB.Query(query, employeeCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var userCode, name, phone sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &userCode, &name, &phone, &createdAt); err != nil {
			return nil, err
		}

		customer := map[string]interface{}{
			"id":         id,
			"user_code":  getStringValue(userCode),
			"name":       getStringValue(name),
			"phone":      getStringValue(phone),
			"created_at": createdAt,
		}

		// 获取用户的默认地址信息
		defaultAddress, err := GetDefaultAddressByUserID(id)
		if err == nil && defaultAddress != nil {
			customer["default_address"] = map[string]interface{}{
				"name":       defaultAddress.Name,
				"contact":    defaultAddress.Contact,
				"phone":      defaultAddress.Phone,
				"address":    defaultAddress.Address,
				"store_type": defaultAddress.StoreType,
			}
		} else {
			customer["default_address"] = nil
		}

		// 补充统计信息：地址数量、下单次数
		if addrCount, err := CountAddressesByUserID(id); err == nil {
			customer["address_count"] = addrCount
		}
		if orderCount, err := CountOrdersByUserID(id); err == nil {
			customer["order_count"] = orderCount
		}

		customers = append(customers, customer)
	}

	return customers, nil
}

// CountCustomersByEmployeeCode 统计员工绑定的客户数量
func CountCustomersByEmployeeCode(employeeCode string) (int, error) {
	var count int
	err := database.DB.QueryRow(`
		SELECT COUNT(*) FROM mini_app_users WHERE sales_code = ?
	`, employeeCode).Scan(&count)
	return count, err
}

// getStringValue 辅助函数，处理sql.NullString
func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// UpdateEmployee 更新员工信息
func UpdateEmployee(id int, updateData map[string]interface{}) error {
	updates := []string{}
	args := []interface{}{}

	if phone, ok := updateData["phone"].(string); ok && phone != "" {
		// 检查手机号是否被其他员工使用
		var exists int
		err := database.DB.QueryRow(`
			SELECT COUNT(*) FROM employees WHERE phone = ? AND id != ?
		`, phone, id).Scan(&exists)
		if err != nil {
			return err
		}
		if exists > 0 {
			return fmt.Errorf("手机号已被其他员工使用")
		}
		updates = append(updates, "phone = ?")
		args = append(args, phone)
	}

	if password, ok := updateData["password"].(string); ok && password != "" {
		updates = append(updates, "password = ?")
		args = append(args, password)
	}

	if name, ok := updateData["name"].(string); ok {
		updates = append(updates, "name = ?")
		args = append(args, name)
	}

	if isDelivery, ok := updateData["is_delivery"].(bool); ok {
		updates = append(updates, "is_delivery = ?")
		args = append(args, isDelivery)
	}

	if isSales, ok := updateData["is_sales"].(bool); ok {
		updates = append(updates, "is_sales = ?")
		args = append(args, isSales)
	}

	if status, ok := updateData["status"].(bool); ok {
		updates = append(updates, "status = ?")
		args = append(args, status)
	}

	if len(updates) == 0 {
		return fmt.Errorf("没有要更新的字段")
	}

	updates = append(updates, "updated_at = NOW()")
	args = append(args, id)

	query := "UPDATE employees SET " + updates[0]
	for i := 1; i < len(updates); i++ {
		query += ", " + updates[i]
	}
	query += " WHERE id = ?"

	_, err := database.DB.Exec(query, args...)
	return err
}

// DeleteEmployee 删除员工（软删除，设置为禁用状态）
func DeleteEmployee(id int) error {
	_, err := database.DB.Exec(`
		UPDATE employees SET status = 0, updated_at = NOW() WHERE id = ?
	`, id)
	return err
}

// GetEmployeesByEmployeeCodes 批量获取员工信息
func GetEmployeesByEmployeeCodes(employeeCodes []string) (map[string]*Employee, error) {
	if len(employeeCodes) == 0 {
		return make(map[string]*Employee), nil
	}

	// 构建 IN 查询
	placeholders := make([]string, len(employeeCodes))
	args := make([]interface{}, len(employeeCodes))
	for i, code := range employeeCodes {
		placeholders[i] = "?"
		args[i] = code
	}

	query := fmt.Sprintf(`
		SELECT id, employee_code, phone, password, name, is_delivery, is_sales, status, created_at, updated_at
		FROM employees
		WHERE employee_code IN (%s)`, strings.Join(placeholders, ","))

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	employees := make(map[string]*Employee)
	for rows.Next() {
		var (
			employee              Employee
			isDelivery, isSales, status sql.NullInt64
		)

		err := rows.Scan(
			&employee.ID,
			&employee.EmployeeCode,
			&employee.Phone,
			&employee.Password,
			&employee.Name,
			&isDelivery,
			&isSales,
			&status,
			&employee.CreatedAt,
			&employee.UpdatedAt,
		)
		if err != nil {
			continue
		}

		employee.IsDelivery = isDelivery.Valid && isDelivery.Int64 == 1
		employee.IsSales = isSales.Valid && isSales.Int64 == 1
		employee.Status = status.Valid && status.Int64 == 1

		employees[employee.EmployeeCode] = &employee
	}

	return employees, nil
}
