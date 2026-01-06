package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go_backend/internal/config"
	"go_backend/internal/database"
	"go_backend/internal/model"
	"go_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// 响应辅助函数

// successResponse 返回成功响应
func successResponse(c *gin.Context, data interface{}, message string) {
	if message == "" {
		message = "success"
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data, "message": message})
}

// SaveImageIndex 保存图片索引到数据库（辅助函数，可导出供其他文件使用）
func SaveImageIndex(fileURL string, category string, fileName string, fileSize int64, contentType string) {
	cfg := config.Config.MinIO
	baseURLPrefix := fmt.Sprintf("%s/%s/", cfg.BaseURL, cfg.Bucket)
	var objectName string
	if strings.HasPrefix(fileURL, baseURLPrefix) {
		objectName = fileURL[len(baseURLPrefix):]
	} else {
		// 如果URL格式不符合预期，尝试从完整URL中提取
		parts := strings.Split(fileURL, "/")
		if len(parts) >= 3 {
			objectName = strings.Join(parts[len(parts)-2:], "/")
		} else {
			objectName = parts[len(parts)-1]
		}
	}

	// 提取文件名
	if strings.Contains(objectName, "/") {
		parts := strings.Split(objectName, "/")
		fileName = parts[len(parts)-1]
	}

	// 创建图片索引记录
	imgIndex := &model.ImageIndex{
		ObjectName: objectName,
		ObjectURL:  fileURL,
		Category:   category,
		FileName:   fileName,
		FileSize:   fileSize,
		FileType:   contentType,
		UploadedAt: time.Now(),
	}

	// 写入数据库索引（失败不影响上传，只记录日志）
	if err := model.CreateImageIndex(database.DB, imgIndex); err != nil {
		// 如果是重复键错误，忽略（可能已存在）
		if !strings.Contains(err.Error(), "Duplicate entry") {
			log.Printf("写入图片索引失败: %v（图片已上传，但索引未记录）", err)
		}
	} else {
		log.Printf("图片索引已写入数据库: %s", objectName)
	}
}

// errorResponse 返回错误响应
func errorResponse(c *gin.Context, statusCode int, code int, message string) {
	c.JSON(statusCode, gin.H{"code": code, "message": message})
}

// badRequestResponse 返回400错误响应
func badRequestResponse(c *gin.Context, message string) {
	errorResponse(c, http.StatusBadRequest, 400, message)
}

// internalErrorResponse 返回500错误响应
func internalErrorResponse(c *gin.Context, message string) {
	errorResponse(c, http.StatusInternalServerError, 500, message)
}

// notFoundResponse 返回404错误响应
func notFoundResponse(c *gin.Context, message string) {
	errorResponse(c, http.StatusNotFound, 404, message)
}

// unauthorizedResponse 返回401错误响应
func unauthorizedResponse(c *gin.Context, message string) {
	errorResponse(c, http.StatusUnauthorized, 401, message)
}

// 购物车数据存储（内存存储，实际项目中建议使用数据库或Redis）
// 参数解析辅助函数

// parseID 解析URL参数中的ID
func parseID(c *gin.Context, paramName string) (int, bool) {
	idStr := c.Param(paramName)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		badRequestResponse(c, "无效的"+paramName)
		return 0, false
	}
	return id, true
}

// parseQueryInt 解析查询参数中的整数，如果解析失败或不存在则返回默认值
func parseQueryInt(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}

// GetCarousels 获取轮播图列表
func GetCarousels(c *gin.Context) {
	// 从数据库获取启用状态的轮播图
	carousels, err := model.GetCarousels(database.DB)
	if err != nil {
		log.Printf("获取轮播图失败: %v", err)
		internalErrorResponse(c, "获取轮播图失败: "+err.Error())
		return
	}

	successResponse(c, carousels, "")
}

// GetCategories 获取分类列表（支持二级分类）
func GetCategories(c *gin.Context) {
	// 从数据库获取所有分类并构建树形结构
	categories, err := model.GetAllCategories()
	if err != nil {
		log.Printf("获取分类失败: %v", err)
		internalErrorResponse(c, "获取分类失败: "+err.Error())
		return
	}

	successResponse(c, categories, "")
}

// GetAllCategoriesForAdmin 获取所有分类（管理后台）
func GetAllCategoriesForAdmin(c *gin.Context) {
	// 从数据库获取所有分类并构建树形结构
	categories, err := model.GetAllCategories()
	if err != nil {
		log.Printf("获取分类失败: %v", err)
		internalErrorResponse(c, "获取分类失败: "+err.Error())
		return
	}

	log.Printf("成功获取到分类数量: %d", len(categories))

	successResponse(c, categories, "")
}

