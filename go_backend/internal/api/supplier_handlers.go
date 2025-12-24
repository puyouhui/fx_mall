package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go_backend/internal/database"
	"go_backend/internal/model"
	"go_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// calculateOrderItemCost 计算订单项的成本价
// 参数：specName - 规格名称, productSpecsJSON - 商品规格JSON字符串, quantity - 数量
// 返回：成本价 * 数量
func calculateOrderItemCost(specName string, productSpecsJSON sql.NullString, quantity int) float64 {
	costPrice := 0.0
	if productSpecsJSON.Valid {
		var specs []model.Spec
		if err := json.Unmarshal([]byte(productSpecsJSON.String), &specs); err == nil {
			// 优先根据规格名称匹配
			for _, spec := range specs {
				if spec.Name == specName {
					costPrice = spec.Cost
					break
				}
			}
			// 如果没找到匹配的规格，使用第一个规格的成本价
			if costPrice == 0 && len(specs) > 0 {
				costPrice = specs[0].Cost
			}
		}
	}
	return costPrice * float64(quantity)
}

// SupplierAuthMiddleware 供应商JWT认证中间件
func SupplierAuthMiddleware() gin.HandlerFunc {
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

		// 验证通过，将供应商信息存入上下文
		c.Set("supplierID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// ==================== 供应商管理 API ====================

// GetAllSuppliers 获取所有供应商（管理员）
func GetAllSuppliers(c *gin.Context) {
	suppliers, err := model.GetAllSuppliers(database.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取供应商列表失败: " + err.Error()})
		return
	}

	// 不返回密码字段
	var suppliersList []map[string]interface{}
	for _, supplier := range suppliers {
		supplierInfo := map[string]interface{}{
			"id":         supplier.ID,
			"name":       supplier.Name,
			"contact":    supplier.Contact,
			"phone":      supplier.Phone,
			"email":      supplier.Email,
			"address":    supplier.Address,
			"username":   supplier.Username,
			"status":     supplier.Status,
			"created_at": supplier.CreatedAt,
			"updated_at": supplier.UpdatedAt,
		}
		if supplier.Latitude != nil {
			supplierInfo["latitude"] = *supplier.Latitude
		}
		if supplier.Longitude != nil {
			supplierInfo["longitude"] = *supplier.Longitude
		}
		suppliersList = append(suppliersList, supplierInfo)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": suppliersList, "message": "success"})
}

// GetSupplierByID 根据ID获取供应商信息
func GetSupplierByID(c *gin.Context) {
	id := c.Param("id")
	supplierID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的供应商ID"})
		return
	}

	supplier, err := model.GetSupplierByID(database.DB, supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取供应商信息失败: " + err.Error()})
		return
	}

	if supplier == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "供应商不存在"})
		return
	}

	// 不返回密码
	supplierInfo := map[string]interface{}{
		"id":         supplier.ID,
		"name":       supplier.Name,
		"contact":    supplier.Contact,
		"phone":      supplier.Phone,
		"email":      supplier.Email,
		"address":    supplier.Address,
		"username":   supplier.Username,
		"status":     supplier.Status,
		"created_at": supplier.CreatedAt,
		"updated_at": supplier.UpdatedAt,
	}
	if supplier.Latitude != nil {
		supplierInfo["latitude"] = *supplier.Latitude
	}
	if supplier.Longitude != nil {
		supplierInfo["longitude"] = *supplier.Longitude
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": supplierInfo, "message": "success"})
}

