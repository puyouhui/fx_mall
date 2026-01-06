package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ImageIndex 图片索引模型
type ImageIndex struct {
	ID         int       `json:"id"`
	ObjectName string    `json:"object_name"`
	ObjectURL  string    `json:"object_url"`
	Category   string    `json:"category"`
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
	FileType   string    `json:"file_type"`
	Width      *int      `json:"width,omitempty"`
	Height     *int      `json:"height,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateImageIndex 创建图片索引记录
func CreateImageIndex(db *sql.DB, img *ImageIndex) error {
	query := `
		INSERT INTO image_index 
		(object_name, object_url, category, file_name, file_size, file_type, width, height, uploaded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, img.ObjectName, img.ObjectURL, img.Category,
		img.FileName, img.FileSize, img.FileType, img.Width, img.Height, img.UploadedAt)
	return err
}

// DeleteImageIndexByURL 根据URL删除图片索引
func DeleteImageIndexByURL(db *sql.DB, imageURL string) error {
	query := `DELETE FROM image_index WHERE object_url = ?`
	_, err := db.Exec(query, imageURL)
	return err
}

// BatchDeleteImageIndex 批量删除图片索引
func BatchDeleteImageIndex(db *sql.DB, imageURLs []string) error {
	if len(imageURLs) == 0 {
		return nil
	}

	placeholders := make([]string, len(imageURLs))
	args := make([]interface{}, len(imageURLs))
	for i, url := range imageURLs {
		placeholders[i] = "?"
		args[i] = url
	}

	query := `DELETE FROM image_index WHERE object_url IN (` +
		strings.Join(placeholders, ",") + `)`
	_, err := db.Exec(query, args...)
	return err
}

// GetImageListWithPagination 从数据库分页查询图片列表
func GetImageListWithPagination(db *sql.DB, category string, pageNum, pageSize int) ([]map[string]interface{}, int, error) {
	var whereClause string
	var args []interface{}

	if category != "" {
		whereClause = "WHERE category = ?"
		args = append(args, category)
	}

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM image_index ` + whereClause
	var total int
	err := db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询图片总数失败: %v", err)
	}

	// 分页查询
	offset := (pageNum - 1) * pageSize
	query := `
		SELECT id, object_name, object_url, category, file_name, file_size, 
		       file_type, width, height, uploaded_at, created_at, updated_at
		FROM image_index 
		` + whereClause + `
		ORDER BY uploaded_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询图片列表失败: %v", err)
	}
	defer rows.Close()

	var images []map[string]interface{}
	for rows.Next() {
		var img ImageIndex
		err := rows.Scan(&img.ID, &img.ObjectName, &img.ObjectURL, &img.Category,
			&img.FileName, &img.FileSize, &img.FileType, &img.Width, &img.Height,
			&img.UploadedAt, &img.CreatedAt, &img.UpdatedAt)
		if err != nil {
			continue
		}

		images = append(images, map[string]interface{}{
			"name":      img.ObjectName,
			"url":       img.ObjectURL,
			"size":      img.FileSize,
			"updatedAt": img.UploadedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return images, total, nil
}