// CreateCategory 创建分类
func CreateCategory(c *gin.Context) {
	var category model.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		badRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证ParentID是否存在
	if category.ParentID != 0 {
		parent, err := model.GetCategoryByID(category.ParentID)
		if err != nil {
			log.Printf("验证父分类失败: %v", err)
			internalErrorResponse(c, "验证父分类失败: "+err.Error())
			return
		}
		if parent == nil {
			badRequestResponse(c, "父分类不存在")
			return
		}
	}

	// 创建分类
	if err := model.CreateCategory(&category); err != nil {
		log.Printf("创建分类失败: %v", err)
		internalErrorResponse(c, "创建分类失败: "+err.Error())
		return
	}

	successResponse(c, category, "创建成功")
}

// UpdateCategory 更新分类
func UpdateCategory(c *gin.Context) {
	categoryID, ok := parseID(c, "id")
	if !ok {
		return
	}

	var updateData model.Category
	if err := c.ShouldBindJSON(&updateData); err != nil {
		badRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 查找分类
	category, err := model.GetCategoryByID(categoryID)
	if err != nil {
		log.Printf("获取分类失败: %v", err)
		internalErrorResponse(c, "获取分类失败: "+err.Error())
		return
	}

	if category == nil {
		notFoundResponse(c, "分类不存在")
		return
	}

	// 更新分类信息
	category.Name = updateData.Name
	category.ParentID = updateData.ParentID
	category.Sort = updateData.Sort
	category.Status = updateData.Status
	category.Icon = updateData.Icon

	// 更新分类
	if err := model.UpdateCategory(category); err != nil {
		log.Printf("更新分类失败: %v", err)
		internalErrorResponse(c, "更新分类失败: "+err.Error())
		return
	}

	successResponse(c, category, "更新成功")
}

// DeleteCategory 删除分类
func DeleteCategory(c *gin.Context) {
	categoryID, ok := parseID(c, "id")
	if !ok {
		return
	}

	// 查找分类
	category, err := model.GetCategoryByID(categoryID)
	if err != nil {
		log.Printf("获取分类失败: %v", err)
		internalErrorResponse(c, "获取分类失败: "+err.Error())
		return
	}

	if category == nil {
		notFoundResponse(c, "分类不存在")
		return
	}

	// 删除分类
	if err := model.DeleteCategory(categoryID); err != nil {
		if err == sql.ErrTxDone {
			badRequestResponse(c, "该分类下有子分类，不能删除")
			return
		}
		log.Printf("删除分类失败: %v", err)
		internalErrorResponse(c, "删除分类失败: "+err.Error())
		return
	}

	successResponse(c, nil, "删除成功")
}

// 商品管理API

// GetAllProductsForAdmin 获取所有商品（管理后台）
func GetAllProductsForAdmin(c *gin.Context) {
	// 获取查询参数
	keyword := c.Query("keyword")
	categoryIDStr := c.Query("categoryId")
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)

	var products []model.Product
	var total int
	var err error

	// 如果有分类ID，使用分类筛选
	if categoryIDStr != "" {
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的分类ID"})
			return
		}
		// 使用分类筛选和分页
		products, total, err = model.GetProductsByCategoryWithPagination(categoryID, pageNum, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品失败: " + err.Error()})
			return
		}
		// 如果有关键词，在结果中进一步筛选
		if keyword != "" {
			filteredProducts := []model.Product{}
			for _, product := range products {
				if strings.Contains(strings.ToLower(product.Name), strings.ToLower(keyword)) ||
					strings.Contains(strings.ToLower(product.Description), strings.ToLower(keyword)) {
					filteredProducts = append(filteredProducts, product)
				}
			}
			products = filteredProducts
			total = len(filteredProducts)
		}
	} else if keyword != "" {
		// 只有关键词，使用搜索功能
		products, total, err = model.SearchProductsWithPagination(keyword, pageNum, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "搜索商品失败: " + err.Error()})
			return
		}
	} else {
		// 没有筛选条件，获取所有商品（带分页）
		products, total, err = model.GetAllProductsWithPagination(pageNum, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品失败: " + err.Error()})
			return
		}
	}

	// 获取分类名称 - 直接从数据库查询所有分类，确保能获取到所有分类名称
	categoryMap := make(map[int]string)
	categoryQuery := "SELECT id, name FROM categories"
	categoryRows, err := database.DB.Query(categoryQuery)
	if err == nil {
		defer categoryRows.Close()
		for categoryRows.Next() {
			var categoryID int
			var categoryName string
			if err := categoryRows.Scan(&categoryID, &categoryName); err == nil {
				categoryMap[categoryID] = categoryName
			}
		}
	}

	// 获取供应商名称
	supplierMap := make(map[int]string)
	var selfOperatedSupplierID int
	suppliers, err := model.GetAllSuppliers(database.DB)
	if err == nil {
		for _, supplier := range suppliers {
			supplierMap[supplier.ID] = supplier.Name
			// 记录自营供应商ID
			if supplier.Username == "self_operated" {
				selfOperatedSupplierID = supplier.ID
			}
		}
	}

	// 转换商品数据格式，添加分类名称和供应商信息
	result := []map[string]interface{}{}
	for _, product := range products {
		productMap := make(map[string]interface{})
		productMap["id"] = product.ID
		productMap["name"] = product.Name
		productMap["description"] = product.Description
		productMap["original_price"] = product.OriginalPrice
		productMap["price"] = product.Price
		productMap["category_id"] = product.CategoryID
		productMap["category_name"] = categoryMap[product.CategoryID]

		// 处理供应商信息：如果为null，自动绑定自营供应商
		if product.SupplierID != nil {
			productMap["supplier_id"] = *product.SupplierID
			productMap["supplier_name"] = supplierMap[*product.SupplierID]
		} else {
			// 如果供应商ID为null，自动设置为自营供应商
			if selfOperatedSupplierID > 0 {
				productMap["supplier_id"] = selfOperatedSupplierID
				productMap["supplier_name"] = supplierMap[selfOperatedSupplierID]
			} else {
				productMap["supplier_id"] = nil
				productMap["supplier_name"] = ""
			}
		}
		productMap["is_special"] = product.IsSpecial
		productMap["images"] = product.Images
		productMap["specs"] = product.Specs
		productMap["status"] = product.Status
		productMap["created_at"] = product.CreatedAt
		productMap["updated_at"] = product.UpdatedAt
		result = append(result, productMap)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result, "total": total, "message": "success"})
}

