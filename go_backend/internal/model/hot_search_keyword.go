package model

import (
	"database/sql"
	"time"

	"go_backend/internal/database"
)

// HotSearchKeyword 热门搜索关键词模型
type HotSearchKeyword struct {
	ID        int       `json:"id"`
	Keyword   string    `json:"keyword"`
	Sort      int       `json:"sort"`
	Status    int       `json:"status"` // 1: 启用, 0: 禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetActiveHotSearchKeywords 获取启用的热门搜索关键词（小程序用）
func GetActiveHotSearchKeywords() ([]string, error) {
	query := `SELECT keyword FROM hot_search_keywords WHERE status = 1 ORDER BY sort ASC, id ASC`
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keywords []string
	for rows.Next() {
		var keyword string
		if err := rows.Scan(&keyword); err != nil {
			return nil, err
		}
		keywords = append(keywords, keyword)
	}
	return keywords, nil
}

// GetAllHotSearchKeywords 获取所有热门搜索关键词（管理后台用）
func GetAllHotSearchKeywords() ([]HotSearchKeyword, error) {
	query := `SELECT id, keyword, sort, status, created_at, updated_at FROM hot_search_keywords ORDER BY sort ASC, id ASC`
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keywords []HotSearchKeyword
	for rows.Next() {
		var k HotSearchKeyword
		if err := rows.Scan(&k.ID, &k.Keyword, &k.Sort, &k.Status, &k.CreatedAt, &k.UpdatedAt); err != nil {
			return nil, err
		}
		keywords = append(keywords, k)
	}
	return keywords, nil
}

// CreateHotSearchKeyword 创建热门搜索关键词
func CreateHotSearchKeyword(keyword string, sort, status int) (*HotSearchKeyword, error) {
	query := `INSERT INTO hot_search_keywords (keyword, sort, status, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())`
	result, err := database.DB.Exec(query, keyword, sort, status)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &HotSearchKeyword{
		ID:      int(id),
		Keyword: keyword,
		Sort:    sort,
		Status:  status,
	}, nil
}

// UpdateHotSearchKeyword 更新热门搜索关键词
func UpdateHotSearchKeyword(id int, keyword string, sort, status int) error {
	query := `UPDATE hot_search_keywords SET keyword = ?, sort = ?, status = ?, updated_at = NOW() WHERE id = ?`
	_, err := database.DB.Exec(query, keyword, sort, status, id)
	return err
}

// DeleteHotSearchKeyword 删除热门搜索关键词
func DeleteHotSearchKeyword(id int) error {
	query := `DELETE FROM hot_search_keywords WHERE id = ?`
	_, err := database.DB.Exec(query, id)
	return err
}

// GetHotSearchKeywordByID 根据ID获取热门搜索关键词
func GetHotSearchKeywordByID(id int) (*HotSearchKeyword, error) {
	query := `SELECT id, keyword, sort, status, created_at, updated_at FROM hot_search_keywords WHERE id = ?`
	var k HotSearchKeyword
	err := database.DB.QueryRow(query, id).Scan(&k.ID, &k.Keyword, &k.Sort, &k.Status, &k.CreatedAt, &k.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}


