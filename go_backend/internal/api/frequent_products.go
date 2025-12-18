package api

import (
	"go_backend/internal/database"
	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// FrequentProduct 常购商品
type FrequentProduct struct {
	ProductID   int            `json:"product_id"`
	ProductName string         `json:"product_name"`
	SpecName    string         `json:"spec_name"`
	Image       string         `json:"image"`
	BuyCount    int            `json:"buy_count"` // 购买次数
	Product     *model.Product `json:"product,omitempty"`
}

// GetFrequentProducts 获取用户常购商品列表
func GetFrequentProducts(c *gin.Context) {
	// 从中间件获取用户
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}
	userID := user.ID

	// 查询用户的订单明细，按商品+规格分组统计购买次数
	query := `
		SELECT 
			oi.product_id,
			oi.product_name,
			oi.spec_name,
			oi.image,
			COUNT(*) as buy_count
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.user_id = ?
		GROUP BY oi.product_id, oi.product_name, oi.spec_name, oi.image
		ORDER BY buy_count DESC
		LIMIT 50
	`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		internalErrorResponse(c, "查询常购商品失败")
		return
	}
	defer rows.Close()

	var products []FrequentProduct
	for rows.Next() {
		var p FrequentProduct
		if err := rows.Scan(&p.ProductID, &p.ProductName, &p.SpecName, &p.Image, &p.BuyCount); err != nil {
			continue
		}

		// 获取商品详情（包含规格信息用于价格显示）
		product, err := model.GetProductByID(p.ProductID)
		if err == nil && product != nil {
			p.Product = product
		}

		products = append(products, p)
	}

	if products == nil {
		products = []FrequentProduct{}
	}

	successResponse(c, products, "")
}