// CreateSupplier 创建供应商
func CreateSupplier(c *gin.Context) {
	var supplierData struct {
		Name      string   `json:"name" binding:"required"`
		Contact   string   `json:"contact"`
		Phone     string   `json:"phone"`
		Email     string   `json:"email"`
		Address   string   `json:"address"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
		Username  string   `json:"username" binding:"required"`
		Password  string   `json:"password" binding:"required,min=6"`
		Status    int      `json:"status"`
	}

	if err := c.ShouldBindJSON(&supplierData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 检查用户名是否已存在
	existingSupplier, err := model.GetSupplierByUsername(database.DB, supplierData.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "检查用户名失败: " + err.Error()})
		return
	}
	if existingSupplier != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户名已存在"})
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(supplierData.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "加密密码失败: " + err.Error()})
		return
	}

	supplier := &model.Supplier{
		Name:      supplierData.Name,
		Contact:   supplierData.Contact,
		Phone:     supplierData.Phone,
		Email:     supplierData.Email,
		Address:   supplierData.Address,
		Latitude:  supplierData.Latitude,
		Longitude: supplierData.Longitude,
		Username:  supplierData.Username,
		Password:  hashedPassword,
		Status:    1, // 默认启用
	}

	if supplierData.Status > 0 {
		supplier.Status = supplierData.Status
	}

	if err := model.CreateSupplier(database.DB, supplier); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建供应商失败: " + err.Error()})
		return
	}

	// 返回创建结果（不包含密码）
	supplierInfo := map[string]interface{}{
		"id":         supplier.ID,
		"name":       supplier.Name,
		"contact":    supplier.Contact,
		"phone":      supplier.Phone,
		"email":      supplier.Email,
		"address":    supplier.Address,
		"username":   supplier.Username,
		"status":     supplier.Status,
		"created_at": supplier.CreatedAt,
		"updated_at": supplier.UpdatedAt,
	}
	if supplier.Latitude != nil {
		supplierInfo["latitude"] = *supplier.Latitude
	}
	if supplier.Longitude != nil {
		supplierInfo["longitude"] = *supplier.Longitude
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": supplierInfo, "message": "创建成功"})
}

// UpdateSupplier 更新供应商信息
func UpdateSupplier(c *gin.Context) {
	id := c.Param("id")
	supplierID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的供应商ID"})
		return
	}

	var updateData struct {
		Name      string   `json:"name" binding:"required"`
		Contact   string   `json:"contact"`
		Phone     string   `json:"phone"`
		Email     string   `json:"email"`
		Address   string   `json:"address"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
		Username  string   `json:"username" binding:"required"`
		Status    int      `json:"status"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 查找供应商
	supplier, err := model.GetSupplierByID(database.DB, supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取供应商失败: " + err.Error()})
		return
	}

	if supplier == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "供应商不存在"})
		return
	}

	// 如果用户名改变，检查新用户名是否已存在
	if updateData.Username != supplier.Username {
		existingSupplier, err := model.GetSupplierByUsername(database.DB, updateData.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "检查用户名失败: " + err.Error()})
			return
		}
		if existingSupplier != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户名已存在"})
			return
		}
	}

	// 更新供应商信息
	supplier.Name = updateData.Name
	supplier.Contact = updateData.Contact
	supplier.Phone = updateData.Phone
	supplier.Email = updateData.Email
	supplier.Address = updateData.Address
	supplier.Latitude = updateData.Latitude
	supplier.Longitude = updateData.Longitude
	supplier.Username = updateData.Username
	supplier.Status = updateData.Status

	if err := model.UpdateSupplier(database.DB, supplier); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新供应商失败: " + err.Error()})
		return
	}

	// 返回更新结果（不包含密码）
	supplierInfo := map[string]interface{}{
		"id":         supplier.ID,
		"name":       supplier.Name,
		"contact":    supplier.Contact,
		"phone":      supplier.Phone,
		"email":      supplier.Email,
		"address":    supplier.Address,
		"username":   supplier.Username,
		"status":     supplier.Status,
		"created_at": supplier.CreatedAt,
		"updated_at": supplier.UpdatedAt,
	}
	if supplier.Latitude != nil {
		supplierInfo["latitude"] = *supplier.Latitude
	}
	if supplier.Longitude != nil {
		supplierInfo["longitude"] = *supplier.Longitude
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": supplierInfo, "message": "更新成功"})
}

// DeleteSupplier 删除供应商（软删除）
func DeleteSupplier(c *gin.Context) {
	id := c.Param("id")
	supplierID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的供应商ID"})
		return
	}

	// 检查是否是自营供应商，禁止删除
	supplier, err := model.GetSupplierByID(database.DB, supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取供应商信息失败: " + err.Error()})
		return
	}
	if supplier == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "供应商不存在"})
		return
	}

	// 检查是否是自营供应商（通过用户名判断）
	if supplier.Username == "self_operated" {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "不能删除系统默认的'自营'供应商"})
		return
	}

	if err := model.DeleteSupplier(database.DB, supplierID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除供应商失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// SupplierLogin 供应商登录
func SupplierLogin(c *gin.Context) {
	var loginReq struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 根据用户名获取供应商信息
	supplier, err := model.GetSupplierByUsername(database.DB, loginReq.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "登录失败: " + err.Error()})
		return
	}

	if supplier == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	// 检查供应商状态
	if supplier.Status == 0 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "供应商账号已被禁用"})
		return
	}

	// 使用bcrypt验证密码
	if !utils.CheckPasswordHash(loginReq.Password, supplier.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	// 使用JWT库生成token
	token, err := utils.GenerateSupplierToken(supplier.Username, supplier.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成token失败: " + err.Error()})
		return
	}

	// 返回登录成功响应
	loginRes := struct {
		Token    string `json:"token"`
		Supplier struct {
			ID        int       `json:"id"`
			Name      string    `json:"name"`
			Username  string    `json:"username"`
			CreatedAt time.Time `json:"created_at"`
		} `json:"supplier"`
	}{
		Token: token,
		Supplier: struct {
			ID        int       `json:"id"`
			Name      string    `json:"name"`
			Username  string    `json:"username"`
			CreatedAt time.Time `json:"created_at"`
		}{supplier.ID, supplier.Name, supplier.Username, supplier.CreatedAt},
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": loginRes, "message": "登录成功"})
}

// GetSupplierProducts 供应商获取自己的商品列表
func GetSupplierProducts(c *gin.Context) {
	// 从上下文中获取供应商ID
	supplierIDInterface, exists := c.Get("supplierID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	supplierID, ok := supplierIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 获取分页参数
	pageNum := 1
	pageSize := 20
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			pageNum = p
		}
	}
	if sizeStr := c.Query("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	// 查询该供应商的商品
	var products []map[string]interface{}
	var total int

	// 计算偏移量
	offset := (pageNum - 1) * pageSize

	// 获取总数量
	countQuery := "SELECT COUNT(*) FROM products WHERE supplier_id = ? AND status = 1"
	if err := database.DB.QueryRow(countQuery, supplierID).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品数量失败: " + err.Error()})
		return
	}

	// 获取分页数据
	query := "SELECT id, name, description, original_price, price, category_id, supplier_id, is_special, images, specs, status, created_at, updated_at FROM products WHERE supplier_id = ? AND status = 1 ORDER BY id DESC LIMIT ? OFFSET ?"
	rows, err := database.DB.Query(query, supplierID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品列表失败: " + err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var product model.Product
		var imagesJSON, specsJSON string
		var dbPrice, dbOriginalPrice sql.NullFloat64
		var dbSupplierID sql.NullInt64

		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &dbOriginalPrice, &dbPrice, &product.CategoryID, &dbSupplierID, &product.IsSpecial, &imagesJSON, &specsJSON, &product.Status, &product.CreatedAt, &product.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "扫描商品数据失败: " + err.Error()})
			return
		}

		// 处理可空字段
		if dbPrice.Valid {
			product.Price = dbPrice.Float64
		}
		if dbOriginalPrice.Valid {
			product.OriginalPrice = dbOriginalPrice.Float64
		}
		if dbSupplierID.Valid {
			supplierIDVal := int(dbSupplierID.Int64)
			product.SupplierID = &supplierIDVal
		}

		// 解析JSON字符串到切片
		if err := json.Unmarshal([]byte(imagesJSON), &product.Images); err != nil {
			product.Images = []string{}
		}
		if err := json.Unmarshal([]byte(specsJSON), &product.Specs); err != nil {
			product.Specs = []model.Spec{}
		}

		// 计算规格数量和成本价范围
		specCount := len(product.Specs)
		minCostPrice := 0.0
		maxCostPrice := 0.0

		if specCount > 0 {
			// 找到所有有成本价的规格
			validCosts := make([]float64, 0)
			for _, spec := range product.Specs {
				if spec.Cost > 0 {
					validCosts = append(validCosts, spec.Cost)
				}
			}

			if len(validCosts) > 0 {
				minCostPrice = validCosts[0]
				maxCostPrice = validCosts[0]
				for _, cost := range validCosts {
					if cost < minCostPrice {
						minCostPrice = cost
					}
					if cost > maxCostPrice {
						maxCostPrice = cost
					}
				}
			}
		}

		// 如果没有规格或规格中没有成本价，使用商品价格字段（向后兼容）
		if minCostPrice == 0 && dbPrice.Valid {
			minCostPrice = dbPrice.Float64
			maxCostPrice = dbPrice.Float64
		}

		// 将成本价设置到Price字段，供前端显示（使用最低成本价）
		product.Price = minCostPrice

		// 构建返回的商品数据，包含规格数量和成本价范围
		productData := map[string]interface{}{
			"id":             product.ID,
			"name":           product.Name,
			"description":    product.Description,
			"price":          minCostPrice,
			"original_price": product.OriginalPrice,
			"category_id":    product.CategoryID,
			"supplier_id":    product.SupplierID,
			"is_special":     product.IsSpecial,
			"images":         product.Images,
			"specs":          product.Specs,
			"status":         product.Status,
			"created_at":     product.CreatedAt,
			"updated_at":     product.UpdatedAt,
			"spec_count":     specCount,
			"min_cost_price": minCostPrice,
			"max_cost_price": maxCostPrice,
		}

		products = append(products, productData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":      products,
			"total":     total,
			"page":      pageNum,
			"page_size": pageSize,
		},
		"message": "success",
	})
}

// GetSupplierProductDetail 供应商获取自己的商品详情
func GetSupplierProductDetail(c *gin.Context) {
	// 从上下文中获取供应商ID
	supplierIDInterface, exists := c.Get("supplierID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	supplierID, ok := supplierIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 获取商品ID
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供商品ID"})
		return
	}

	productID, err := strconv.Atoi(idStr)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "商品ID格式错误"})
		return
	}

	// 获取商品详情
	product, err := model.GetProductByID(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品详情失败: " + err.Error()})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在"})
		return
	}

	// 验证商品是否属于该供应商
	if product.SupplierID == nil || *product.SupplierID != supplierID {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权访问此商品"})
		return
	}

	// 计算成本价（取所有规格中成本价的最小值，如果没有规格则使用商品价格）
	costPrice := 0.0
	if len(product.Specs) > 0 {
		costPrice = product.Specs[0].Cost
		for _, spec := range product.Specs {
			if spec.Cost > 0 && (costPrice == 0 || spec.Cost < costPrice) {
				costPrice = spec.Cost
			}
		}
	}
	// 如果没有规格或规格中没有成本价，使用商品价格字段（向后兼容）
	if costPrice == 0 {
		costPrice = product.Price
	}

	// 构建返回数据（供应商后台显示成本价）
	responseData := map[string]interface{}{
		"id":             product.ID,
		"name":           product.Name,
		"description":    product.Description,
		"price":          costPrice, // 成本价
		"original_price": product.OriginalPrice,
		"category_id":    product.CategoryID,
		"is_special":     product.IsSpecial,
		"images":         product.Images,
		"specs":          product.Specs,
		"status":         product.Status,
		"created_at":     product.CreatedAt,
		"updated_at":     product.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    responseData,
		"message": "success",
	})
}

// GetSupplierOrders 供应商获取包含自己商品的订单列表
func GetSupplierOrders(c *gin.Context) {
	// 从上下文中获取供应商ID
	supplierIDInterface, exists := c.Get("supplierID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	supplierID, ok := supplierIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 获取分页参数
	pageNum := 1
	pageSize := 20
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			pageNum = p
		}
	}
	if sizeStr := c.Query("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	// 获取状态筛选参数
	statusFilter := c.Query("status") // pending_pickup, picked

	// 计算偏移量
	offset := (pageNum - 1) * pageSize

	// 构建状态条件
	statusCondition := ""
	if statusFilter == "pending_pickup" {
		// 待取货：pending_delivery 或 pending_pickup
		statusCondition = "AND o.status IN ('pending_delivery', 'pending_pickup')"
	} else if statusFilter == "picked" {
		// 已取货：delivering, delivered, paid
		statusCondition = "AND o.status IN ('delivering', 'delivered', 'paid')"
	}

	// 获取总数量：查询包含该供应商商品的订单数量
	countQuery := `
		SELECT COUNT(DISTINCT o.id)
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ?
		` + statusCondition
	var total int
	if err := database.DB.QueryRow(countQuery, supplierID).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单数量失败: " + err.Error()})
		return
	}

	// 获取订单列表（先获取订单基本信息，成本价在Go代码中计算）
	query := `
		SELECT DISTINCT
			o.id,
			o.order_number,
			o.status,
			o.created_at,
			o.updated_at,
			COALESCE(u.user_code, '') as user_code
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		LEFT JOIN mini_app_users u ON o.user_id = u.id
		WHERE p.supplier_id = ?
		` + statusCondition + `
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := database.DB.Query(query, supplierID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单列表失败: " + err.Error()})
		return
	}
	defer rows.Close()

	orders := make([]map[string]interface{}, 0)
	for rows.Next() {
		var orderID int
		var orderNumber, orderStatus, userCode string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&orderID, &orderNumber, &orderStatus, &createdAt, &updatedAt, &userCode); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "扫描订单数据失败: " + err.Error()})
			return
		}

		// 获取该订单中该供应商的商品数量和总成本
		itemsQuery := `
			SELECT 
				oi.id,
				oi.spec_name,
				oi.quantity,
				p.specs as product_specs
			FROM order_items oi
			INNER JOIN products p ON oi.product_id = p.id
			WHERE oi.order_id = ? AND p.supplier_id = ?
		`
		itemsRows, err := database.DB.Query(itemsQuery, orderID, supplierID)
		itemCount := 0
		totalCost := 0.0
		if err == nil {
			defer itemsRows.Close()
			for itemsRows.Next() {
				var itemID int
				var specName string
				var quantity int
				var productSpecsJSON sql.NullString

				if err := itemsRows.Scan(&itemID, &specName, &quantity, &productSpecsJSON); err == nil {
					itemCount++
					// 计算成本价
					costPrice := 0.0
					if productSpecsJSON.Valid {
						var specs []model.Spec
						if err := json.Unmarshal([]byte(productSpecsJSON.String), &specs); err == nil {
							// 优先根据规格名称匹配
							for _, spec := range specs {
								if spec.Name == specName {
									costPrice = spec.Cost
									break
								}
							}
							// 如果没找到匹配的规格，使用第一个规格的成本价
							if costPrice == 0 && len(specs) > 0 {
								costPrice = specs[0].Cost
							}
						}
					}
					totalCost += costPrice * float64(quantity)
				}
			}
		}

		// 映射订单状态为供应商视角的状态
		supplierStatus := "待取货"
		if orderStatus == "delivering" || orderStatus == "delivered" || orderStatus == "paid" {
			supplierStatus = "已取货"
		}

		orderData := map[string]interface{}{
			"id":           orderID,
			"order_number": orderNumber,
			"user_code":    userCode,
			"status":       supplierStatus,
			"item_count":   itemCount,
			"total_cost":   totalCost,
			"created_at":   createdAt,
			"updated_at":   updatedAt,
		}

		orders = append(orders, orderData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":      orders,
			"total":     total,
			"page":      pageNum,
			"page_size": pageSize,
		},
		"message": "success",
	})
}

