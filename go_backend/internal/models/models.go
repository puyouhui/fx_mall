package models

import "time"

// Category 分类模型
type Category struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	ParentID int    `json:"parent_id"` // 父分类ID，0表示一级分类
	Sort     int    `json:"sort"`      // 排序号
	Status   bool   `json:"status"`    // 状态，true表示启用，false表示禁用
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// Product 商品模型
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CategoryID  int       `json:"category_id"`
	Price       float64   `json:"price,omitempty"` // 特价商品价格，非特价则不显示
	IsSpecial   bool      `json:"is_special"`      // 是否特价商品
	Images      []string  `json:"images"`          // 商品图片
	Specs       []Spec    `json:"specs"`           // 商品规格
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Spec 商品规格
type Spec struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Carousel 轮播图模型
type Carousel struct {
	ID     int    `json:"id"`
	Image  string `json:"image"`
	Link   string `json:"link"`
	Title  string `json:"title"`
	Order  int    `json:"order"`
	Active bool   `json:"active"`
}

// CartItem 购物车项
type CartItem struct {
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price,omitempty"`
}