// 辅助函数：从分类树中提取所有分类ID和名称
func extractCategories(categories []map[string]interface{}, categoryMap map[int]string) {
	for _, category := range categories {
		// 获取分类ID和名称
		id, idOk := category["id"].(float64)
		name, nameOk := category["name"].(string)
		if idOk && nameOk {
			categoryMap[int(id)] = name
		}

		// 递归处理子分类
		children, childrenOk := category["children"].([]map[string]interface{})
		if childrenOk && len(children) > 0 {
			extractCategories(children, categoryMap)
		}
	}
}

// CreateProduct 创建商品
func CreateProduct(c *gin.Context) {
	var product model.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 验证分类ID是否存在
	category, err := model.GetCategoryByID(product.CategoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "验证分类失败: " + err.Error()})
		return
	}
	if category == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "分类不存在"})
		return
	}

	// 如果没有选择供应商，默认绑定"自营"供应商
	if product.SupplierID == nil {
		selfOperatedSupplier, err := model.GetSupplierByUsername(database.DB, "self_operated")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取默认供应商失败: " + err.Error()})
			return
		}
		if selfOperatedSupplier == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "系统错误：默认供应商不存在"})
			return
		}
		selfOperatedID := selfOperatedSupplier.ID
		product.SupplierID = &selfOperatedID
	}

	// 创建商品
	product.Status = 1
	if err := model.CreateProduct(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建商品失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": product, "message": "创建成功"})
}

// UpdateProduct 更新商品
func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	productID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的商品ID"})
		return
	}

	var updateData model.Product
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 查找商品
	product, err := model.GetProductByID(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品失败: " + err.Error()})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在"})
		return
	}

	// 验证分类ID是否存在
	if updateData.CategoryID > 0 {
		category, err := model.GetCategoryByID(updateData.CategoryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "验证分类失败: " + err.Error()})
			return
		}
		if category == nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "分类不存在"})
			return
		}
	}

	// 如果没有选择供应商，默认绑定"自营"供应商
	supplierID := updateData.SupplierID
	if supplierID == nil {
		selfOperatedSupplier, err := model.GetSupplierByUsername(database.DB, "self_operated")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取默认供应商失败: " + err.Error()})
			return
		}
		if selfOperatedSupplier == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "系统错误：默认供应商不存在"})
			return
		}
		selfOperatedID := selfOperatedSupplier.ID
		supplierID = &selfOperatedID
	}

	// 更新商品信息
	product.Name = updateData.Name
	product.Description = updateData.Description
	product.OriginalPrice = updateData.OriginalPrice
	product.Price = updateData.Price
	product.CategoryID = updateData.CategoryID
	product.SupplierID = supplierID
	product.IsSpecial = updateData.IsSpecial
	product.Images = updateData.Images
	product.Specs = updateData.Specs
	product.Status = updateData.Status

	// 更新商品
	if err := model.UpdateProduct(product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新商品失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": product, "message": "更新成功"})
}