// GetSupplierOrderDetail 供应商获取订单详情（只包含该供应商的商品）
func GetSupplierOrderDetail(c *gin.Context) {
	// 从上下文中获取供应商ID
	supplierIDInterface, exists := c.Get("supplierID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	supplierID, ok := supplierIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 获取订单ID
	orderIDStr := c.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	// 获取订单基本信息（包含客户编号）
	var orderNumber, orderStatus, userCode string
	var createdAt, updatedAt time.Time
	orderQuery := `
		SELECT o.order_number, o.status, o.created_at, o.updated_at, COALESCE(u.user_code, '') as user_code
		FROM orders o
		LEFT JOIN mini_app_users u ON o.user_id = u.id
		WHERE o.id = ?
	`
	if err := database.DB.QueryRow(orderQuery, orderID).Scan(&orderNumber, &orderStatus, &createdAt, &updatedAt, &userCode); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单信息失败: " + err.Error()})
		return
	}

	// 映射订单状态为供应商视角的状态
	supplierStatus := "待取货"
	if orderStatus == "delivering" || orderStatus == "delivered" || orderStatus == "paid" {
		supplierStatus = "已取货"
	}

	// 获取该订单中该供应商的商品明细
	itemsQuery := `
		SELECT 
			oi.id,
			oi.product_id,
			oi.product_name,
			oi.spec_name,
			oi.quantity,
			oi.image,
			p.specs as product_specs
		FROM order_items oi
		INNER JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = ? AND p.supplier_id = ?
		ORDER BY oi.id ASC
	`

	itemsRows, err := database.DB.Query(itemsQuery, orderID, supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单明细失败: " + err.Error()})
		return
	}
	defer itemsRows.Close()

	items := make([]map[string]interface{}, 0)
	totalCost := 0.0

	for itemsRows.Next() {
		var itemID, productID, quantity int
		var productName, specName, image string
		var productSpecsJSON sql.NullString

		if err := itemsRows.Scan(&itemID, &productID, &productName, &specName, &quantity, &image, &productSpecsJSON); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "扫描订单明细失败: " + err.Error()})
			return
		}

		// 计算成本价
		costPrice := 0.0
		if productSpecsJSON.Valid {
			var specs []model.Spec
			if err := json.Unmarshal([]byte(productSpecsJSON.String), &specs); err == nil {
				// 优先根据规格名称匹配
				for _, spec := range specs {
					if spec.Name == specName {
						costPrice = spec.Cost
						break
					}
				}
				// 如果没找到匹配的规格，使用第一个规格的成本价
				if costPrice == 0 && len(specs) > 0 {
					costPrice = specs[0].Cost
				}
			}
		}

		itemCost := costPrice * float64(quantity)
		totalCost += itemCost

		itemData := map[string]interface{}{
			"id":           itemID,
			"product_id":   productID,
			"product_name": productName,
			"spec_name":    specName,
			"quantity":     quantity,
			"cost_price":   costPrice,
			"item_cost":    itemCost,
			"image":        image,
		}

		items = append(items, itemData)
	}

	// 构建返回数据
	responseData := map[string]interface{}{
		"id":           orderID,
		"order_number": orderNumber,
		"user_code":    userCode,
		"status":       supplierStatus,
		"item_count":   len(items),
		"total_cost":   totalCost,
		"items":        items,
		"created_at":   createdAt,
		"updated_at":   updatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    responseData,
		"message": "success",
	})
}

