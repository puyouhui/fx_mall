package model

import (
	"database/sql"
	"fmt"
	"time"

	"go_backend/internal/database"
)

// Category 分类模型
type Category struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ParentID  int       `json:"parent_id"`
	Sort      int       `json:"sort"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Icon      string    `json:"icon"`
}

// GetAllCategories 获取所有分类（包含树形结构）
func GetAllCategories() ([]map[string]interface{}, error) {
	fmt.Println("开始执行GetAllCategories函数")
	var categories []Category

	// 查询所有分类，不按排序字段排序
	query := "SELECT id, name, parent_id, sort, status, created_at, updated_at, icon FROM categories ORDER BY parent_id ASC"
	fmt.Println("执行SQL查询: ", query)
	rows, err := database.DB.Query(query)
	if err != nil {
		fmt.Println("SQL查询失败: ", err)
		return nil, err
	}
	defer rows.Close()

	// 扫描结果到分类切片
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.ParentID, &category.Sort, &category.Status, &category.CreatedAt, &category.UpdatedAt, &category.Icon); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	// 构建树形结构，不包含排序字段
	categoryMap := make(map[int]map[string]interface{})
	var rootCategories []map[string]interface{}

	// 先将所有分类转换为map并放入map中
	for _, category := range categories {
		categoryMap[category.ID] = map[string]interface{}{
			"id":         category.ID,
			"name":       category.Name,
			"parent_id":  category.ParentID,
			"status":     category.Status,
			"created_at": category.CreatedAt,
			"updated_at": category.UpdatedAt,
			"icon":       category.Icon,
			"children":   []map[string]interface{}{},
		}
	}

	// 构建树形结构
	for _, category := range categories {
		if category.ParentID == 0 {
			// 根分类
			rootCategories = append(rootCategories, categoryMap[category.ID])
		} else {
			// 子分类
			if parent, exists := categoryMap[category.ParentID]; exists {
				children := parent["children"].([]map[string]interface{})
				children = append(children, categoryMap[category.ID])
				parent["children"] = children
			}
		}
	}

	return rootCategories, nil
}

// GetCategoryByID 根据ID获取分类
func GetCategoryByID(id int) (*Category, error) {
	var category Category
	query := "SELECT id, name, parent_id, sort, status, created_at, updated_at, icon FROM categories WHERE id = ?"
	err := database.DB.QueryRow(query, id).Scan(&category.ID, &category.Name, &category.ParentID, &category.Sort, &category.Status, &category.CreatedAt, &category.UpdatedAt, &category.Icon)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

// CreateCategory 创建分类，不处理排序字段
func CreateCategory(category *Category) error {
	// 不处理排序字段，设置默认值0
	query := "INSERT INTO categories (name, parent_id, sort, status, created_at, updated_at, icon) VALUES (?, ?, 0, ?, NOW(), NOW(), ?)"
	res, err := database.DB.Exec(query, category.Name, category.ParentID, category.Status, category.Icon)
	if err != nil {
		return err
	}

	// 获取插入的ID
	lastID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	category.ID = int(lastID)

	return nil
}

// UpdateCategory 更新分类，不处理排序字段
func UpdateCategory(category *Category) error {
	// 不处理排序字段
	query := "UPDATE categories SET name = ?, parent_id = ?, status = ?, icon = ?, updated_at = NOW() WHERE id = ?"
	_, err := database.DB.Exec(query, category.Name, category.ParentID, category.Status, category.Icon, category.ID)
	return err
}

// DeleteCategory 删除分类
func DeleteCategory(id int) error {
	// 先检查是否有子分类
	var hasChildren bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM categories WHERE parent_id = ?)"
	err := database.DB.QueryRow(checkQuery, id).Scan(&hasChildren)
	if err != nil {
		return err
	}

	if hasChildren {
		// 有子分类，不允许删除
		return sql.ErrTxDone
	}

	// 删除分类
	deleteQuery := "DELETE FROM categories WHERE id = ?"
	_, err = database.DB.Exec(deleteQuery, id)
	return err
}