// UpdateProductSpecialStatus 快速更新商品精选状态
func UpdateProductSpecialStatus(c *gin.Context) {
	id := c.Param("id")
	productID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的商品ID"})
		return
	}

	var updateData struct {
		IsSpecial bool `json:"is_special"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 查找商品
	product, err := model.GetProductByID(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品失败: " + err.Error()})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在"})
		return
	}

	// 更新精选状态
	query := "UPDATE products SET is_special = ?, updated_at = NOW() WHERE id = ?"
	_, err = database.DB.Exec(query, updateData.IsSpecial, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新精选状态失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteProduct 删除商品
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	productID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的商品ID"})
		return
	}

	// 查找商品
	product, err := model.GetProductByID(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品失败: " + err.Error()})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在"})
		return
	}

	// 检查是否有未完成的订单使用该商品
	var unfinishedOrderCount int
	checkQuery := `
		SELECT COUNT(DISTINCT o.id)
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		WHERE oi.product_id = ?
		  AND o.status NOT IN ('paid', 'cancelled')
	`
	err = database.DB.QueryRow(checkQuery, productID).Scan(&unfinishedOrderCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "检查订单失败: " + err.Error()})
		return
	}

	if unfinishedOrderCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("无法删除商品：该商品在 %d 个未完成订单中使用，请先完成或取消这些订单", unfinishedOrderCount),
		})
		return
	}

	// 删除商品（软删除）
	if err := model.DeleteProduct(productID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除商品失败: " + err.Error()})
		return
	}

	// 异步删除MinIO中的图片，避免影响主流程
	go func(images []string) {
		for _, image := range images {
			if err := utils.DeleteFile(image); err != nil {
				// 只记录错误，不影响主流程
				log.Printf("删除MinIO图片失败: %v", err)
			}
		}
	}(product.Images)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ListImages 列出MinIO桶中的所有图片（支持分页和目录过滤）
// 优化：使用数据库索引查询，实现真正的分页，性能从O(n)提升到O(log n)
func ListImages(c *gin.Context) {
	// 解析分页参数
	pageNum := 1
	pageSize := 30 // 默认每页30个（3行 x 10列）

	if pageNumStr := c.Query("pageNum"); pageNumStr != "" {
		if pn, err := strconv.Atoi(pageNumStr); err == nil && pn > 0 {
			pageNum = pn
		}
	}

	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
			// 限制最大pageSize，避免请求过大
			if pageSize > 500 {
				pageSize = 500
			}
		}
	}

	// 解析目录分类参数
	category := strings.TrimSpace(c.Query("category"))

	// 从数据库查询（真正的分页，性能优化）
	images, total, err := model.GetImageListWithPagination(database.DB, category, pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取图片列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"data":     images,
		"total":    total,
		"pageNum":  pageNum,
		"pageSize": pageSize,
		"message":  "获取成功",
	})
}

// BatchDeleteImages 批量删除图片
func BatchDeleteImages(c *gin.Context) {
	var req struct {
		ImageURLs []string `json:"imageUrls" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	if len(req.ImageURLs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请至少选择一个图片",
		})
		return
	}

	// 先删除数据库索引
	if err := model.BatchDeleteImageIndex(database.DB, req.ImageURLs); err != nil {
		log.Printf("删除图片索引失败: %v（继续删除MinIO文件）", err)
		// 不返回错误，继续删除MinIO文件
	}

	// 再删除MinIO中的文件
	err := utils.BatchDeleteImages(req.ImageURLs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除图片失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// UploadProductImage 上传商品图片到MinIO
func UploadProductImage(c *gin.Context) {
	// 检查是否有文件上传
	file, headers, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择要上传的图片: " + err.Error()})
		return
	}
	defer file.Close()

	if headers.Size > 5*1024*1024 { // 限制文件大小为5MB
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片大小不能超过5MB"})
		return
	}
	if !isImageFile(headers.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传JPG、PNG或GIF格式的图片"})
		return
	}

	// 上传图片到MinIO，商品图片存到products目录
	fileURL, err := utils.UploadFile("product", c.Request, "products")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "图片上传失败: " + err.Error()})
		return
	}

	// 写入数据库索引
	SaveImageIndex(fileURL, "products", headers.Filename, headers.Size, headers.Header.Get("Content-Type"))

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": map[string]string{"imageUrl": fileURL}, "message": "图片上传成功"})
}

// UploadImageWithCategory 上传图片到MinIO（支持指定目录分类，用于图库管理）
func UploadImageWithCategory(c *gin.Context) {
	// 检查是否有文件上传
	file, headers, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择要上传的图片: " + err.Error()})
		return
	}
	defer file.Close()

	if headers.Size > 10*1024*1024 { // 限制文件大小为10MB
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片大小不能超过10MB"})
		return
	}
	if !isImageFile(headers.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传JPG、PNG或GIF格式的图片"})
		return
	}

	// 获取目录分类参数，默认为"others"（无关紧要的图片）
	category := strings.TrimSpace(c.PostForm("category"))
	if category == "" {
		category = "others"
	}

	// 验证目录分类是否合法
	// 分类说明：
	// - products: 商品图片
	// - carousels: 轮播图（商品展示）
	// - categories: 分类图标（商品分类）
	// - users: 用户相关（头像、地址头像等）
	// - delivery: 配送相关（配送照片等）
	// - others: 其他临时或无关紧要的图片
	// - rich-content: 富文本内容图片
	validCategories := map[string]bool{
		"products":     true,
		"carousels":    true,
		"categories":   true,
		"users":        true,
		"delivery":     true,
		"others":       true,
		"rich-content": true,
	}
	if !validCategories[category] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的目录分类"})
		return
	}

	// 上传图片到MinIO
	fileURL, err := utils.UploadFile("image", c.Request, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "图片上传失败: " + err.Error()})
		return
	}

	// 写入数据库索引
	SaveImageIndex(fileURL, category, headers.Filename, headers.Size, headers.Header.Get("Content-Type"))

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": map[string]string{"imageUrl": fileURL}, "message": "图片上传成功"})
}

