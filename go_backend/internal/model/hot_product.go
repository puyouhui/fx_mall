package model

import (
	"database/sql"
	"encoding/json"
	"time"
)

// HotProduct 热销产品关联结构体
type HotProduct struct {
	ID        int       `json:"id"`
	ProductID int      `json:"product_id"`
	Product   *Product  `json:"product,omitempty"` // 关联的商品信息
	Sort      int       `json:"sort"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetAllHotProducts 获取所有热销产品（包括禁用状态）
func GetAllHotProducts(db *sql.DB) ([]HotProduct, error) {
	query := `SELECT hp.id, hp.product_id, hp.sort, hp.status, hp.created_at, hp.updated_at 
			  FROM hot_products hp 
			  ORDER BY hp.sort ASC, hp.id DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hotProducts []HotProduct
	for rows.Next() {
		var hp HotProduct
		if err := rows.Scan(&hp.ID, &hp.ProductID, &hp.Sort, &hp.Status, &hp.CreatedAt, &hp.UpdatedAt); err != nil {
			return nil, err
		}
		hotProducts = append(hotProducts, hp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hotProducts, nil
}

// GetHotProducts 获取启用的热销产品（小程序用）
func GetHotProducts(db *sql.DB) ([]Product, error) {
	query := `SELECT p.id, p.name, p.description, p.original_price, p.price, p.category_id, 
			  p.supplier_id, p.is_special, p.images, p.specs, p.status, p.created_at, p.updated_at
			  FROM hot_products hp
			  INNER JOIN products p ON hp.product_id = p.id
			  WHERE hp.status = 1 AND p.status = 1
			  ORDER BY hp.sort ASC, hp.id DESC`
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		var imagesJSON, specsJSON string
		var dbPrice, dbOriginalPrice sql.NullFloat64
		var dbSupplierID sql.NullInt64

		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &dbOriginalPrice, &dbPrice, 
			&product.CategoryID, &dbSupplierID, &product.IsSpecial, &imagesJSON, &specsJSON, 
			&product.Status, &product.CreatedAt, &product.UpdatedAt); err != nil {
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
			supplierIDVal := int(dbSupplierID.Int64)
			product.SupplierID = &supplierIDVal
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// GetHotProductByID 根据ID获取热销产品关联
func GetHotProductByID(db *sql.DB, id int) (*HotProduct, error) {
	var hp HotProduct
	query := "SELECT id, product_id, sort, status, created_at, updated_at FROM hot_products WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&hp.ID, &hp.ProductID, &hp.Sort, &hp.Status, &hp.CreatedAt, &hp.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &hp, nil
}

// GetHotProductByProductID 根据商品ID获取热销产品关联
func GetHotProductByProductID(db *sql.DB, productID int) (*HotProduct, error) {
	var hp HotProduct
	query := "SELECT id, product_id, sort, status, created_at, updated_at FROM hot_products WHERE product_id = ?"
	err := db.QueryRow(query, productID).Scan(&hp.ID, &hp.ProductID, &hp.Sort, &hp.Status, &hp.CreatedAt, &hp.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &hp, nil
}

// CreateHotProduct 创建热销产品关联
func CreateHotProduct(db *sql.DB, hp *HotProduct) error {
	query := "INSERT INTO hot_products (product_id, sort, status, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())"
	result, err := db.Exec(query, hp.ProductID, hp.Sort, hp.Status)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	hp.ID = int(id)

	return nil
}

// UpdateHotProduct 更新热销产品关联
func UpdateHotProduct(db *sql.DB, hp *HotProduct) error {
	query := "UPDATE hot_products SET product_id = ?, sort = ?, status = ?, updated_at = NOW() WHERE id = ?"
	_, err := db.Exec(query, hp.ProductID, hp.Sort, hp.Status, hp.ID)
	return err
}

// DeleteHotProduct 删除热销产品关联
func DeleteHotProduct(db *sql.DB, id int) error {
	query := "DELETE FROM hot_products WHERE id = ?"
	_, err := db.Exec(query, id)
	return err
}

// UpdateHotProductSortItem 排序项结构
type UpdateHotProductSortItem struct {
	ID   int
	Sort int
}

// UpdateHotProductSort 批量更新排序
func UpdateHotProductSort(db *sql.DB, items []UpdateHotProductSortItem) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "UPDATE hot_products SET sort = ?, updated_at = NOW() WHERE id = ?"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		_, err := stmt.Exec(item.Sort, item.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

