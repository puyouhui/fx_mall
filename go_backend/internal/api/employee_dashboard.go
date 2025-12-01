package api

import (
	"net/http"

	"go_backend/internal/database"
	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetEmployeeDashboard 获取当前员工的首页概览信息
// 包含：基础信息、关联客户数量、订单总数、待配送订单数量等
func GetEmployeeDashboard(c *gin.Context) {
	employee, ok := getEmployeeFromContext(c)
	if !ok {
		return
	}

	// 基础信息
	data := gin.H{
		"id":            employee.ID,
		"employee_code": employee.EmployeeCode,
		"name":          employee.Name,
		"phone":         employee.Phone,
		"is_delivery":   employee.IsDelivery,
		"is_sales":      employee.IsSales,
	}

	// 销售员相关统计
	if employee.IsSales {
		customerCount, _ := model.CountCustomersByEmployeeCode(employee.EmployeeCode)
		totalOrders, pendingDeliveryOrders, todayOrders, _ := model.CountOrdersBySalesCode(employee.EmployeeCode)

		// 预览前几个客户（最多3个）
		customersPreview := make([]map[string]interface{}, 0)
		if customers, err := model.GetCustomersByEmployeeCode(employee.EmployeeCode); err == nil {
			for i, cst := range customers {
				if i >= 3 {
					break
				}
				item := map[string]interface{}{}
				if name, ok := cst["name"]; ok {
					item["name"] = name
				}
				if phone, ok := cst["phone"]; ok {
					item["phone"] = phone
				}
				if userCode, ok := cst["user_code"]; ok {
					item["user_code"] = userCode
				}
				if addr, ok := cst["default_address"]; ok {
					item["default_address"] = addr
				}
				customersPreview = append(customersPreview, item)
			}
		}

		data["customer_count"] = customerCount
		data["customers_preview"] = customersPreview
		data["order_total"] = totalOrders
		data["order_pending_delivery"] = pendingDeliveryOrders
		data["order_today"] = todayOrders
	}

	// 配送员相关统计（全局待配送订单数量）
	if employee.IsDelivery {
		var pendingTotal int
		err := database.DB.QueryRow(`
			SELECT COUNT(*) FROM orders 
			WHERE status IN ('pending_delivery', 'pending')
		`).Scan(&pendingTotal)
		if err == nil {
			data["delivery_pending_total"] = pendingTotal
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    data,
	})
}