// GetSpecialProducts 获取精选商品
func GetSpecialProducts(c *gin.Context) {
	// 解析分页参数
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)

	// 从数据库获取特价商品
	specialProducts, total, err := model.GetSpecialProductsWithPagination(pageNum, pageSize)
	if err != nil {
		log.Printf("获取精选商品失败: %v", err)
		internalErrorResponse(c, "获取精选商品失败: "+err.Error())
		return
	}

	// 返回数据
	result := map[string]interface{}{
		"list":     specialProducts,
		"total":    total,
		"pageNum":  pageNum,
		"pageSize": pageSize,
	}
	successResponse(c, result, "")
}

// SearchProductSuggestions 搜索商品建议
// 搜索范围：商品名称和描述
func SearchProductSuggestions(c *gin.Context) {
	// 从查询参数中获取搜索关键词
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		// 如果关键词为空，返回空数组
		successResponse(c, []string{}, "")
		return
	}

	// 解析限制参数
	limit := parseQueryInt(c, "limit", 10)
	if limit <= 0 {
		limit = 10 // 默认返回10条建议
	}
	if limit > 50 {
		limit = 50 // 最大限制50条
	}

	// 从数据库获取商品建议（搜索范围：商品名称和描述）
	suggestions, err := model.SearchProductSuggestions(keyword, limit)
	if err != nil {
		log.Printf("获取搜索建议失败: %v", err)
		internalErrorResponse(c, "获取搜索建议失败: "+err.Error())
		return
	}

	successResponse(c, suggestions, "")
}

// SearchProducts 搜索商品
// 搜索范围：商品名称和描述
func SearchProducts(c *gin.Context) {
	// 从查询参数中获取搜索关键词
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		badRequestResponse(c, "搜索关键词不能为空")
		return
	}

	// 解析分页参数
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)

	// 限制分页参数范围
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 最大每页100条
	}

	// 从数据库搜索商品（搜索范围：商品名称和描述）
	products, total, err := model.SearchProductsWithPagination(keyword, pageNum, pageSize)
	if err != nil {
		log.Printf("搜索商品失败: %v", err)
		internalErrorResponse(c, "搜索商品失败: "+err.Error())
		return
	}

	// 构建返回数据结构
	result := map[string]interface{}{
		"list":     products,
		"total":    total,
		"pageNum":  pageNum,
		"pageSize": pageSize,
	}

	successResponse(c, result, "")
}

// GetProductsByCategory 获取分类下的商品
func GetProductsByCategory(c *gin.Context) {
	// 从查询参数中获取分类ID
	categoryIDStr := c.Query("categoryId")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		badRequestResponse(c, "分类ID无效")
		return
	}

	// 解析分页参数
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)

	// 从数据库获取分类下的商品
	products, total, err := model.GetProductsByCategoryWithPagination(categoryID, pageNum, pageSize)
	if err != nil {
		log.Printf("获取分类商品失败: %v", err)
		internalErrorResponse(c, "获取分类商品失败: "+err.Error())
		return
	}

	// 构建返回数据结构
	result := map[string]interface{}{
		"list":     products,
		"total":    total,
		"pageNum":  pageNum,
		"pageSize": pageSize,
	}

	successResponse(c, result, "")
}