// GetSupplierDashboard 获取供应商数据总览
func GetSupplierDashboard(c *gin.Context) {
	// 从上下文中获取供应商ID
	supplierIDInterface, exists := c.Get("supplierID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	supplierID, ok := supplierIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 获取时间范围参数：today, 7days, month, year
	period := c.Query("period")
	if period == "" {
		period = "today" // 默认今日
	}

	// 根据时间范围构建日期条件
	var dateCondition string
	switch period {
	case "today":
		dateCondition = "DATE(o.created_at) = CURDATE()"
	case "7days":
		dateCondition = "DATE(o.created_at) >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)"
	case "month":
		dateCondition = "YEAR(o.created_at) = YEAR(CURDATE()) AND MONTH(o.created_at) = MONTH(CURDATE())"
	case "year":
		dateCondition = "YEAR(o.created_at) = YEAR(CURDATE())"
	default:
		dateCondition = "DATE(o.created_at) = CURDATE()"
	}

	// 1. 我供应的商品数量（不受时间范围影响）
	var totalProducts int
	productCountQuery := "SELECT COUNT(*) FROM products WHERE supplier_id = ? AND status = 1"
	if err := database.DB.QueryRow(productCountQuery, supplierID).Scan(&totalProducts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品数量失败: " + err.Error()})
		return
	}

	// 2. 订单数量（已取货的订单：状态为 delivering, delivered, paid）
	var orderCount int
	orderCountQuery := `
		SELECT COUNT(DISTINCT o.id) 
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('delivering', 'delivered', 'paid')
		AND ` + dateCondition
	if err := database.DB.QueryRow(orderCountQuery, supplierID).Scan(&orderCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单数量失败: " + err.Error()})
		return
	}

	// 3. 货物件数（已取货订单的商品总件数）
	var itemCount int
	itemCountQuery := `
		SELECT COALESCE(SUM(oi.quantity), 0)
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('delivering', 'delivered', 'paid')
		AND ` + dateCondition
	if err := database.DB.QueryRow(itemCountQuery, supplierID).Scan(&itemCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取货物件数失败: " + err.Error()})
		return
	}

	// 4. 已成交总额（已取货之后的订单总成本：状态为 delivering, delivered, paid）
	totalSales := 0.0
	totalSalesQuery := `
		SELECT oi.spec_name, oi.quantity, p.specs as product_specs
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('delivering', 'delivered', 'paid')
		AND ` + dateCondition
	rows, err := database.DB.Query(totalSalesQuery, supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取已成交总额失败: " + err.Error()})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var specName string
		var quantity int
		var productSpecsJSON sql.NullString
		if err := rows.Scan(&specName, &quantity, &productSpecsJSON); err == nil {
			totalSales += calculateOrderItemCost(specName, productSpecsJSON, quantity)
		}
	}

	// 5. 待备货订单数量（该供应商的产品，已经下单，配送员还没有完成取货的：状态为 pending_delivery 或 pending_pickup）
	var pendingOrderCount int
	pendingOrderCountQuery := `
		SELECT COUNT(DISTINCT o.id)
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('pending_delivery', 'pending_pickup')
	`
	if err := database.DB.QueryRow(pendingOrderCountQuery, supplierID).Scan(&pendingOrderCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取待备货订单数量失败: " + err.Error()})
		return
	}

	// 6. 待备货商品数（待备货订单的商品总件数）
	var pendingItemCount int
	pendingItemCountQuery := `
		SELECT COALESCE(SUM(oi.quantity), 0)
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('pending_delivery', 'pending_pickup')
	`
	if err := database.DB.QueryRow(pendingItemCountQuery, supplierID).Scan(&pendingItemCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取待备货商品数失败: " + err.Error()})
		return
	}

	// 7. 待备货金额（待备货订单的总成本）
	pendingTotal := 0.0
	pendingAmountQuery := `
		SELECT oi.spec_name, oi.quantity, p.specs as product_specs
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('pending_delivery', 'pending_pickup')
	`
	pendingRows, err := database.DB.Query(pendingAmountQuery, supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取待备货金额失败: " + err.Error()})
		return
	}
	defer pendingRows.Close()
	for pendingRows.Next() {
		var specName string
		var quantity int
		var productSpecsJSON sql.NullString
		if err := pendingRows.Scan(&specName, &quantity, &productSpecsJSON); err == nil {
			pendingTotal += calculateOrderItemCost(specName, productSpecsJSON, quantity)
		}
	}

	// 8. 平均订单金额
	avgOrderAmount := 0.0
	if orderCount > 0 {
		avgOrderAmount = totalSales / float64(orderCount)
	}

	// 9. 热销商品TOP 5（根据销量，查询全部时间，使用成本价计算）
	type TopProduct struct {
		ProductID   int     `json:"product_id"`
		ProductName string  `json:"product_name"`
		TotalQty    int     `json:"total_qty"`
		TotalAmount float64 `json:"total_amount"`
	}
	topProducts := make([]TopProduct, 0)
	// 先查询所有订单项，然后在Go代码中按商品分组计算
	topProductsQuery := `
		SELECT p.id, p.name, oi.quantity, oi.spec_name, p.specs as product_specs
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('delivering', 'delivered', 'paid')
		ORDER BY p.id
	`
	topRows, err := database.DB.Query(topProductsQuery, supplierID)
	if err == nil {
		defer topRows.Close()
		productMap := make(map[int]*TopProduct)
		for topRows.Next() {
			var productID int
			var productName, specName string
			var quantity int
			var productSpecsJSON sql.NullString
			if err := topRows.Scan(&productID, &productName, &quantity, &specName, &productSpecsJSON); err == nil {
				if product, exists := productMap[productID]; exists {
					product.TotalQty += quantity
					product.TotalAmount += calculateOrderItemCost(specName, productSpecsJSON, quantity)
				} else {
					productMap[productID] = &TopProduct{
						ProductID:   productID,
						ProductName: productName,
						TotalQty:    quantity,
						TotalAmount: calculateOrderItemCost(specName, productSpecsJSON, quantity),
					}
				}
			}
		}
		// 转换为切片并按销量排序
		for _, product := range productMap {
			topProducts = append(topProducts, *product)
		}
		// 按销量降序排序
		for i := 0; i < len(topProducts)-1; i++ {
			for j := i + 1; j < len(topProducts); j++ {
				if topProducts[i].TotalQty < topProducts[j].TotalQty {
					topProducts[i], topProducts[j] = topProducts[j], topProducts[i]
				}
			}
		}
		// 只取前5个
		if len(topProducts) > 5 {
			topProducts = topProducts[:5]
		}
	}

	// 10. 最近订单列表（最近5个订单，查询全部时间，使用和订单管理页面相同的逻辑）
	type RecentOrder struct {
		OrderID     int       `json:"order_id"`
		OrderNumber string    `json:"order_number"`
		UserCode    string    `json:"user_code"`
		Status      string    `json:"status"`
		ItemCount   int       `json:"item_count"`
		TotalCost   float64   `json:"total_cost"`
		CreatedAt   time.Time `json:"created_at"`
	}
	recentOrders := make([]RecentOrder, 0)

	// 获取最近5个订单（使用和订单管理页面相同的查询逻辑）
	recentOrdersQuery := `
		SELECT DISTINCT
			o.id,
			o.order_number,
			o.status,
			o.created_at,
			COALESCE(u.user_code, '') as user_code
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		LEFT JOIN mini_app_users u ON o.user_id = u.id
		WHERE p.supplier_id = ?
		ORDER BY o.created_at DESC
		LIMIT 5
	`
	recentRows, err := database.DB.Query(recentOrdersQuery, supplierID)
	if err == nil {
		defer recentRows.Close()
		for recentRows.Next() {
			var orderID int
			var orderNumber, orderStatus, userCode string
			var createdAt time.Time

			if err := recentRows.Scan(&orderID, &orderNumber, &orderStatus, &createdAt, &userCode); err == nil {
				// 获取该订单中该供应商的商品数量和总成本
				itemsQuery := `
					SELECT 
						oi.id,
						oi.spec_name,
						oi.quantity,
						p.specs as product_specs
					FROM order_items oi
					INNER JOIN products p ON oi.product_id = p.id
					WHERE oi.order_id = ? AND p.supplier_id = ?
				`
				itemsRows, err := database.DB.Query(itemsQuery, orderID, supplierID)
				itemCount := 0
				totalCost := 0.0
				if err == nil {
					defer itemsRows.Close()
					for itemsRows.Next() {
						var itemID int
						var specName string
						var quantity int
						var productSpecsJSON sql.NullString

						if err := itemsRows.Scan(&itemID, &specName, &quantity, &productSpecsJSON); err == nil {
							itemCount++
							// 计算成本价
							costPrice := 0.0
							if productSpecsJSON.Valid {
								var specs []model.Spec
								if err := json.Unmarshal([]byte(productSpecsJSON.String), &specs); err == nil {
									// 优先根据规格名称匹配
									for _, spec := range specs {
										if spec.Name == specName {
											costPrice = spec.Cost
											break
										}
									}
									// 如果没找到匹配的规格，使用第一个规格的成本价
									if costPrice == 0 && len(specs) > 0 {
										costPrice = specs[0].Cost
									}
								}
							}
							totalCost += costPrice * float64(quantity)
						}
					}
				}

				// 映射订单状态为供应商视角的状态
				supplierStatus := "待取货"
				if orderStatus == "delivering" || orderStatus == "delivered" || orderStatus == "paid" {
					supplierStatus = "已取货"
				}

				recentOrders = append(recentOrders, RecentOrder{
					OrderID:     orderID,
					OrderNumber: orderNumber,
					UserCode:    userCode,
					Status:      supplierStatus,
					ItemCount:   itemCount,
					TotalCost:   totalCost,
					CreatedAt:   createdAt,
				})
			}
		}
	}

	// 11. 计算对比数据（与上一个时间段对比，仅用于今日和本月，使用成本价计算）
	var previousPeriodOrderCount int
	previousAmount := 0.0
	if period == "today" {
		// 对比昨天
		countQuery := `
			SELECT COUNT(DISTINCT o.id)
			FROM orders o
			INNER JOIN order_items oi ON o.id = oi.order_id
			INNER JOIN products p ON oi.product_id = p.id
			WHERE p.supplier_id = ? 
			AND o.status IN ('delivering', 'delivered', 'paid')
			AND DATE(o.created_at) = DATE_SUB(CURDATE(), INTERVAL 1 DAY)
		`
		database.DB.QueryRow(countQuery, supplierID).Scan(&previousPeriodOrderCount)

		// 计算昨天的成本总额
		amountQuery := `
			SELECT oi.spec_name, oi.quantity, p.specs as product_specs
			FROM orders o
			INNER JOIN order_items oi ON o.id = oi.order_id
			INNER JOIN products p ON oi.product_id = p.id
			WHERE p.supplier_id = ? 
			AND o.status IN ('delivering', 'delivered', 'paid')
			AND DATE(o.created_at) = DATE_SUB(CURDATE(), INTERVAL 1 DAY)
		`
		prevRows, err := database.DB.Query(amountQuery, supplierID)
		if err == nil {
			defer prevRows.Close()
			for prevRows.Next() {
				var specName string
				var quantity int
				var productSpecsJSON sql.NullString
				if err := prevRows.Scan(&specName, &quantity, &productSpecsJSON); err == nil {
					previousAmount += calculateOrderItemCost(specName, productSpecsJSON, quantity)
				}
			}
		}
	} else if period == "month" {
		// 对比上个月
		countQuery := `
			SELECT COUNT(DISTINCT o.id)
			FROM orders o
			INNER JOIN order_items oi ON o.id = oi.order_id
			INNER JOIN products p ON oi.product_id = p.id
			WHERE p.supplier_id = ? 
			AND o.status IN ('delivering', 'delivered', 'paid')
			AND YEAR(o.created_at) = YEAR(DATE_SUB(CURDATE(), INTERVAL 1 MONTH))
			AND MONTH(o.created_at) = MONTH(DATE_SUB(CURDATE(), INTERVAL 1 MONTH))
		`
		database.DB.QueryRow(countQuery, supplierID).Scan(&previousPeriodOrderCount)

		// 计算上个月的成本总额
		amountQuery := `
			SELECT oi.spec_name, oi.quantity, p.specs as product_specs
			FROM orders o
			INNER JOIN order_items oi ON o.id = oi.order_id
			INNER JOIN products p ON oi.product_id = p.id
			WHERE p.supplier_id = ? 
			AND o.status IN ('delivering', 'delivered', 'paid')
			AND YEAR(o.created_at) = YEAR(DATE_SUB(CURDATE(), INTERVAL 1 MONTH))
			AND MONTH(o.created_at) = MONTH(DATE_SUB(CURDATE(), INTERVAL 1 MONTH))
		`
		prevRows, err := database.DB.Query(amountQuery, supplierID)
		if err == nil {
			defer prevRows.Close()
			for prevRows.Next() {
				var specName string
				var quantity int
				var productSpecsJSON sql.NullString
				if err := prevRows.Scan(&specName, &quantity, &productSpecsJSON); err == nil {
					previousAmount += calculateOrderItemCost(specName, productSpecsJSON, quantity)
				}
			}
		}
	}

	// 计算增长率
	orderGrowthRate := 0.0
	if previousPeriodOrderCount > 0 {
		orderGrowthRate = (float64(orderCount-previousPeriodOrderCount) / float64(previousPeriodOrderCount)) * 100
	} else if orderCount > 0 {
		orderGrowthRate = 100.0
	}

	amountGrowthRate := 0.0
	if previousAmount > 0 {
		amountGrowthRate = ((totalSales - previousAmount) / previousAmount) * 100
	} else if totalSales > 0 {
		amountGrowthRate = 100.0
	}

	// 12. 最近15日销售情况统计（已取货完成的订单，包含所有15天，没有数据的日期填充0）
	type DailySales struct {
		Date        string  `json:"date"`
		OrderCount  int     `json:"order_count"`
		ItemCount   int     `json:"item_count"`
		SalesAmount float64 `json:"sales_amount"`
	}

	var dailySales []DailySales

	// 先查询有数据的日期
	salesMap := make(map[string]DailySales)

	// 查询所有订单项，然后在Go代码中按日期分组计算成本价
	dailySalesQuery := `
		SELECT 
			DATE(o.created_at) as date,
			oi.quantity,
			oi.spec_name,
			p.specs as product_specs
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('delivering', 'delivered', 'paid')
		AND DATE(o.created_at) >= DATE_SUB(CURDATE(), INTERVAL 14 DAY)
		ORDER BY date ASC
	`
	dailyRows, err := database.DB.Query(dailySalesQuery, supplierID)
	if err != nil {
		// 如果查询出错，记录错误但不影响其他数据返回
		log.Printf("查询每日销售数据失败: %v, supplierID: %d", err, supplierID)
	} else {
		defer dailyRows.Close()
		rowCount := 0
		for dailyRows.Next() {
			var dateTime sql.NullTime
			var quantity int
			var specName string
			var productSpecsJSON sql.NullString
			if err := dailyRows.Scan(&dateTime, &quantity, &specName, &productSpecsJSON); err != nil {
				continue
			}
			if !dateTime.Valid {
				continue
			}
			// 将日期格式化为 YYYY-MM-DD
			dateStr := dateTime.Time.Format("2006-01-02")

			// 初始化日期数据
			if _, exists := salesMap[dateStr]; !exists {
				salesMap[dateStr] = DailySales{
					Date:        dateStr,
					OrderCount:  0,
					ItemCount:   0,
					SalesAmount: 0.0,
				}
			}

			// 更新数据
			daily := salesMap[dateStr]
			daily.ItemCount += quantity
			daily.SalesAmount += calculateOrderItemCost(specName, productSpecsJSON, quantity)
			salesMap[dateStr] = daily
			rowCount++
		}

		// 统计每个日期的订单数量（需要单独查询）
		orderCountQuery := `
			SELECT DATE(o.created_at) as date, COUNT(DISTINCT o.id) as order_count
			FROM orders o
			INNER JOIN order_items oi ON o.id = oi.order_id
			INNER JOIN products p ON oi.product_id = p.id
			WHERE p.supplier_id = ? 
			AND o.status IN ('delivering', 'delivered', 'paid')
			AND DATE(o.created_at) >= DATE_SUB(CURDATE(), INTERVAL 14 DAY)
			GROUP BY DATE(o.created_at)
		`
		orderCountRows, err := database.DB.Query(orderCountQuery, supplierID)
		if err == nil {
			defer orderCountRows.Close()
			for orderCountRows.Next() {
				var dateTime sql.NullTime
				var orderCount int
				if err := orderCountRows.Scan(&dateTime, &orderCount); err == nil && dateTime.Valid {
					dateStr := dateTime.Time.Format("2006-01-02")
					if daily, exists := salesMap[dateStr]; exists {
						daily.OrderCount = orderCount
						salesMap[dateStr] = daily
					}
				}
			}
		}
	}

	// 生成完整的15天数据（从14天前到今天，共15天）
	dailySales = make([]DailySales, 0)
	for i := 14; i >= 0; i-- {
		dateStr := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if daily, exists := salesMap[dateStr]; exists {
			dailySales = append(dailySales, daily)
		} else {
			// 如果该日期没有数据，填充0值
			dailySales = append(dailySales, DailySales{
				Date:        dateStr,
				OrderCount:  0,
				ItemCount:   0,
				SalesAmount: 0.0,
			})
		}
	}

	// 返回数据总览
	dashboardData := map[string]interface{}{
		"total_products":       totalProducts,
		"order_count":          orderCount,
		"item_count":           itemCount,
		"total_sales_amount":   totalSales,
		"avg_order_amount":     avgOrderAmount,
		"pending_order_count":  pendingOrderCount,
		"pending_item_count":   pendingItemCount,
		"pending_amount":       pendingTotal,
		"top_products":         topProducts,
		"recent_orders":        recentOrders,
		"order_growth_rate":    orderGrowthRate,
		"amount_growth_rate":   amountGrowthRate,
		"previous_order_count": previousPeriodOrderCount,
		"previous_amount":      previousAmount,
		"daily_sales":          dailySales,
		"period":               period,
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": dashboardData, "message": "获取成功"})
}

// ==================== 货物管理 API ====================

// GetTodayGoodsStats 获取今日货物统计
func GetTodayGoodsStats(c *gin.Context) {
	// 从上下文中获取供应商ID
	supplierIDInterface, exists := c.Get("supplierID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	supplierID, ok := supplierIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 1. 我供应的商品数量（不受时间范围影响）
	var totalProducts int
	productCountQuery := "SELECT COUNT(*) FROM products WHERE supplier_id = ? AND status = 1"
	if err := database.DB.QueryRow(productCountQuery, supplierID).Scan(&totalProducts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品数量失败: " + err.Error()})
		return
	}

	// 2. 待备货统计（状态为 pending_delivery 或 pending_pickup，不限制日期，因为所有待备货的都需要准备）
	var pendingItemCount int
	var pendingGoodsCount int
	pendingItemCountQuery := `
		SELECT 
			COALESCE(SUM(oi.quantity), 0),
			(SELECT COUNT(DISTINCT CONCAT(oi2.product_id, '-', oi2.spec_name))
			 FROM orders o2
			 INNER JOIN order_items oi2 ON o2.id = oi2.order_id
			 INNER JOIN products p2 ON oi2.product_id = p2.id
			 WHERE p2.supplier_id = ? 
			 AND o2.status IN ('pending_delivery', 'pending_pickup'))
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('pending_delivery', 'pending_pickup')
	`
	if err := database.DB.QueryRow(pendingItemCountQuery, supplierID, supplierID).Scan(&pendingItemCount, &pendingGoodsCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取待备货统计失败: " + err.Error()})
		return
	}

	// 计算待备货金额
	pendingTotal := 0.0
	pendingAmountQuery := `
		SELECT oi.spec_name, oi.quantity, p.specs as product_specs
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('pending_delivery', 'pending_pickup')
	`
	pendingRows, err := database.DB.Query(pendingAmountQuery, supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取待备货金额失败: " + err.Error()})
		return
	}
	defer pendingRows.Close()
	for pendingRows.Next() {
		var specName string
		var quantity int
		var productSpecsJSON sql.NullString
		if err := pendingRows.Scan(&specName, &quantity, &productSpecsJSON); err == nil {
			pendingTotal += calculateOrderItemCost(specName, productSpecsJSON, quantity)
		}
	}

	// 3. 今日已取货统计（状态为 delivering, delivered, paid）
	var pickedItemCount int
	var pickedGoodsCount int
	pickedItemCountQuery := `
		SELECT 
			COALESCE(SUM(oi.quantity), 0),
			(SELECT COUNT(DISTINCT CONCAT(oi2.product_id, '-', oi2.spec_name))
			 FROM orders o2
			 INNER JOIN order_items oi2 ON o2.id = oi2.order_id
			 INNER JOIN products p2 ON oi2.product_id = p2.id
			 WHERE p2.supplier_id = ? 
			 AND o2.status IN ('delivering', 'delivered', 'paid')
			 AND DATE(o2.created_at) = CURDATE())
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('delivering', 'delivered', 'paid')
		AND DATE(o.created_at) = CURDATE()
	`
	if err := database.DB.QueryRow(pickedItemCountQuery, supplierID, supplierID).Scan(&pickedItemCount, &pickedGoodsCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取已取货统计失败: " + err.Error()})
		return
	}

	// 计算已取货金额
	pickedTotal := 0.0
	pickedAmountQuery := `
		SELECT oi.spec_name, oi.quantity, p.specs as product_specs
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('delivering', 'delivered', 'paid')
		AND DATE(o.created_at) = CURDATE()
	`
	pickedRows, err := database.DB.Query(pickedAmountQuery, supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取已取货金额失败: " + err.Error()})
		return
	}
	defer pickedRows.Close()
	for pickedRows.Next() {
		var specName string
		var quantity int
		var productSpecsJSON sql.NullString
		if err := pickedRows.Scan(&specName, &quantity, &productSpecsJSON); err == nil {
			pickedTotal += calculateOrderItemCost(specName, productSpecsJSON, quantity)
		}
	}

	// 4. 总统计
	totalItemCount := pendingItemCount + pickedItemCount
	totalGoodsCount := pendingGoodsCount + pickedGoodsCount
	totalAmount := pendingTotal + pickedTotal

	// 返回统计数据
	statsData := map[string]interface{}{
		"total": map[string]interface{}{
			"total_products":    totalProducts,
			"total_item_count":  totalItemCount,
			"total_amount":      totalAmount,
			"total_goods_count": totalGoodsCount,
		},
		"pending": map[string]interface{}{
			"item_count":   pendingItemCount,
			"goods_count":  pendingGoodsCount,
			"total_amount": pendingTotal,
		},
		"picked": map[string]interface{}{
			"item_count":   pickedItemCount,
			"goods_count":  pickedGoodsCount,
			"total_amount": pickedTotal,
		},
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": statsData, "message": "获取成功"})
}

// GetTodayPendingGoods 获取今日待备货货物列表
func GetTodayPendingGoods(c *gin.Context) {
	// 从上下文中获取供应商ID
	supplierIDInterface, exists := c.Get("supplierID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	supplierID, ok := supplierIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 查询待备货货物（不限制日期，因为所有待备货的都需要准备），按商品和规格分组
	query := `
		SELECT 
			p.id as product_id,
			oi.product_name,
			oi.spec_name,
			oi.image,
			SUM(oi.quantity) as total_quantity,
			p.specs as product_specs
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('pending_delivery', 'pending_pickup')
		GROUP BY p.id, oi.product_name, oi.spec_name, oi.image, p.specs
		ORDER BY p.id, oi.spec_name
	`

	rows, err := database.DB.Query(query, supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取待备货货物列表失败: " + err.Error()})
		return
	}
	defer rows.Close()

	goodsList := make([]map[string]interface{}, 0)
	for rows.Next() {
		var productID int
		var productName, specName string
		var image sql.NullString
		var totalQuantity int
		var productSpecsJSON sql.NullString

		if err := rows.Scan(&productID, &productName, &specName, &image, &totalQuantity, &productSpecsJSON); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "扫描货物数据失败: " + err.Error()})
			return
		}

		// 计算成本价
		costPrice := 0.0
		if productSpecsJSON.Valid {
			var specs []model.Spec
			if err := json.Unmarshal([]byte(productSpecsJSON.String), &specs); err == nil {
				// 优先根据规格名称匹配
				for _, spec := range specs {
					if spec.Name == specName {
						costPrice = spec.Cost
						break
					}
				}
				// 如果没找到匹配的规格，使用第一个规格的成本价
				if costPrice == 0 && len(specs) > 0 {
					costPrice = specs[0].Cost
				}
			}
		}

		totalCost := costPrice * float64(totalQuantity)

		goodsData := map[string]interface{}{
			"product_id":   productID,
			"product_name": productName,
			"spec_name":    specName,
			"image":        image.String,
			"quantity":     totalQuantity,
			"cost_price":   costPrice,
			"total_cost":   totalCost,
		}

		goodsList = append(goodsList, goodsData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list": goodsList,
		},
		"message": "获取成功",
	})
}

// GetTodayPickedGoods 获取今日已取货货物列表
func GetTodayPickedGoods(c *gin.Context) {
	// 从上下文中获取供应商ID
	supplierIDInterface, exists := c.Get("supplierID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	supplierID, ok := supplierIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 查询今日已取货货物，按商品和规格分组
	query := `
		SELECT 
			p.id as product_id,
			oi.product_name,
			oi.spec_name,
			oi.image,
			SUM(oi.quantity) as total_quantity,
			p.specs as product_specs
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('delivering', 'delivered', 'paid')
		AND DATE(o.created_at) = CURDATE()
		GROUP BY p.id, oi.product_name, oi.spec_name, oi.image, p.specs
		ORDER BY p.id, oi.spec_name
	`

	rows, err := database.DB.Query(query, supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取已取货货物列表失败: " + err.Error()})
		return
	}
	defer rows.Close()

	goodsList := make([]map[string]interface{}, 0)
	for rows.Next() {
		var productID int
		var productName, specName string
		var image sql.NullString
		var totalQuantity int
		var productSpecsJSON sql.NullString

		if err := rows.Scan(&productID, &productName, &specName, &image, &totalQuantity, &productSpecsJSON); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "扫描货物数据失败: " + err.Error()})
			return
		}

		// 计算成本价
		costPrice := 0.0
		if productSpecsJSON.Valid {
			var specs []model.Spec
			if err := json.Unmarshal([]byte(productSpecsJSON.String), &specs); err == nil {
				// 优先根据规格名称匹配
				for _, spec := range specs {
					if spec.Name == specName {
						costPrice = spec.Cost
						break
					}
				}
				// 如果没找到匹配的规格，使用第一个规格的成本价
				if costPrice == 0 && len(specs) > 0 {
					costPrice = specs[0].Cost
				}
			}
		}

		totalCost := costPrice * float64(totalQuantity)

		goodsData := map[string]interface{}{
			"product_id":   productID,
			"product_name": productName,
			"spec_name":    specName,
			"image":        image.String,
			"quantity":     totalQuantity,
			"cost_price":   costPrice,
			"total_cost":   totalCost,
		}

		goodsList = append(goodsList, goodsData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list": goodsList,
		},
		"message": "获取成功",
	})
}

