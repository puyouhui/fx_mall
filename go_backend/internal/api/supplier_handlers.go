package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go_backend/internal/database"
	"go_backend/internal/model"
	"go_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

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
	var products []model.Product
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

		products = append(products, product)
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