// GetProductDetail 获取商品详情
func GetProductDetail(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	// 从数据库获取商品详情
	product, err := model.GetProductByID(id)
	if err != nil {
		log.Printf("获取商品详情失败: %v", err)
		internalErrorResponse(c, "获取商品详情失败: "+err.Error())
		return
	}

	if product == nil {
		notFoundResponse(c, "商品不存在")
		return
	}

	// 转换数据结构以匹配前端期望
	// 将specs转换为specifications
	specifications := make([]struct {
		Name string `json:"name"`
	}, len(product.Specs))

	for i, spec := range product.Specs {
		specifications[i] = struct {
			Name string `json:"name"`
		}{Name: spec.Name}
	}

	// 构建返回的数据结构，添加前端需要的额外字段
	responseData := struct {
		ID             int      `json:"id"`
		Name           string   `json:"name"`
		Description    string   `json:"description"`
		Price          float64  `json:"price"`
		OriginalPrice  float64  `json:"original_price"`
		CategoryID     int      `json:"category_id"`
		CategoryName   string   `json:"category_name"` // 分类名称（默认值）
		SupplierID     *int     `json:"supplier_id"`   // 供应商ID
		IsSpecial      bool     `json:"is_special"`
		Images         []string `json:"images"`
		Specifications []struct {
			Name string `json:"name"`
		} `json:"specifications"`
		Specs     []model.Spec `json:"specs"`   // 完整规格信息，包含名称、描述、价格和原价
		Stock     int          `json:"stock"`   // 库存（默认值）
		Sales     int          `json:"sales"`   // 销量（默认值）
		Details   string       `json:"details"` // 详细描述（默认值）
		CreatedAt time.Time    `json:"created_at"`
		UpdatedAt time.Time    `json:"updated_at"`
	}{}

	// 填充数据
	responseData.ID = product.ID
	responseData.Name = product.Name
	responseData.Description = product.Description
	responseData.Price = product.Price
	responseData.OriginalPrice = product.OriginalPrice
	responseData.CategoryID = product.CategoryID
	responseData.CategoryName = "商品分类"           // 可以后续从数据库获取真实分类名称
	responseData.SupplierID = product.SupplierID // 供应商ID
	responseData.IsSpecial = product.IsSpecial
	responseData.Images = product.Images
	responseData.Specifications = specifications
	responseData.Specs = product.Specs         // 完整的规格信息，包含名称、描述、价格和原价
	responseData.Stock = 100                   // 可以后续从数据库获取真实库存
	responseData.Sales = 50                    // 可以后续从数据库获取真实销量
	responseData.Details = product.Description // 使用描述作为详情
	responseData.CreatedAt = product.CreatedAt
	responseData.UpdatedAt = product.UpdatedAt

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": responseData, "message": "success"})
}

// AdminLogin 管理员登录
func AdminLogin(c *gin.Context) {
	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 根据用户名获取管理员信息
	admin, err := model.GetAdminByUsername(database.DB, loginReq.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "登录失败: " + err.Error()})
		return
	}

	if admin == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	// 使用bcrypt验证密码
	if !utils.CheckPasswordHash(loginReq.Password, admin.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	// 使用JWT库生成token
	token, err := utils.GenerateToken(admin.Username, admin.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成token失败: " + err.Error()})
		return
	}

	// 返回登录成功响应
	loginRes := struct {
		Token string `json:"token"`
		Admin struct {
			ID        int       `json:"id"`
			Username  string    `json:"username"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"admin"`
	}{
		Token: token,
		Admin: struct {
			ID        int       `json:"id"`
			Username  string    `json:"username"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		}{admin.ID, admin.Username, admin.CreatedAt, admin.UpdatedAt},
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": loginRes, "message": "登录成功"})
}

// AdminLogout 管理员登出
func AdminLogout(c *gin.Context) {
	// 在JWT方案中，登出通常由前端处理（删除本地存储的token）
	// 后端可以选择维护一个token黑名单，但这需要额外的存储
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "登出成功"})
}

// GetAdminInfo 获取管理员信息
func GetAdminInfo(c *gin.Context) {
	// 从上下文中获取管理员信息（需要AuthMiddleware配合）
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	// 将interface{}转换为int
	adminID, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 从数据库查找管理员
	admin, err := model.GetAdminByID(database.DB, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取管理员信息失败: " + err.Error()})
		return
	}

	if admin == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "管理员不存在"})
		return
	}

	// 返回管理员信息（不包含密码）
	adminInfo := struct {
		ID        int       `json:"id"`
		Username  string    `json:"username"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}{admin.ID, admin.Username, admin.CreatedAt, admin.UpdatedAt}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": adminInfo, "message": "success"})
}

// ChangePassword 修改管理员密码
func ChangePassword(c *gin.Context) {
	// 从上下文中获取管理员ID
	adminIDInterface, exists := c.Get("adminID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	adminID, ok := adminIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	var changePasswordReq struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&changePasswordReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 获取当前管理员信息
	admin, err := model.GetAdminByID(database.DB, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取管理员信息失败: " + err.Error()})
		return
	}

	if admin == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "管理员不存在"})
		return
	}

	// 验证旧密码
	if !utils.CheckPasswordHash(changePasswordReq.OldPassword, admin.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "原密码错误"})
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(changePasswordReq.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "加密密码失败: " + err.Error()})
		return
	}

	// 更新密码
	if err := model.UpdateAdminPassword(database.DB, adminID, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新密码失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "密码修改成功"})
}

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取token
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "请先登录"})
			c.Abort()
			return
		}

		// 移除Bearer前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// 使用JWT库验证token
		claims, err := utils.ParseToken(token)
		if err != nil {
			// 处理token验证失败的情况
			if err.Error() == "token is expired" {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "登录已过期，请重新登录"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的token"})
			}
			c.Abort()
			return
		}

		// 验证通过，将管理员信息存入上下文
		c.Set("adminID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// 轮播图管理API

// GetAllCarouselsForAdmin 获取所有轮播图（管理后台）
func GetAllCarouselsForAdmin(c *gin.Context) {
	// 从数据库获取所有轮播图（包括禁用状态）
	carousels, err := model.GetAllCarousels(database.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取轮播图失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": carousels, "message": "success"})
}

// UploadCarouselImage 上传轮播图图片到MinIO
func UploadCarouselImage(c *gin.Context) {
	// 检查是否有文件上传
	file, headers, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择要上传的图片: " + err.Error()})
		return
	}
	defer file.Close()

	if headers.Size > 5*1024*1024 { // 限制文件大小为5MB
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片大小不能超过5MB"})
		return
	}
	if !isImageFile(headers.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传JPG、PNG或GIF格式的图片"})
		return
	}

	// 上传图片到MinIO，轮播图存到carousels目录
	fileURL, err := utils.UploadFile("carousel", c.Request, "carousels")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "图片上传失败: " + err.Error()})
		return
	}

	// 写入数据库索引
	SaveImageIndex(fileURL, "carousels", headers.Filename, headers.Size, headers.Header.Get("Content-Type"))

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": map[string]string{"imageUrl": fileURL}, "message": "图片上传成功"})
}