// ==================== 历史记录 API ====================

// GetHistoryByDate 获取历史记录列表（按天）
func GetHistoryByDate(c *gin.Context) {
	// 从上下文中获取供应商ID
	supplierIDInterface, exists := c.Get("supplierID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	supplierID, ok := supplierIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 获取分页参数
	pageNum := 1
	pageSize := 20
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			pageNum = p
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// 获取日期筛选参数
	dateFilter := ""
	if dateStr := c.Query("date"); dateStr != "" {
		dateFilter = fmt.Sprintf("AND DATE(o.created_at) = '%s'", dateStr)
	}

	// 计算偏移量
	offset := (pageNum - 1) * pageSize

	// 获取总数量：查询有该供应商商品的订单日期数量
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT DATE(o.created_at))
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ?
		%s
	`, dateFilter)
	var total int
	if err := database.DB.QueryRow(countQuery, supplierID).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取历史记录数量失败: " + err.Error()})
		return
	}

	// 获取历史记录列表（按日期分组）
	query := fmt.Sprintf(`
		SELECT 
			daily_stats.order_date,
			daily_stats.pending_item_count,
			daily_stats.picked_item_count,
			daily_stats.total_item_count,
			(SELECT COUNT(DISTINCT CONCAT(oi2.product_id, '-', oi2.spec_name))
			 FROM orders o2
			 INNER JOIN order_items oi2 ON o2.id = oi2.order_id
			 INNER JOIN products p2 ON oi2.product_id = p2.id
			 WHERE p2.supplier_id = ?
			 AND DATE(o2.created_at) = daily_stats.order_date
			 %s) as total_goods_count
		FROM (
			SELECT 
				DATE(o.created_at) as order_date,
				SUM(CASE WHEN o.status IN ('pending_delivery', 'pending_pickup') THEN oi.quantity ELSE 0 END) as pending_item_count,
				SUM(CASE WHEN o.status IN ('delivering', 'delivered', 'paid') THEN oi.quantity ELSE 0 END) as picked_item_count,
				SUM(oi.quantity) as total_item_count
			FROM orders o
			INNER JOIN order_items oi ON o.id = oi.order_id
			INNER JOIN products p ON oi.product_id = p.id
			WHERE p.supplier_id = ?
			%s
			GROUP BY DATE(o.created_at)
		) as daily_stats
		ORDER BY daily_stats.order_date DESC
		LIMIT ? OFFSET ?
	`, dateFilter, dateFilter)

	rows, err := database.DB.Query(query, supplierID, supplierID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取历史记录列表失败: " + err.Error()})
		return
	}
	defer rows.Close()

	historyList := make([]map[string]interface{}, 0)
	for rows.Next() {
		var orderDate time.Time
		var pendingItemCount, pickedItemCount, totalItemCount, totalGoodsCount int

		if err := rows.Scan(&orderDate, &pendingItemCount, &pickedItemCount, &totalItemCount, &totalGoodsCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "扫描历史记录数据失败: " + err.Error()})
			return
		}

		// 计算该日期的总金额和已取货金额
		dateStr := orderDate.Format("2006-01-02")
		totalAmount := 0.0
		pickedAmount := 0.0

		// 计算总金额
		amountQuery := `
			SELECT oi.spec_name, oi.quantity, p.specs as product_specs
			FROM orders o
			INNER JOIN order_items oi ON o.id = oi.order_id
			INNER JOIN products p ON oi.product_id = p.id
			WHERE p.supplier_id = ? 
			AND DATE(o.created_at) = ?
		`
		amountRows, err := database.DB.Query(amountQuery, supplierID, dateStr)
		if err == nil {
			defer amountRows.Close()
			for amountRows.Next() {
				var specName string
				var quantity int
				var productSpecsJSON sql.NullString
				if err := amountRows.Scan(&specName, &quantity, &productSpecsJSON); err == nil {
					totalAmount += calculateOrderItemCost(specName, productSpecsJSON, quantity)
				}
			}
		}

		// 计算已取货金额（只计算状态为delivering, delivered, paid的订单）
		pickedAmountQuery := `
			SELECT oi.spec_name, oi.quantity, p.specs as product_specs
			FROM orders o
			INNER JOIN order_items oi ON o.id = oi.order_id
			INNER JOIN products p ON oi.product_id = p.id
			WHERE p.supplier_id = ? 
			AND DATE(o.created_at) = ?
			AND o.status IN ('delivering', 'delivered', 'paid')
		`
		pickedAmountRows, err := database.DB.Query(pickedAmountQuery, supplierID, dateStr)
		if err == nil {
			defer pickedAmountRows.Close()
			for pickedAmountRows.Next() {
				var specName string
				var quantity int
				var productSpecsJSON sql.NullString
				if err := pickedAmountRows.Scan(&specName, &quantity, &productSpecsJSON); err == nil {
					pickedAmount += calculateOrderItemCost(specName, productSpecsJSON, quantity)
				}
			}
		}

		historyData := map[string]interface{}{
			"date":               dateStr,
			"pending_item_count": pendingItemCount,
			"picked_item_count":  pickedItemCount,
			"total_item_count":   totalItemCount,
			"total_goods_count":  totalGoodsCount,
			"total_amount":       totalAmount,
			"picked_amount":      pickedAmount,
		}

		historyList = append(historyList, historyData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":      historyList,
			"total":     total,
			"page":      pageNum,
			"page_size": pageSize,
		},
		"message": "获取成功",
	})
}

// GetHistoryDetail 获取某天的历史详情
func GetHistoryDetail(c *gin.Context) {
	// 从上下文中获取供应商ID
	supplierIDInterface, exists := c.Get("supplierID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	supplierID, ok := supplierIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部服务器错误"})
		return
	}

	// 获取日期参数
	dateStr := c.Param("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "日期参数不能为空"})
		return
	}

	// 验证日期格式
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "日期格式错误，应为 YYYY-MM-DD"})
		return
	}

	// 获取该日期的统计信息
	var pendingItemCount, pickedItemCount, totalItemCount, totalGoodsCount int
	statsQuery := `
		SELECT 
			SUM(CASE WHEN o.status IN ('pending_delivery', 'pending_pickup') THEN oi.quantity ELSE 0 END) as pending_item_count,
			SUM(CASE WHEN o.status IN ('delivering', 'delivered', 'paid') THEN oi.quantity ELSE 0 END) as picked_item_count,
			SUM(oi.quantity) as total_item_count,
			(SELECT COUNT(DISTINCT CONCAT(oi2.product_id, '-', oi2.spec_name))
			 FROM orders o2
			 INNER JOIN order_items oi2 ON o2.id = oi2.order_id
			 INNER JOIN products p2 ON oi2.product_id = p2.id
			 WHERE p2.supplier_id = ? 
			 AND DATE(o2.created_at) = ?) as total_goods_count
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND DATE(o.created_at) = ?
	`
	if err := database.DB.QueryRow(statsQuery, supplierID, dateStr, supplierID, dateStr).Scan(&pendingItemCount, &pickedItemCount, &totalItemCount, &totalGoodsCount); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "该日期无历史记录"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取历史统计失败: " + err.Error()})
		return
	}

	// 计算总金额
	totalAmount := 0.0
	amountQuery := `
		SELECT oi.spec_name, oi.quantity, p.specs as product_specs
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND DATE(o.created_at) = ?
	`
	amountRows, err := database.DB.Query(amountQuery, supplierID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取历史金额失败: " + err.Error()})
		return
	}
	defer amountRows.Close()
	for amountRows.Next() {
		var specName string
		var quantity int
		var productSpecsJSON sql.NullString
		if err := amountRows.Scan(&specName, &quantity, &productSpecsJSON); err == nil {
			totalAmount += calculateOrderItemCost(specName, productSpecsJSON, quantity)
		}
	}

	// 获取待备货货物明细（只显示该日期创建的待备货订单，与统计信息保持一致）
	pendingGoodsQuery := `
		SELECT 
			p.id as product_id,
			oi.product_name,
			oi.spec_name,
			oi.image,
			SUM(oi.quantity) as total_quantity,
			p.specs as product_specs
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('pending_delivery', 'pending_pickup')
		AND DATE(o.created_at) = ?
		GROUP BY p.id, oi.product_name, oi.spec_name, oi.image, p.specs
		ORDER BY p.id, oi.spec_name
	`
	pendingGoodsRows, err := database.DB.Query(pendingGoodsQuery, supplierID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取待备货货物明细失败: " + err.Error()})
		return
	}
	defer pendingGoodsRows.Close()

	pendingGoods := make([]map[string]interface{}, 0)
	for pendingGoodsRows.Next() {
		var productID int
		var productName, specName string
		var image sql.NullString
		var totalQuantity int
		var productSpecsJSON sql.NullString

		if err := pendingGoodsRows.Scan(&productID, &productName, &specName, &image, &totalQuantity, &productSpecsJSON); err != nil {
			continue
		}

		// 计算成本价
		costPrice := 0.0
		if productSpecsJSON.Valid {
			var specs []model.Spec
			if err := json.Unmarshal([]byte(productSpecsJSON.String), &specs); err == nil {
				for _, spec := range specs {
					if spec.Name == specName {
						costPrice = spec.Cost
						break
					}
				}
				if costPrice == 0 && len(specs) > 0 {
					costPrice = specs[0].Cost
				}
			}
		}

		totalCost := costPrice * float64(totalQuantity)

		goodsData := map[string]interface{}{
			"product_id":   productID,
			"product_name": productName,
			"spec_name":    specName,
			"image":        image.String,
			"quantity":     totalQuantity,
			"cost_price":   costPrice,
			"total_cost":   totalCost,
		}

		pendingGoods = append(pendingGoods, goodsData)
	}

	// 获取已取货货物明细
	pickedGoodsQuery := `
		SELECT 
			p.id as product_id,
			oi.product_name,
			oi.spec_name,
			oi.image,
			SUM(oi.quantity) as total_quantity,
			p.specs as product_specs
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('delivering', 'delivered', 'paid')
		AND DATE(o.created_at) = ?
		GROUP BY p.id, oi.product_name, oi.spec_name, oi.image, p.specs
		ORDER BY p.id, oi.spec_name
	`
	pickedGoodsRows, err := database.DB.Query(pickedGoodsQuery, supplierID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取已取货货物明细失败: " + err.Error()})
		return
	}
	defer pickedGoodsRows.Close()

	pickedGoods := make([]map[string]interface{}, 0)
	pickedAmount := 0.0
	for pickedGoodsRows.Next() {
		var productID int
		var productName, specName string
		var image sql.NullString
		var totalQuantity int
		var productSpecsJSON sql.NullString

		if err := pickedGoodsRows.Scan(&productID, &productName, &specName, &image, &totalQuantity, &productSpecsJSON); err != nil {
			continue
		}

		// 计算成本价
		costPrice := 0.0
		if productSpecsJSON.Valid {
			var specs []model.Spec
			if err := json.Unmarshal([]byte(productSpecsJSON.String), &specs); err == nil {
				for _, spec := range specs {
					if spec.Name == specName {
						costPrice = spec.Cost
						break
					}
				}
				if costPrice == 0 && len(specs) > 0 {
					costPrice = specs[0].Cost
				}
			}
		}

		totalCost := costPrice * float64(totalQuantity)
		pickedAmount += totalCost

		goodsData := map[string]interface{}{
			"product_id":   productID,
			"product_name": productName,
			"spec_name":    specName,
			"image":        image.String,
			"quantity":     totalQuantity,
			"cost_price":   costPrice,
			"total_cost":   totalCost,
		}

		pickedGoods = append(pickedGoods, goodsData)
	}

	// 返回历史详情
	detailData := map[string]interface{}{
		"date":               dateStr,
		"pending_item_count": pendingItemCount,
		"picked_item_count":  pickedItemCount,
		"total_item_count":   totalItemCount,
		"total_goods_count":  totalGoodsCount,
		"total_amount":       totalAmount,
		"picked_amount":      pickedAmount,
		"pending_goods":      pendingGoods,
		"picked_goods":       pickedGoods,
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": detailData, "message": "获取成功"})
}

// GetMobilePendingGoods 移动端获取待备货货物列表（不需要token，通过供应商账号和ID参数验证）
func GetMobilePendingGoods(c *gin.Context) {
	// 获取URL参数
	supplierUsername := c.Query("name") // name参数是供应商账号
	supplierIDStr := c.Query("ID")
	if supplierIDStr == "" {
		supplierIDStr = c.Query("id") // 兼容小写id
	}

	// 验证参数
	if supplierUsername == "" || supplierIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少必要参数：name（供应商账号）和 ID"})
		return
	}

	// 转换供应商ID
	supplierID, err := strconv.Atoi(supplierIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "ID参数格式错误"})
		return
	}

	// 验证供应商是否存在，并且账号和ID匹配
	var actualSupplierID int
	var supplierStatus int
	checkQuery := "SELECT id, status FROM suppliers WHERE username = ? AND id = ?"
	if err := database.DB.QueryRow(checkQuery, supplierUsername, supplierID).Scan(&actualSupplierID, &supplierStatus); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "供应商账号或ID验证失败"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "验证供应商失败: " + err.Error()})
		return
	}

	// 检查供应商状态
	if supplierStatus != 1 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "供应商账号已禁用"})
		return
	}

	// 查询待备货货物（不限制日期，因为所有待备货的都需要准备），按商品和规格分组
	query := `
		SELECT 
			p.id as product_id,
			oi.product_name,
			oi.spec_name,
			oi.image,
			SUM(oi.quantity) as total_quantity,
			p.specs as product_specs
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE p.supplier_id = ? 
		AND o.status IN ('pending_delivery', 'pending_pickup')
		GROUP BY p.id, oi.product_name, oi.spec_name, oi.image, p.specs
		ORDER BY p.id, oi.spec_name
	`

	rows, err := database.DB.Query(query, supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取待备货货物列表失败: " + err.Error()})
		return
	}
	defer rows.Close()

	goodsList := make([]map[string]interface{}, 0)
	for rows.Next() {
		var productID int
		var productName, specName string
		var image sql.NullString
		var totalQuantity int
		var productSpecsJSON sql.NullString

		if err := rows.Scan(&productID, &productName, &specName, &image, &totalQuantity, &productSpecsJSON); err != nil {
			continue
		}

		// 计算成本价
		costPrice := 0.0
		if productSpecsJSON.Valid {
			var specs []model.Spec
			if err := json.Unmarshal([]byte(productSpecsJSON.String), &specs); err == nil {
				// 优先根据规格名称匹配
				for _, spec := range specs {
					if spec.Name == specName {
						costPrice = spec.Cost
						break
					}
				}
				// 如果没找到匹配的规格，使用第一个规格的成本价
				if costPrice == 0 && len(specs) > 0 {
					costPrice = specs[0].Cost
				}
			}
		}

		totalCost := costPrice * float64(totalQuantity)

		goodsData := map[string]interface{}{
			"product_id":   productID,
			"product_name": productName,
			"spec_name":    specName,
			"image":        image.String,
			"quantity":     totalQuantity,
			"cost_price":   costPrice,
			"total_cost":   totalCost,
		}

		goodsList = append(goodsList, goodsData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    gin.H{"list": goodsList},
		"message": "获取成功",
	})
}
