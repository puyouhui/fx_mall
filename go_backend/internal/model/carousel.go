package model

import (
	"database/sql"
	"time"
)

// Carousel 轮播图结构体
type Carousel struct {
	ID        int       `json:"id"`
	Image     string    `json:"image"`
	Title     string    `json:"title"`
	Link      string    `json:"link"`
	Sort      int       `json:"sort"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetCarousels 获取所有轮播图
func GetCarousels(db *sql.DB) ([]Carousel, error) {
	query := "SELECT id, image, title, link, sort, status, created_at, updated_at FROM carousels WHERE status = 1 ORDER BY sort ASC"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var carousels []Carousel
	for rows.Next() {
		var carousel Carousel
		if err := rows.Scan(&carousel.ID, &carousel.Image, &carousel.Title, &carousel.Link, &carousel.Sort, &carousel.Status, &carousel.CreatedAt, &carousel.UpdatedAt); err != nil {
			return nil, err
		}
		carousels = append(carousels, carousel)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return carousels, nil
}

// GetAllCarousels 获取所有轮播图（包括禁用状态）
func GetAllCarousels(db *sql.DB) ([]Carousel, error) {
	query := "SELECT id, image, title, link, sort, status, created_at, updated_at FROM carousels ORDER BY sort ASC"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var carousels []Carousel
	for rows.Next() {
		var carousel Carousel
		if err := rows.Scan(&carousel.ID, &carousel.Image, &carousel.Title, &carousel.Link, &carousel.Sort, &carousel.Status, &carousel.CreatedAt, &carousel.UpdatedAt); err != nil {
			return nil, err
		}
		carousels = append(carousels, carousel)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return carousels, nil
}

// GetCarouselByID 根据ID获取轮播图
func GetCarouselByID(db *sql.DB, id int) (*Carousel, error) {
	var carousel Carousel
	query := "SELECT id, image, title, link, sort, status, created_at, updated_at FROM carousels WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&carousel.ID, &carousel.Image, &carousel.Title, &carousel.Link, &carousel.Sort, &carousel.Status, &carousel.CreatedAt, &carousel.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &carousel, nil
}

// CreateCarousel 创建轮播图
func CreateCarousel(db *sql.DB, carousel *Carousel) error {
	query := "INSERT INTO carousels (image, title, link, sort, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	result, err := db.Exec(query, carousel.Image, carousel.Title, carousel.Link, carousel.Sort, carousel.Status, time.Now(), time.Now())
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	carousel.ID = int(id)

	return nil
}

// UpdateCarousel 更新轮播图
func UpdateCarousel(db *sql.DB, carousel *Carousel) error {
	query := "UPDATE carousels SET image = ?, title = ?, link = ?, sort = ?, status = ?, updated_at = ? WHERE id = ?"
	_, err := db.Exec(query, carousel.Image, carousel.Title, carousel.Link, carousel.Sort, carousel.Status, time.Now(), carousel.ID)
	return err
}

// DeleteCarousel 删除轮播图
func DeleteCarousel(db *sql.DB, id int) error {
	query := "DELETE FROM carousels WHERE id = ?"
	_, err := db.Exec(query, id)
	return err
}