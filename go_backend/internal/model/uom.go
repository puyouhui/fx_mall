package model

import (
	"database/sql"
	"time"

	"go_backend/internal/database"
)

// UomCategory 单位类别
type UomCategory struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	BaseUnitID  *int       `json:"base_unit_id,omitempty"`
	Sort        int        `json:"sort"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Units       []UomUnit  `json:"units,omitempty"`
	BaseUnit    *UomUnit   `json:"base_unit,omitempty"`
}

// UomUnit 单位
type UomUnit struct {
	ID         int       `json:"id"`
	CategoryID int       `json:"category_id"`
	Name       string    `json:"name"`
	Ratio      float64   `json:"ratio"`
	IsBase     int       `json:"is_base"`
	Sort       int       `json:"sort"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// GetAllUomCategories 获取所有单位类别（含单位列表）
func GetAllUomCategories() ([]UomCategory, error) {
	query := `SELECT id, name, base_unit_id, sort, created_at, updated_at 
		FROM uom_categories ORDER BY sort ASC, id ASC`
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []UomCategory
	for rows.Next() {
		var c UomCategory
		var dbBaseUnitID sql.NullInt64
		if err := rows.Scan(&c.ID, &c.Name, &dbBaseUnitID, &c.Sort, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		if dbBaseUnitID.Valid {
			id := int(dbBaseUnitID.Int64)
			c.BaseUnitID = &id
		}
		categories = append(categories, c)
	}

	// 加载每个类别的单位
	for i := range categories {
		units, err := GetUomUnitsByCategoryID(categories[i].ID)
		if err != nil {
			return nil, err
		}
		categories[i].Units = units
		for j := range units {
			if units[j].IsBase == 1 {
				categories[i].BaseUnit = &units[j]
				break
			}
		}
	}

	return categories, nil
}

// GetUomUnitsByCategoryID 根据类别ID获取单位列表
func GetUomUnitsByCategoryID(categoryID int) ([]UomUnit, error) {
	query := `SELECT id, category_id, name, ratio, is_base, sort, created_at, updated_at 
		FROM uom_units WHERE category_id = ? ORDER BY sort ASC, id ASC`
	rows, err := database.DB.Query(query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []UomUnit
	for rows.Next() {
		var u UomUnit
		if err := rows.Scan(&u.ID, &u.CategoryID, &u.Name, &u.Ratio, &u.IsBase, &u.Sort, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		units = append(units, u)
	}
	return units, nil
}

// GetUomCategoryByID 根据ID获取单位类别
func GetUomCategoryByID(id int) (*UomCategory, error) {
	var c UomCategory
	var dbBaseUnitID sql.NullInt64
	query := `SELECT id, name, base_unit_id, sort, created_at, updated_at 
		FROM uom_categories WHERE id = ?`
	err := database.DB.QueryRow(query, id).Scan(&c.ID, &c.Name, &dbBaseUnitID, &c.Sort, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if dbBaseUnitID.Valid {
		id := int(dbBaseUnitID.Int64)
		c.BaseUnitID = &id
	}
	units, _ := GetUomUnitsByCategoryID(c.ID)
	c.Units = units
	return &c, nil
}

// CreateUomCategory 创建单位类别
func CreateUomCategory(c *UomCategory) error {
	query := `INSERT INTO uom_categories (name, base_unit_id, sort, created_at, updated_at) 
		VALUES (?, ?, ?, NOW(), NOW())`
	res, err := database.DB.Exec(query, c.Name, c.BaseUnitID, c.Sort)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	c.ID = int(id)
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateUomCategory 更新单位类别
func UpdateUomCategory(c *UomCategory) error {
	query := `UPDATE uom_categories SET name = ?, base_unit_id = ?, sort = ?, updated_at = NOW() WHERE id = ?`
	_, err := database.DB.Exec(query, c.Name, c.BaseUnitID, c.Sort, c.ID)
	return err
}

// DeleteUomCategory 删除单位类别（会级联删除单位）
func DeleteUomCategory(id int) error {
	_, err := database.DB.Exec("DELETE FROM uom_categories WHERE id = ?", id)
	return err
}

// GetUomUnitByID 根据ID获取单位
func GetUomUnitByID(id int) (*UomUnit, error) {
	var u UomUnit
	query := `SELECT id, category_id, name, ratio, is_base, sort, created_at, updated_at 
		FROM uom_units WHERE id = ?`
	err := database.DB.QueryRow(query, id).Scan(&u.ID, &u.CategoryID, &u.Name, &u.Ratio, &u.IsBase, &u.Sort, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// CreateUomUnit 创建单位
func CreateUomUnit(u *UomUnit) error {
	query := `INSERT INTO uom_units (category_id, name, ratio, is_base, sort, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())`
	res, err := database.DB.Exec(query, u.CategoryID, u.Name, u.Ratio, u.IsBase, u.Sort)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	u.ID = int(id)
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateUomUnit 更新单位
func UpdateUomUnit(u *UomUnit) error {
	query := `UPDATE uom_units SET name = ?, ratio = ?, is_base = ?, sort = ?, updated_at = NOW() 
		WHERE id = ?`
	_, err := database.DB.Exec(query, u.Name, u.Ratio, u.IsBase, u.Sort, u.ID)
	return err
}

// DeleteUomUnit 删除单位
func DeleteUomUnit(id int) error {
	_, err := database.DB.Exec("DELETE FROM uom_units WHERE id = ?", id)
	return err
}

// CountBaseUnitsInCategory 统计类别下基准单位数量
func CountBaseUnitsInCategory(categoryID int) (int, error) {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM uom_units WHERE category_id = ? AND is_base = 1", categoryID).Scan(&count)
	return count, err
}

// CheckUomUnitNameExists 检查单位名称是否已存在（同一类别下）
func CheckUomUnitNameExists(categoryID int, name string, excludeID int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM uom_units WHERE category_id = ? AND name = ? AND id != ?`
	err := database.DB.QueryRow(query, categoryID, name, excludeID).Scan(&count)
	return count > 0, err
}

// CheckUomCategoryNameExists 检查类别名称是否已存在
func CheckUomCategoryNameExists(name string, excludeID int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM uom_categories WHERE name = ? AND id != ?`
	err := database.DB.QueryRow(query, name, excludeID).Scan(&count)
	return count > 0, err
}

// GetDefaultUomCategoryID 获取默认「件」单位类别ID（用于兼容老数据）
func GetDefaultUomCategoryID() (int, error) {
	var id int
	err := database.DB.QueryRow("SELECT id FROM uom_categories WHERE name = '件' LIMIT 1").Scan(&id)
	return id, err
}

// GetDefaultUomUnitID 获取默认「件」单位ID（基准单位）
func GetDefaultUomUnitID() (int, error) {
	var id int
	err := database.DB.QueryRow("SELECT uu.id FROM uom_units uu JOIN uom_categories uc ON uu.category_id = uc.id WHERE uc.name = '件' AND uu.is_base = 1 LIMIT 1").Scan(&id)
	return id, err
}