// 判断是否为图片文件
func isImageFile(filename string) bool {
	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
	}

	// 获取文件扩展名
	extension := ""
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			extension = filename[i:]
			break
		}
	}

	return imageExtensions[strings.ToLower(extension)]
}

// UploadCategoryImage 上传分类图标到MinIO
func UploadCategoryImage(c *gin.Context) {
	// 检查是否有文件上传
	file, headers, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择要上传的图片: " + err.Error()})
		return
	}
	defer file.Close()

	if headers.Size > 5*1024*1024 { // 限制文件大小为5MB
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片大小不能超过5MB"})
		return
	}
	if !isImageFile(headers.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传JPG、PNG或GIF格式的图片"})
		return
	}

	// 上传图片到MinIO，分类图标存到categories目录
	fileURL, err := utils.UploadFile("category", c.Request, "categories")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "图片上传失败: " + err.Error()})
		return
	}

	// 写入数据库索引
	SaveImageIndex(fileURL, "categories", headers.Filename, headers.Size, headers.Header.Get("Content-Type"))

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": map[string]string{"url": fileURL}, "message": "图片上传成功"})
}

// CreateCarousel 创建轮播图
func CreateCarousel(c *gin.Context) {
	var carousel model.Carousel
	if err := c.ShouldBindJSON(&carousel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 创建轮播图
	if err := model.CreateCarousel(database.DB, &carousel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建轮播图失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": carousel, "message": "创建成功"})
}

// UpdateCarousel 更新轮播图
func UpdateCarousel(c *gin.Context) {
	id := c.Param("id")
	carouselID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的轮播图ID"})
		return
	}

	var updateData model.Carousel
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 查找轮播图
	carousel, err := model.GetCarouselByID(database.DB, carouselID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取轮播图失败: " + err.Error()})
		return
	}

	if carousel == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "轮播图不存在"})
		return
	}

	// 更新轮播图信息
	carousel.Image = updateData.Image
	carousel.Link = updateData.Link
	carousel.Sort = updateData.Sort
	carousel.Status = updateData.Status

	// 更新轮播图
	if err := model.UpdateCarousel(database.DB, carousel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新轮播图失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": carousel, "message": "更新成功"})
}

// DeleteCarousel 删除轮播图
func DeleteCarousel(c *gin.Context) {
	id := c.Param("id")
	carouselID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的轮播图ID"})
		return
	}

	// 查找轮播图
	carousel, err := model.GetCarouselByID(database.DB, carouselID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取轮播图失败: " + err.Error()})
		return
	}

	if carousel == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "轮播图不存在"})
		return
	}

	// 先保存图片URL，然后删除轮播图
	imageURL := carousel.Image
	if err := model.DeleteCarousel(database.DB, carouselID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除轮播图失败: " + err.Error()})
		return
	}

	// 异步删除MinIO中的图片，避免影响主流程
	go func(url string) {
		if err := utils.DeleteFile(url); err != nil {
			// 只记录错误，不影响主流程
			log.Printf("删除MinIO图片失败: %v", err)
		}
	}(imageURL)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ==================== 热销产品管理 ====================

// GetHotProducts 获取热销产品（小程序用）
func GetHotProducts(c *gin.Context) {
	products, err := model.GetHotProducts(database.DB)
	if err != nil {
		log.Printf("获取热销产品失败: %v", err)
		internalErrorResponse(c, "获取热销产品失败: "+err.Error())
		return
	}

	successResponse(c, products, "获取成功")
}

