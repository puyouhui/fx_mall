package api

import (
	"strconv"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// AddFavoriteRequest 添加收藏请求
type AddFavoriteRequest struct {
	ProductID int `json:"product_id" binding:"required"`
}

// GetUserFavorites 获取用户收藏列表
func GetUserFavorites(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	favorites, err := model.GetFavoritesByUserID(user.ID)
	if err != nil {
		internalErrorResponse(c, "获取收藏列表失败: "+err.Error())
		return
	}

	successResponse(c, favorites, "")
}

// AddFavorite 添加收藏
func AddFavorite(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	var req AddFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	if req.ProductID <= 0 {
		badRequestResponse(c, "商品ID不合法")
		return
	}

	// 检查商品是否存在
	product, err := model.GetProductByID(req.ProductID)
	if err != nil {
		internalErrorResponse(c, "获取商品失败: "+err.Error())
		return
	}
	if product == nil {
		notFoundResponse(c, "商品不存在")
		return
	}

	favorite, err := model.CreateFavorite(user.ID, req.ProductID)
	if err != nil {
		internalErrorResponse(c, "添加收藏失败: "+err.Error())
		return
	}

	successResponse(c, favorite, "收藏成功")
}

// DeleteFavorite 删除收藏
func DeleteFavorite(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	// 验证收藏是否属于当前用户
	favorite, err := model.GetFavoriteByID(id)
	if err != nil {
		notFoundResponse(c, "收藏不存在")
		return
	}
	if favorite.UserID != user.ID {
		unauthorizedResponse(c, "无权操作")
		return
	}

	if err := model.DeleteFavorite(id); err != nil {
		internalErrorResponse(c, "删除收藏失败: "+err.Error())
		return
	}

	successResponse(c, nil, "取消收藏成功")
}

// DeleteFavoriteByProductID 通过商品ID删除收藏
func DeleteFavoriteByProductID(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	productID, ok := parseID(c, "productId")
	if !ok {
		return
	}

	if err := model.DeleteFavoriteByUserAndProduct(user.ID, productID); err != nil {
		internalErrorResponse(c, "删除收藏失败: "+err.Error())
		return
	}

	successResponse(c, nil, "取消收藏成功")
}

// CheckFavorite 检查商品是否已收藏
func CheckFavorite(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	productIDStr := c.Query("product_id")
	if productIDStr == "" {
		badRequestResponse(c, "缺少商品ID参数")
		return
	}

	productID, err := strconv.Atoi(productIDStr)
	if err != nil || productID <= 0 {
		badRequestResponse(c, "商品ID不合法")
		return
	}

	favorite, err := model.GetFavoriteByUserAndProduct(user.ID, productID)
	if err != nil {
		internalErrorResponse(c, "检查收藏状态失败: "+err.Error())
		return
	}

	result := map[string]interface{}{
		"is_favorite": favorite != nil,
	}
	if favorite != nil {
		result["favorite_id"] = favorite.ID
	}

	successResponse(c, result, "")
}

