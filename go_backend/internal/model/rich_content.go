package model

import (
	"database/sql"
	"time"

	"go_backend/internal/database"
)

// RichContent 富文本内容模型
type RichContent struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`        // 富文本标题
	Content     string     `json:"content"`      // 富文本HTML内容
	ContentType string     `json:"content_type"` // 内容类型：notice(通知), activity(活动), other(其他)
	Status      string     `json:"status"`       // 状态：draft(草稿), published(已发布), archived(已归档)
	PublishedAt *time.Time `json:"published_at"` // 发布时间
	ViewCount   int        `json:"view_count"`   // 浏览次数
	CreatedBy   string     `json:"created_by"`   // 创建人
	UpdatedBy   string     `json:"updated_by"`   // 更新人
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateRichContent 创建富文本内容
func CreateRichContent(content *RichContent) error {
	content.Status = "draft"
	if content.ContentType == "" {
		content.ContentType = "notice"
	}

	query := `
		INSERT INTO rich_contents (title, content, content_type, status, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	result, err := database.DB.Exec(query, content.Title, content.Content, content.ContentType, content.Status, content.CreatedBy, content.UpdatedBy)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	content.ID = int(id)

	return nil
}

// GetRichContentByID 根据ID获取富文本内容
func GetRichContentByID(id int) (*RichContent, error) {
	var content RichContent
	var publishedAt sql.NullTime

	query := `
		SELECT id, title, content, content_type, status, published_at, view_count, created_by, updated_by, created_at, updated_at
		FROM rich_contents
		WHERE id = ?
	`

	err := database.DB.QueryRow(query, id).Scan(
		&content.ID, &content.Title, &content.Content, &content.ContentType, &content.Status,
		&publishedAt, &content.ViewCount, &content.CreatedBy, &content.UpdatedBy,
		&content.CreatedAt, &content.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if publishedAt.Valid {
		content.PublishedAt = &publishedAt.Time
	}

	return &content, nil
}

// GetRichContentByIDAndIncrementView 根据ID获取富文本内容并增加浏览次数
func GetRichContentByIDAndIncrementView(id int) (*RichContent, error) {
	content, err := GetRichContentByID(id)
	if err != nil {
		return nil, err
	}

	// 增加浏览次数
	updateQuery := `UPDATE rich_contents SET view_count = view_count + 1 WHERE id = ?`
	_, err = database.DB.Exec(updateQuery, id)
	if err != nil {
		return nil, err
	}

	content.ViewCount++
	return content, nil
}

// GetAllRichContents 获取所有富文本内容列表（分页）
func GetAllRichContents(page, pageSize int, contentType, status string) ([]RichContent, int64, error) {
	var contents []RichContent
	var total int64

	// 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if contentType != "" {
		whereClause += " AND content_type = ?"
		args = append(args, contentType)
	}

	if status != "" {
		whereClause += " AND status = ?"
		args = append(args, status)
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM rich_contents " + whereClause
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `
		SELECT id, title, content, content_type, status, published_at, view_count, created_by, updated_by, created_at, updated_at
		FROM rich_contents
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

	for rows.Next() {
		var content RichContent
		var publishedAt sql.NullTime

		err := rows.Scan(
			&content.ID, &content.Title, &content.Content, &content.ContentType, &content.Status,
			&publishedAt, &content.ViewCount, &content.CreatedBy, &content.UpdatedBy,
			&content.CreatedAt, &content.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if publishedAt.Valid {
			content.PublishedAt = &publishedAt.Time
		}

		contents = append(contents, content)
	}

	return contents, total, nil
}

// UpdateRichContent 更新富文本内容
func UpdateRichContent(id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE rich_contents SET "
	args := []interface{}{}
	first := true

	for key, value := range updates {
		if !first {
			query += ", "
		}
		query += key + " = ?"
		args = append(args, value)
		first = false
	}

	query += ", updated_at = NOW() WHERE id = ?"
	args = append(args, id)

	_, err := database.DB.Exec(query, args...)
	return err
}

// PublishRichContent 发布富文本内容
func PublishRichContent(id int, publishedBy string) error {
	query := `
		UPDATE rich_contents 
		SET status = 'published', published_at = NOW(), updated_by = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err := database.DB.Exec(query, publishedBy, id)
	return err
}

// ArchiveRichContent 归档富文本内容
func ArchiveRichContent(id int, archivedBy string) error {
	query := `
		UPDATE rich_contents 
		SET status = 'archived', updated_by = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err := database.DB.Exec(query, archivedBy, id)
	return err
}

// DeleteRichContent 删除富文本内容
func DeleteRichContent(id int) error {
	query := `DELETE FROM rich_contents WHERE id = ?`
	_, err := database.DB.Exec(query, id)
	return err
}

// GetPublishedRichContents 获取已发布的富文本内容列表（供小程序使用）
func GetPublishedRichContents(page, pageSize int, contentType string) ([]RichContent, int64, error) {
	var contents []RichContent
	var total int64

	// 构建查询条件
	whereClause := "WHERE status = 'published'"
	args := []interface{}{}

	if contentType != "" {
		whereClause += " AND content_type = ?"
		args = append(args, contentType)
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM rich_contents " + whereClause
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询，只返回必要字段
	offset := (page - 1) * pageSize
	query := `
		SELECT id, title, content_type, published_at, view_count
		FROM rich_contents
		` + whereClause + `
		ORDER BY published_at DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, pageSize, offset)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var content RichContent
		var publishedAt sql.NullTime

		err := rows.Scan(
			&content.ID, &content.Title, &content.ContentType, &publishedAt, &content.ViewCount,
		)
		if err != nil {
			return nil, 0, err
		}

		if publishedAt.Valid {
			content.PublishedAt = &publishedAt.Time
		}

		contents = append(contents, content)
	}

	return contents, total, nil
}