// GetAllHotProductsForAdmin 获取所有热销产品（管理后台）
func GetAllHotProductsForAdmin(c *gin.Context) {
	hotProducts, err := model.GetAllHotProducts(database.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取热销产品失败: " + err.Error()})
		return
	}

	// 获取所有商品信息
	allProducts, err := model.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品列表失败: " + err.Error()})
		return
	}

	// 创建商品映射
	productMap := make(map[int]*model.Product)
	for i := range allProducts {
		productMap[allProducts[i].ID] = &allProducts[i]
	}

	// 填充商品信息
	result := make([]gin.H, 0)
	for _, hp := range hotProducts {
		product := productMap[hp.ProductID]
		if product != nil {
			result = append(result, gin.H{
				"id":         hp.ID,
				"product_id": hp.ProductID,
				"product": gin.H{
					"id":          product.ID,
					"name":        product.Name,
					"description": product.Description,
					"images":      product.Images,
					"price":       product.Price,
					"category_id": product.CategoryID,
					"status":      product.Status,
				},
				"sort":       hp.Sort,
				"status":     hp.Status,
				"created_at": hp.CreatedAt,
				"updated_at": hp.UpdatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result, "message": "获取成功"})
}

// CreateHotProduct 创建热销产品关联
func CreateHotProduct(c *gin.Context) {
	var req struct {
		ProductID int `json:"product_id" binding:"required"`
		Sort      int `json:"sort"`
		Status    int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 检查商品是否存在
	product, err := model.GetProductByID(req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品信息失败: " + err.Error()})
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在"})
		return
	}

	// 检查是否已经存在
	existing, err := model.GetHotProductByProductID(database.DB, req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "检查热销产品失败: " + err.Error()})
		return
	}
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该商品已经是热销产品"})
		return
	}

	// 如果没有指定排序，自动设置为最大值+1
	if req.Sort == 0 {
		allHotProducts, err := model.GetAllHotProducts(database.DB)
		if err == nil && len(allHotProducts) > 0 {
			maxSort := 0
			for _, hp := range allHotProducts {
				if hp.Sort > maxSort {
					maxSort = hp.Sort
				}
			}
			req.Sort = maxSort + 1
		} else {
			req.Sort = 1
		}
	}

	// 默认状态为启用
	if req.Status == 0 {
		req.Status = 1
	}

	hp := &model.HotProduct{
		ProductID: req.ProductID,
		Sort:      req.Sort,
		Status:    req.Status,
	}

	if err := model.CreateHotProduct(database.DB, hp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建热销产品失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": hp, "message": "创建成功"})
}

// UpdateHotProduct 更新热销产品关联
func UpdateHotProduct(c *gin.Context) {
	id := c.Param("id")
	hotProductID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的热销产品ID"})
		return
	}

	var req struct {
		ProductID int `json:"product_id"`
		Sort      int `json:"sort"`
		Status    int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 查找热销产品
	hp, err := model.GetHotProductByID(database.DB, hotProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取热销产品失败: " + err.Error()})
		return
	}
	if hp == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "热销产品不存在"})
		return
	}

	// 如果更新了商品ID，检查新商品是否存在且不重复
	if req.ProductID > 0 && req.ProductID != hp.ProductID {
		product, err := model.GetProductByID(req.ProductID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品信息失败: " + err.Error()})
			return
		}
		if product == nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在"})
			return
		}

		// 检查新商品是否已经是热销产品
		existing, err := model.GetHotProductByProductID(database.DB, req.ProductID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "检查热销产品失败: " + err.Error()})
			return
		}
		if existing != nil && existing.ID != hotProductID {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该商品已经是热销产品"})
			return
		}
		hp.ProductID = req.ProductID
	}

	if req.Sort > 0 {
		hp.Sort = req.Sort
	}
	if req.Status > 0 {
		hp.Status = req.Status
	}

	if err := model.UpdateHotProduct(database.DB, hp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新热销产品失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": hp, "message": "更新成功"})
}

// DeleteHotProduct 删除热销产品关联
func DeleteHotProduct(c *gin.Context) {
	id := c.Param("id")
	hotProductID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的热销产品ID"})
		return
	}

	if err := model.DeleteHotProduct(database.DB, hotProductID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除热销产品失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// UpdateHotProductSort 批量更新热销产品排序
func UpdateHotProductSort(c *gin.Context) {
	var req struct {
		Items []struct {
			ID   int `json:"id" binding:"required"`
			Sort int `json:"sort" binding:"required"`
		} `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	items := make([]model.UpdateHotProductSortItem, len(req.Items))
	for i, item := range req.Items {
		items[i].ID = item.ID
		items[i].Sort = item.Sort
	}

	if err := model.UpdateHotProductSort(database.DB, items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新排序失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新排序成功"})
}
