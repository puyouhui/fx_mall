package api

import (
	"net/http"
	"strconv"
	"strings"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// AddPurchaseListItemRequest 请求参数
type AddPurchaseListItemRequest struct {
	ProductID int    `json:"product_id" binding:"required"`
	SpecName  string `json:"spec_name"`
	Quantity  int    `json:"quantity"`
}

// UpdatePurchaseListItemRequest 更新数量请求
type UpdatePurchaseListItemRequest struct {
	Quantity int `json:"quantity" binding:"required"`
}

func getMiniUserFromContext(c *gin.Context) (*model.MiniAppUser, bool) {
	openIDValue, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return nil, false
	}

	openID := openIDValue.(string)
	user, err := model.GetMiniAppUserByUniqueID(openID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return nil, false
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return nil, false
	}
	return user, true
}

// AddPurchaseListItem 添加采购单项
func AddPurchaseListItem(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	var req AddPurchaseListItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if req.ProductID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "商品ID不合法"})
		return
	}
	if req.Quantity == 0 {
		req.Quantity = 1
	}
	if req.Quantity < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "数量必须大于0"})
		return
	}

	product, err := model.GetProductByID(req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品失败: " + err.Error()})
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在"})
		return
	}

	specName := strings.TrimSpace(req.SpecName)
	if specName == "" {
		if len(product.Specs) == 1 {
			specName = product.Specs[0].Name
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择规格"})
			return
		}
	}

	var matchedSpec *model.Spec
	for i := range product.Specs {
		if product.Specs[i].Name == specName {
			matchedSpec = &product.Specs[i]
			break
		}
	}
	if matchedSpec == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "规格不存在"})
		return
	}

	specSnapshot := model.PurchaseSpecSnapshot{
		Name:           matchedSpec.Name,
		Description:    matchedSpec.Description,
		Cost:           matchedSpec.Cost,
		WholesalePrice: matchedSpec.WholesalePrice,
		RetailPrice:    matchedSpec.RetailPrice,
	}

	image := ""
	if len(product.Images) > 0 {
		image = product.Images[0]
	}

	item := &model.PurchaseListItem{
		UserID:       user.ID,
		ProductID:    product.ID,
		ProductName:  product.Name,
		ProductImage: image,
		SpecName:     specName,
		SpecSnapshot: specSnapshot,
		Quantity:     req.Quantity,
		IsSpecial:    product.IsSpecial,
	}

	result, err := model.AddOrUpdatePurchaseListItem(item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存采购单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result, "message": "添加成功"})
}

// GetPurchaseListItems 获取采购单列表
func GetPurchaseListItems(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	items, err := model.GetPurchaseListItemsByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": items, "message": "获取成功"})
}

// UpdatePurchaseListItem 更新采购单项
func UpdatePurchaseListItem(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	itemID, err := strconv.Atoi(idStr)
	if err != nil || itemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的采购单项ID"})
		return
	}

	var req UpdatePurchaseListItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if err := model.UpdatePurchaseListItemQuantity(itemID, user.ID, req.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	items, err := model.GetPurchaseListItemsByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取采购单失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": items, "message": "更新成功"})
}

// DeletePurchaseListItem 删除采购单项
func DeletePurchaseListItem(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	itemID, err := strconv.Atoi(idStr)
	if err != nil || itemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的采购单项ID"})
		return
	}

	if err := model.DeletePurchaseListItem(itemID, user.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ClearPurchaseList 清空采购单
func ClearPurchaseList(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	if err := model.ClearPurchaseList(user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "清空采购单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "采购单已清空"})
}
