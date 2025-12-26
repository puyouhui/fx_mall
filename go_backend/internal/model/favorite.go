package model

import (
	"database/sql"
	"time"

	"go_backend/internal/database"
)

// Favorite 收藏模型
type Favorite struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	ProductID   int       `json:"product_id"`
	ProductName string    `json:"product_name,omitempty"`
	ProductImage string   `json:"product_image,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// 关联信息
	Product *Product `json:"product,omitempty"`
}

// CreateFavorite 创建收藏
func CreateFavorite(userID, productID int) (*Favorite, error) {
	// 检查是否已收藏
	existing, err := GetFavoriteByUserAndProduct(userID, productID)
	if err == nil && existing != nil {
		// 已收藏，直接返回
		return existing, nil
	}

	// 获取商品信息
	product, err := GetProductByID(productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, sql.ErrNoRows
	}

	// 获取商品图片
	productImage := ""
	if len(product.Images) > 0 {
		productImage = product.Images[0]
	}

	query := `
		INSERT INTO favorites (user_id, product_id, product_name, product_image)
		VALUES (?, ?, ?, ?)
	`

	result, err := database.DB.Exec(query, userID, productID, product.Name, productImage)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetFavoriteByID(int(id))
}

// GetFavoriteByID 根据ID获取收藏
func GetFavoriteByID(id int) (*Favorite, error) {
	query := `
		SELECT id, user_id, product_id, product_name, product_image, created_at, updated_at
		FROM favorites
		WHERE id = ?
	`

	var f Favorite
	err := database.DB.QueryRow(query, id).Scan(
		&f.ID,
		&f.UserID,
		&f.ProductID,
		&f.ProductName,
		&f.ProductImage,
		&f.CreatedAt,
		&f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &f, nil
}

// GetFavoriteByUserAndProduct 根据用户ID和商品ID获取收藏
func GetFavoriteByUserAndProduct(userID, productID int) (*Favorite, error) {
	query := `
		SELECT id, user_id, product_id, product_name, product_image, created_at, updated_at
		FROM favorites
		WHERE user_id = ? AND product_id = ?
		LIMIT 1
	`

	var f Favorite
	err := database.DB.QueryRow(query, userID, productID).Scan(
		&f.ID,
		&f.UserID,
		&f.ProductID,
		&f.ProductName,
		&f.ProductImage,
		&f.CreatedAt,
		&f.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &f, nil
}

// GetFavoritesByUserID 获取用户的收藏列表
func GetFavoritesByUserID(userID int) ([]*Favorite, error) {
	query := `
		SELECT id, user_id, product_id, product_name, product_image, created_at, updated_at
		FROM favorites
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []*Favorite
	for rows.Next() {
		var f Favorite
		err := rows.Scan(
			&f.ID,
			&f.UserID,
			&f.ProductID,
			&f.ProductName,
			&f.ProductImage,
			&f.CreatedAt,
			&f.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// 获取商品详情
		product, err := GetProductByID(f.ProductID)
		if err == nil && product != nil {
			f.Product = product
		}

		favorites = append(favorites, &f)
	}

	return favorites, nil
}

// DeleteFavorite 删除收藏
func DeleteFavorite(id int) error {
	query := `DELETE FROM favorites WHERE id = ?`
	_, err := database.DB.Exec(query, id)
	return err
}

// DeleteFavoriteByUserAndProduct 根据用户ID和商品ID删除收藏
func DeleteFavoriteByUserAndProduct(userID, productID int) error {
	query := `DELETE FROM favorites WHERE user_id = ? AND product_id = ?`
	_, err := database.DB.Exec(query, userID, productID)
	return err
}

