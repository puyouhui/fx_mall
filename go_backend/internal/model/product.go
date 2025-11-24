package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go_backend/internal/database"
)

// Spec 商品规格结构体
type Spec struct {
	Name           string  `json:"name"`
	WholesalePrice float64 `json:"wholesale_price"` // 批发价
	RetailPrice    float64 `json:"retail_price"`    // 零售价
	Cost           float64 `json:"cost"`            // 成本（用于计算利润）
	Description    string  `json:"description"`     // 规格描述（例如：≈1.5元/瓶）
}

// Product 商品模型
type Product struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	OriginalPrice float64   `json:"original_price,omitempty"` // 原价（废弃，使用规格价格）
	Price         float64   `json:"price,omitempty"`          // 现价（废弃，使用规格价格）
	CategoryID    int       `json:"category_id"`
	SupplierID    *int      `json:"supplier_id,omitempty"` // 供应商ID（可空）
	IsSpecial     bool      `json:"is_special"`            // 是否精选商品
	Images        []string  `json:"images"`                // 商品图片
	Specs         []Spec    `json:"specs"`                 // 商品规格
	Status        int       `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 注意：价格范围由前端计算，后端不再计算
}

// GetSpecialProductsWithPagination 获取精选商品并支持分页
func GetSpecialProductsWithPagination(pageNum, pageSize int) ([]Product, int, error) {
	var products []Product
	var total int

	// 计算偏移量
	offset := (pageNum - 1) * pageSize

	// 获取总数量
	countQuery := "SELECT COUNT(*) FROM products WHERE is_special = 1 AND status = 1"
	if err := database.DB.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	query := "SELECT id, name, description, original_price, price, category_id, is_special, images, specs, status, created_at, updated_at FROM products WHERE is_special = 1 AND status = 1 ORDER BY created_at DESC LIMIT ? OFFSET ?"
	rows, err := database.DB.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var product Product
		var imagesJSON, specsJSON string
		var dbPrice, dbOriginalPrice sql.NullFloat64 // 使用可空类型

		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &dbOriginalPrice, &dbPrice, &product.CategoryID, &product.IsSpecial, &imagesJSON, &specsJSON, &product.Status, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, 0, err
		}

		// 处理可空字段
		if dbPrice.Valid {
			product.Price = dbPrice.Float64
		}
		if dbOriginalPrice.Valid {
			product.OriginalPrice = dbOriginalPrice.Float64
		}

		// 解析JSON字符串到切片
		if err := json.Unmarshal([]byte(imagesJSON), &product.Images); err != nil {
			product.Images = []string{}
		}

		if err := json.Unmarshal([]byte(specsJSON), &product.Specs); err != nil {
			product.Specs = []Spec{}
		}

		products = append(products, product)
	}

	return products, total, nil
}

// GetAllProducts 获取所有商品
func GetAllProducts() ([]Product, error) {
	var products []Product

	query := "SELECT id, name, description, original_price, price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at FROM products WHERE status = 1 ORDER BY id DESC"
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var product Product
		var imagesJSON, specsJSON string
		var dbPrice, dbOriginalPrice sql.NullFloat64 // 使用可空类型
		var dbSupplierID sql.NullInt64

		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &dbOriginalPrice, &dbPrice, &product.CategoryID, &dbSupplierID, &product.IsSpecial, &imagesJSON, &specsJSON, &product.Status, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, fmt.Errorf("获取商品失败: %v", err)
		}

		// 处理可空字段
		if dbPrice.Valid {
			product.Price = dbPrice.Float64
		}
		if dbOriginalPrice.Valid {
			product.OriginalPrice = dbOriginalPrice.Float64
		}
		if dbSupplierID.Valid {
			supplierID := int(dbSupplierID.Int64)
			product.SupplierID = &supplierID
		}

		// 解析JSON字符串到切片
		if err := json.Unmarshal([]byte(imagesJSON), &product.Images); err != nil {
			product.Images = []string{}
		}

		if err := json.Unmarshal([]byte(specsJSON), &product.Specs); err != nil {
			product.Specs = []Spec{}
		}

		products = append(products, product)
	}

	return products, nil
}

// GetAllProductsWithPagination 获取所有商品并支持分页
func GetAllProductsWithPagination(pageNum, pageSize int) ([]Product, int, error) {
	var products []Product
	var total int

	// 计算偏移量
	offset := (pageNum - 1) * pageSize

	// 获取总数量
	countQuery := "SELECT COUNT(*) FROM products WHERE status = 1"
	if err := database.DB.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	query := "SELECT id, name, description, original_price, price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at FROM products WHERE status = 1 ORDER BY id DESC LIMIT ? OFFSET ?"
	rows, err := database.DB.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var product Product
		var imagesJSON, specsJSON string
		var dbPrice, dbOriginalPrice sql.NullFloat64 // 使用可空类型
		var dbSupplierID sql.NullInt64

		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &dbOriginalPrice, &dbPrice, &product.CategoryID, &dbSupplierID, &product.IsSpecial, &imagesJSON, &specsJSON, &product.Status, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("获取商品失败: %v", err)
		}

		// 处理可空字段
		if dbPrice.Valid {
			product.Price = dbPrice.Float64
		}
		if dbOriginalPrice.Valid {
			product.OriginalPrice = dbOriginalPrice.Float64
		}
		if dbSupplierID.Valid {
			supplierID := int(dbSupplierID.Int64)
			product.SupplierID = &supplierID
		}

		// 解析JSON字符串到切片
		if err := json.Unmarshal([]byte(imagesJSON), &product.Images); err != nil {
			product.Images = []string{}
		}

		if err := json.Unmarshal([]byte(specsJSON), &product.Specs); err != nil {
			product.Specs = []Spec{}
		}

		products = append(products, product)
	}

	return products, total, nil
}

// GetProductByID 根据ID获取商品
func GetProductByID(id int) (*Product, error) {
	var product Product
	var imagesJSON, specsJSON string
	var dbPrice, dbOriginalPrice sql.NullFloat64 // 使用可空类型

	query := "SELECT id, name, description, original_price, price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at FROM products WHERE id = ?"
	var dbSupplierID sql.NullInt64
	err := database.DB.QueryRow(query, id).Scan(&product.ID, &product.Name, &product.Description, &dbOriginalPrice, &dbPrice, &product.CategoryID, &dbSupplierID, &product.IsSpecial, &imagesJSON, &specsJSON, &product.Status, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 商品不存在
		}
		return nil, err
	}

	// 处理可空字段
	if dbPrice.Valid {
		product.Price = dbPrice.Float64
	}
	if dbOriginalPrice.Valid {
		product.OriginalPrice = dbOriginalPrice.Float64
	}
	if dbSupplierID.Valid {
		supplierID := int(dbSupplierID.Int64)
		product.SupplierID = &supplierID
	}

	// 解析JSON字符串到切片
	if err := json.Unmarshal([]byte(imagesJSON), &product.Images); err != nil {
		product.Images = []string{}
	}

	if err := json.Unmarshal([]byte(specsJSON), &product.Specs); err != nil {
		product.Specs = []Spec{}
	}

	return &product, nil
}

// CreateProduct 创建商品
func CreateProduct(product *Product) error {
	// 序列化图片和规格为JSON字符串
	imagesJSON, err := json.Marshal(product.Images)
	if err != nil {
		return err
	}

	specsJSON, err := json.Marshal(product.Specs)
	if err != nil {
		return err
	}

	// 检查是否有规格数据
	if len(product.Specs) == 0 {
		return fmt.Errorf("商品必须至少有一个规格")
	}

	// 商品本身的价格字段设置为NULL，不使用前端传递的值
	query := "INSERT INTO products (name, description, original_price, price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at) VALUES (?, ?, NULL, NULL, ?, ?, ?, ?, ?, ?, NOW(), NOW())"
	result, err := database.DB.Exec(query, product.Name, product.Description, product.CategoryID, product.SupplierID, product.IsSpecial, imagesJSON, specsJSON, product.Status)
	if err != nil {
		return fmt.Errorf("创建商品失败: %v", err)
	}

	// 获取插入的ID
	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	product.ID = int(lastID)
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	return nil
}

// UpdateProduct 更新商品
func UpdateProduct(product *Product) error {
	// 序列化图片和规格为JSON字符串
	imagesJSON, err := json.Marshal(product.Images)
	if err != nil {
		return err
	}

	specsJSON, err := json.Marshal(product.Specs)
	if err != nil {
		return err
	}

	// 检查是否有规格数据
	if len(product.Specs) == 0 {
		return fmt.Errorf("商品必须至少有一个规格")
	}

	// 商品本身的价格字段设置为NULL
	query := "UPDATE products SET name = ?, description = ?, original_price = NULL, price = NULL, category_id = ?, supplier_id = ?, is_special = ?, images = ?, specs = ?, status = ?, updated_at = NOW() WHERE id = ?"
	_, err = database.DB.Exec(query, product.Name, product.Description, product.CategoryID, product.SupplierID, product.IsSpecial, imagesJSON, specsJSON, product.Status, product.ID)
	if err != nil {
		return err
	}

	product.UpdatedAt = time.Now()

	return nil
}

// DeleteProduct 删除商品（软删除）
func DeleteProduct(id int) error {
	query := "UPDATE products SET status = 0, updated_at = NOW() WHERE id = ?"
	_, err := database.DB.Exec(query, id)
	return err
}

// SearchProductSuggestions 搜索商品建议（只返回商品名称）
// 搜索范围：商品名称(name)和商品描述(description)
func SearchProductSuggestions(keyword string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 10 // 默认返回10条建议
	}

	// 如果关键词为空，返回空数组
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []string{}, nil
	}
	
	var suggestions []string
	searchPattern := "%" + keyword + "%"
	
	// 查询商品名称，搜索范围：商品名称和描述
	// 使用 GROUP BY 去重，按最大 ID 降序排列（最新商品优先）
	query := "SELECT name FROM products WHERE status = 1 AND (name LIKE ? OR description LIKE ?) GROUP BY name ORDER BY MAX(id) DESC LIMIT ?"
	rows, err := database.DB.Query(query, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("搜索商品建议失败: %v", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("扫描商品名称失败: %v", err)
		}
		suggestions = append(suggestions, name)
	}

	// 检查是否有错误
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历搜索结果失败: %v", err)
	}
	
	return suggestions, nil
}

// SearchProductsWithPagination 搜索商品并支持分页
// 搜索范围：商品名称(name)和商品描述(description)
func SearchProductsWithPagination(keyword string, pageNum, pageSize int) ([]Product, int, error) {
	var products []Product
	var total int

	// 如果关键词为空，返回空结果
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []Product{}, 0, nil
	}

	// 计算偏移量
	offset := (pageNum - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	if pageSize <= 0 {
		pageSize = 10 // 默认每页10条
	}

	// 构建搜索查询（搜索商品名称和描述）
	searchPattern := "%" + keyword + "%"
	
	// 先查询总数
	countQuery := "SELECT COUNT(DISTINCT id) FROM products WHERE status = 1 AND (name LIKE ? OR description LIKE ?)"
	err := database.DB.QueryRow(countQuery, searchPattern, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询商品总数失败: %v", err)
	}

	// 查询商品列表（搜索范围：商品名称和描述）
	query := "SELECT id, name, description, original_price, price, category_id, is_special, images, specs, status, created_at, updated_at FROM products WHERE status = 1 AND (name LIKE ? OR description LIKE ?) ORDER BY id DESC LIMIT ? OFFSET ?"
	rows, err := database.DB.Query(query, searchPattern, searchPattern, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索商品失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product Product
		var imagesJSON, specsJSON string
		var dbPrice, dbOriginalPrice sql.NullFloat64

		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &dbOriginalPrice, &dbPrice, &product.CategoryID, &product.IsSpecial, &imagesJSON, &specsJSON, &product.Status, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描商品数据失败: %v", err)
		}

		// 处理可空字段
		if dbPrice.Valid {
			product.Price = dbPrice.Float64
		}
		if dbOriginalPrice.Valid {
			product.OriginalPrice = dbOriginalPrice.Float64
		}

		// 解析JSON字符串到切片
		if err := json.Unmarshal([]byte(imagesJSON), &product.Images); err != nil {
			product.Images = []string{}
		}

		if err := json.Unmarshal([]byte(specsJSON), &product.Specs); err != nil {
			product.Specs = []Spec{}
		}

		products = append(products, product)
	}

	// 检查是否有错误
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("遍历搜索结果失败: %v", err)
	}

	return products, total, nil
}

// GetProductsByCategoryWithPagination 根据分类ID获取商品并支持分页
// 如果categoryID是一级分类，则返回该一级分类及其所有子分类的商品
func GetProductsByCategoryWithPagination(categoryID, pageNum, pageSize int) ([]Product, int, error) {
	var products []Product
	var total int

	// 计算偏移量
	offset := (pageNum - 1) * pageSize

	// 首先检查该分类是否为一级分类（parent_id = 0）
	var parentID int
	var isParentCategory bool
	checkQuery := "SELECT parent_id FROM categories WHERE id = ?"
	err := database.DB.QueryRow(checkQuery, categoryID).Scan(&parentID)
	if err == nil && parentID == 0 {
		isParentCategory = true
	}

	var countQuery string
	var query string
	var args []interface{}

	if isParentCategory {
		// 如果是一级分类，查询该分类及其所有子分类的商品
		// 先获取所有子分类ID
		subCategoryQuery := "SELECT id FROM categories WHERE parent_id = ?"
		subRows, err := database.DB.Query(subCategoryQuery, categoryID)
		if err != nil {
			return nil, 0, fmt.Errorf("查询子分类失败: %v", err)
		}
		defer subRows.Close()

		var categoryIDs []int
		categoryIDs = append(categoryIDs, categoryID) // 包含一级分类本身
		for subRows.Next() {
			var subCategoryID int
			if err := subRows.Scan(&subCategoryID); err == nil {
				categoryIDs = append(categoryIDs, subCategoryID)
			}
		}

		// 构建IN查询
		placeholders := ""
		for i, id := range categoryIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, id)
		}

		countQuery = "SELECT COUNT(*) FROM products WHERE category_id IN (" + placeholders + ") AND status = 1"
		query = "SELECT id, name, description, original_price, price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at FROM products WHERE category_id IN (" + placeholders + ") AND status = 1 ORDER BY created_at DESC LIMIT ? OFFSET ?"
		args = append(args, pageSize, offset)
	} else {
		// 如果是二级分类，直接查询该分类的商品
		countQuery = "SELECT COUNT(*) FROM products WHERE category_id = ? AND status = 1"
		query = "SELECT id, name, description, original_price, price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at FROM products WHERE category_id = ? AND status = 1 ORDER BY created_at DESC LIMIT ? OFFSET ?"
		args = []interface{}{categoryID, pageSize, offset}
	}

	// 获取总数量
	if err := database.DB.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var product Product
		var imagesJSON, specsJSON string
		var dbPrice, dbOriginalPrice sql.NullFloat64 // 使用可空类型
		var dbSupplierID sql.NullInt64

		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &dbOriginalPrice, &dbPrice, &product.CategoryID, &dbSupplierID, &product.IsSpecial, &imagesJSON, &specsJSON, &product.Status, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, 0, err
		}

		// 处理可空字段
		if dbPrice.Valid {
			product.Price = dbPrice.Float64
		}
		if dbOriginalPrice.Valid {
			product.OriginalPrice = dbOriginalPrice.Float64
		}
		if dbSupplierID.Valid {
			supplierID := int(dbSupplierID.Int64)
			product.SupplierID = &supplierID
		}

		// 解析JSON字符串到切片
		if err := json.Unmarshal([]byte(imagesJSON), &product.Images); err != nil {
			product.Images = []string{}
		}

		if err := json.Unmarshal([]byte(specsJSON), &product.Specs); err != nil {
			product.Specs = []Spec{}
		}

		products = append(products, product)
	}

	return products, total, nil
}

// GetProductsByCategoryID 根据分类ID获取商品（兼容旧版本）
func GetProductsByCategoryID(categoryID int) ([]Product, error) {
	var products []Product

	query := "SELECT id, name, description, original_price, price, category_id, is_special, images, specs, status, created_at, updated_at FROM products WHERE category_id = ? AND status = 1 ORDER BY id DESC"
	rows, err := database.DB.Query(query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var product Product
		var imagesJSON, specsJSON string
		var dbPrice, dbOriginalPrice sql.NullFloat64 // 使用可空类型

		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &dbOriginalPrice, &dbPrice, &product.CategoryID, &product.IsSpecial, &imagesJSON, &specsJSON, &product.Status, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, err
		}

		// 处理可空字段
		if dbPrice.Valid {
			product.Price = dbPrice.Float64
		}
		if dbOriginalPrice.Valid {
			product.OriginalPrice = dbOriginalPrice.Float64
		}

		// 解析JSON字符串到切片
		if err := json.Unmarshal([]byte(imagesJSON), &product.Images); err != nil {
			product.Images = []string{}
		}

		if err := json.Unmarshal([]byte(specsJSON), &product.Specs); err != nil {
			product.Specs = []Spec{}
		}

		products = append(products, product)
	}

	return products, nil
}

// GetSpecialProducts 获取精选商品
func GetSpecialProducts() ([]Product, error) {
	var products []Product

	query := "SELECT id, name, description, original_price, price, category_id, is_special, images, specs, status, created_at, updated_at FROM products WHERE is_special = 1 AND status = 1 ORDER BY id DESC LIMIT 10"
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var product Product
		var imagesJSON, specsJSON string
		var dbPrice, dbOriginalPrice sql.NullFloat64 // 使用可空类型

		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &dbOriginalPrice, &dbPrice, &product.CategoryID, &product.IsSpecial, &imagesJSON, &specsJSON, &product.Status, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, err
		}

		// 处理可空字段
		if dbPrice.Valid {
			product.Price = dbPrice.Float64
		}
		if dbOriginalPrice.Valid {
			product.OriginalPrice = dbOriginalPrice.Float64
		}

		// 解析JSON字符串到切片
		if err := json.Unmarshal([]byte(imagesJSON), &product.Images); err != nil {
			product.Images = []string{}
		}

		if err := json.Unmarshal([]byte(specsJSON), &product.Specs); err != nil {
			product.Specs = []Spec{}
		}

		products = append(products, product)
	}

	return products, nil
}
