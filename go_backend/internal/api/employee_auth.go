package api

import (
	"net/http"

	"go_backend/internal/model"
	"go_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// EmployeeLoginRequest 员工登录请求
type EmployeeLoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// EmployeeLogin 员工登录
func EmployeeLogin(c *gin.Context) {
	var req EmployeeLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 验证手机号格式
	if !utils.IsValidPhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "手机号格式不正确"})
		return
	}

	// 获取员工信息
	employee, err := model.GetEmployeeByPhone(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "登录失败: " + err.Error()})
		return
	}

	if employee == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "手机号或密码错误"})
		return
	}

	// 检查员工状态
	if !employee.Status {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "账号已被禁用"})
		return
	}

	// 验证密码
	if !utils.CheckPasswordHash(req.Password, employee.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "手机号或密码错误"})
		return
	}

	// 生成token
	token, err := utils.GenerateEmployeeToken(employee.ID, employee.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成登录凭证失败: " + err.Error()})
		return
	}

	// 返回员工信息（不包含密码）
	employeeData := map[string]interface{}{
		"id":            employee.ID,
		"employee_code": employee.EmployeeCode,
		"phone":         employee.Phone,
		"name":          employee.Name,
		"is_delivery":   employee.IsDelivery,
		"is_sales":      employee.IsSales,
		"status":        employee.Status,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"token":    token,
			"employee": employeeData,
		},
	})
}

// EmployeeAuthMiddleware 员工认证中间件
func EmployeeAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
			c.Abort()
			return
		}

		claims, err := utils.ParseEmployeeToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "登录状态已失效，请重新登录"})
			c.Abort()
			return
		}

		// 获取员工信息并验证状态
		employee, err := model.GetEmployeeByID(claims.EmployeeID)
		if err != nil || employee == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "员工信息不存在"})
			c.Abort()
			return
		}

		if !employee.Status {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "账号已被禁用"})
			c.Abort()
			return
		}

		// 将员工信息存入上下文
		c.Set("employee_id", employee.ID)
		c.Set("employee", employee)
		c.Next()
	}
}

// getEmployeeFromContext 从上下文获取员工信息
func getEmployeeFromContext(c *gin.Context) (*model.Employee, bool) {
	employee, exists := c.Get("employee")
	if !exists {
		return nil, false
	}
	emp, ok := employee.(*model.Employee)
	return emp, ok
}

// GetEmployeeInfo 获取当前员工信息
func GetEmployeeInfo(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	employeeData := map[string]interface{}{
		"id":            employee.ID,
		"employee_code": employee.EmployeeCode,
		"phone":         employee.Phone,
		"name":          employee.Name,
		"is_delivery":   employee.IsDelivery,
		"is_sales":      employee.IsSales,
		"status":        employee.Status,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    employeeData,
	})
}

