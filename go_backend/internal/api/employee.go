package api

import (
	"fmt"
	"net/http"

	"go_backend/internal/model"
	"go_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetEmployees 获取员工列表
func GetEmployees(c *gin.Context) {
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)
	keyword := c.Query("keyword")

	employees, total, err := model.GetEmployees(pageNum, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取员工列表失败: " + err.Error()})
		return
	}

	// 为每个销售员添加绑定的客户信息
	employeesWithCustomers := make([]map[string]interface{}, 0)
	for _, emp := range employees {
		empData := map[string]interface{}{
			"id":            emp.ID,
			"employee_code": emp.EmployeeCode,
			"phone":         emp.Phone,
			"name":          emp.Name,
			"is_delivery":   emp.IsDelivery,
			"is_sales":      emp.IsSales,
			"status":        emp.Status,
			"created_at":    emp.CreatedAt,
			"updated_at":    emp.UpdatedAt,
		}

		// 如果是销售员，获取绑定的客户信息
		if emp.IsSales {
			customerCount, _ := model.CountCustomersByEmployeeCode(emp.EmployeeCode)
			customers, _ := model.GetCustomersByEmployeeCode(emp.EmployeeCode)
			empData["customer_count"] = customerCount
			empData["customers"] = customers
		} else {
			empData["customer_count"] = 0
			empData["customers"] = []interface{}{}
		}

		employeesWithCustomers = append(employeesWithCustomers, empData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"data":  employeesWithCustomers,
		"total": total,
	})
}

// GetSalesEmployees 获取所有销售员列表（用于下拉选择）
func GetSalesEmployees(c *gin.Context) {
	employees, err := model.GetSalesEmployees()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取销售员列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    employees,
	})
}

// GetEmployee 获取员工详情
func GetEmployee(c *gin.Context) {
	idStr := c.Param("id")
	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "员工ID格式错误"})
		return
	}

	employee, err := model.GetEmployeeByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取员工详情失败: " + err.Error()})
		return
	}

	if employee == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "员工不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    employee,
	})
}

type createEmployeeRequest struct {
	Phone      string `json:"phone" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Name       string `json:"name"`
	IsDelivery bool   `json:"is_delivery"`
	IsSales    bool   `json:"is_sales"`
}

// CreateEmployee 创建员工
func CreateEmployee(c *gin.Context) {
	var req createEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 验证手机号格式
	if !utils.IsValidPhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "手机号格式不正确"})
		return
	}

	// 验证密码长度
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码长度不能少于6位"})
		return
	}

	// 至少需要选择一种角色
	if !req.IsDelivery && !req.IsSales {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "至少需要选择一种角色（配送员或销售员）"})
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败: " + err.Error()})
		return
	}

	// 创建员工
	employee, err := model.CreateEmployee(req.Phone, hashedPassword, req.Name, req.IsDelivery, req.IsSales)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建员工失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    employee,
	})
}

type updateEmployeeRequest struct {
	Phone      string `json:"phone"`
	Password   string `json:"password"`
	Name       string `json:"name"`
	IsDelivery *bool  `json:"is_delivery"`
	IsSales    *bool  `json:"is_sales"`
	Status     *bool  `json:"status"`
}

// UpdateEmployee 更新员工
func UpdateEmployee(c *gin.Context) {
	idStr := c.Param("id")
	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "员工ID格式错误"})
		return
	}

	var req updateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 构建更新数据
	updateData := make(map[string]interface{})

	if req.Phone != "" {
		if !utils.IsValidPhone(req.Phone) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "手机号格式不正确"})
			return
		}
		updateData["phone"] = req.Phone
	}

	if req.Password != "" {
		if len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码长度不能少于6位"})
			return
		}
		// 加密密码
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败: " + err.Error()})
			return
		}
		updateData["password"] = hashedPassword
	}

	if req.Name != "" {
		updateData["name"] = req.Name
	}

	if req.IsDelivery != nil {
		updateData["is_delivery"] = *req.IsDelivery
	}

	if req.IsSales != nil {
		updateData["is_sales"] = *req.IsSales
	}

	if req.Status != nil {
		updateData["status"] = *req.Status
	}

	// 如果更新角色，至少需要选择一种角色
	if req.IsDelivery != nil || req.IsSales != nil {
		// 先获取当前员工信息
		employee, err := model.GetEmployeeByID(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取员工信息失败: " + err.Error()})
			return
		}
		if employee == nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "员工不存在"})
			return
		}

		isDelivery := employee.IsDelivery
		isSales := employee.IsSales

		if req.IsDelivery != nil {
			isDelivery = *req.IsDelivery
		}
		if req.IsSales != nil {
			isSales = *req.IsSales
		}

		if !isDelivery && !isSales {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "至少需要选择一种角色（配送员或销售员）"})
			return
		}
	}

	// 更新员工
	err = model.UpdateEmployee(id, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新员工失败: " + err.Error()})
		return
	}

	// 获取更新后的员工信息
	employee, err := model.GetEmployeeByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取更新后的员工信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    employee,
	})
}

// DeleteEmployee 删除员工
func DeleteEmployee(c *gin.Context) {
	idStr := c.Param("id")
	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "员工ID格式错误"})
		return
	}

	err = model.DeleteEmployee(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除员工失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

