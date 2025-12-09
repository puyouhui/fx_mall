package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go_backend/internal/database"
	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetAllDeliveryRecordsForAdmin 获取所有配送记录（后台管理）
func GetAllDeliveryRecordsForAdmin(c *gin.Context) {
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 20)
	keyword := c.Query("keyword")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	records, total, err := model.GetAllDeliveryRecords(pageNum, pageSize, keyword, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取配送记录列表失败: " + err.Error()})
		return
	}

	if len(records) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"list":  []map[string]interface{}{},
				"total": 0,
			},
			"message": "获取成功",
		})
		return
	}

	// 获取订单编号、配送费、创建时间
	orderIDs := make([]int, 0, len(records))
	employeeCodes := make([]string, 0)
	employeeCodeSet := make(map[string]bool)
	for _, record := range records {
		orderIDs = append(orderIDs, record.OrderID)
		if !employeeCodeSet[record.DeliveryEmployeeCode] {
			employeeCodes = append(employeeCodes, record.DeliveryEmployeeCode)
			employeeCodeSet[record.DeliveryEmployeeCode] = true
		}
	}

	log.Printf("配送记录查询: 找到 %d 条记录, 订单IDs: %v\n", len(records), orderIDs)

	orderInfo := make(map[int]map[string]interface{})
	if len(orderIDs) > 0 {
		// 批量查询订单信息
		placeholders := ""
		args := make([]interface{}, len(orderIDs))
		for i, id := range orderIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args[i] = id
		}

		query := "SELECT id, order_number, delivery_fee, delivery_fee_calculation, created_at, status FROM orders WHERE id IN (" + placeholders + ")"
		rows, err := database.DB.Query(query, args...)
		if err != nil {
			log.Printf("查询订单信息失败: %v", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var id int
				var orderNumber string
				var deliveryFee sql.NullFloat64
				var deliveryFeeCalcJSON sql.NullString
				var createdAt time.Time
				var status string
				if err := rows.Scan(&id, &orderNumber, &deliveryFee, &deliveryFeeCalcJSON, &createdAt, &status); err != nil {
					log.Printf("扫描订单信息失败: %v", err)
					continue
				}

				// 优先从 delivery_fee_calculation 中获取配送员实际所得（rider_payable_fee）
				// 如果不存在，则使用客户支付的配送费（delivery_fee）
				riderPayableFee := 0.0
				if deliveryFeeCalcJSON.Valid && deliveryFeeCalcJSON.String != "" {
					var calcResult model.DeliveryFeeCalculationResult
					if err := json.Unmarshal([]byte(deliveryFeeCalcJSON.String), &calcResult); err == nil {
						riderPayableFee = calcResult.RiderPayableFee
						log.Printf("订单 %d: 从delivery_fee_calculation获取配送员实际所得=%.2f\n", id, riderPayableFee)
					} else {
						log.Printf("警告: 订单 %d 的delivery_fee_calculation解析失败: %v\n", id, err)
						// 解析失败，使用客户支付的配送费
						if deliveryFee.Valid {
							riderPayableFee = deliveryFee.Float64
						}
					}
				} else {
					// 没有delivery_fee_calculation，使用客户支付的配送费
					if deliveryFee.Valid {
						riderPayableFee = deliveryFee.Float64
					} else {
						log.Printf("警告: 订单 %d 的配送费为NULL，且没有delivery_fee_calculation\n", id)
					}
				}

				orderInfo[id] = map[string]interface{}{
					"order_number": orderNumber,
					"delivery_fee": riderPayableFee, // 使用配送员实际所得
					"customer_fee": deliveryFee,     // 保留客户支付的配送费（用于调试）
					"created_at":   createdAt,
					"status":       status,
				}
				log.Printf("订单 %d 信息: 订单编号=%s, 配送员实际所得=%.2f, 客户支付配送费=%.2f (Valid=%v), 状态=%s\n",
					id, orderNumber, riderPayableFee, func() float64 {
						if deliveryFee.Valid {
							return deliveryFee.Float64
						}
						return 0.0
					}(), deliveryFee.Valid, status)
			}
		}
	}

	// 批量获取配送员信息
	employees, _ := model.GetEmployeesByEmployeeCodes(employeeCodes)
	employeeMap := make(map[string]string)
	for code, emp := range employees {
		if emp != nil {
			employeeMap[code] = emp.Name
		}
	}

	// 构建返回数据
	list := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		orderData := orderInfo[record.OrderID]
		item := map[string]interface{}{
			"id":                     record.ID,
			"order_id":               record.OrderID,
			"order_number":           "",
			"delivery_employee_code": record.DeliveryEmployeeCode,
			"delivery_employee_name": employeeMap[record.DeliveryEmployeeCode],
			"delivery_fee":           0.0,
			"status":                 "",
			"order_created_at":       nil,
			"completed_at":           record.CompletedAt,
			"created_at":             record.CreatedAt,
			"updated_at":             record.UpdatedAt,
		}

		if orderData != nil {
			if on, ok := orderData["order_number"].(string); ok {
				item["order_number"] = on
			}
			if df, ok := orderData["delivery_fee"].(float64); ok {
				item["delivery_fee"] = df
			} else {
				log.Printf("警告: 订单 %d 的配送费类型转换失败, 值: %v, 类型: %T\n", record.OrderID, orderData["delivery_fee"], orderData["delivery_fee"])
			}
			if s, ok := orderData["status"].(string); ok {
				item["status"] = s
			} else {
				log.Printf("警告: 订单 %d 的状态类型转换失败, 值: %v, 类型: %T\n", record.OrderID, orderData["status"], orderData["status"])
			}
			if ca, ok := orderData["created_at"].(time.Time); ok {
				item["order_created_at"] = ca
			}
		} else {
			log.Printf("警告: 订单 %d 的信息未找到\n", record.OrderID)
		}

		if record.ProductImageURL != nil {
			item["product_image_url"] = *record.ProductImageURL
		}
		if record.DoorplateImageURL != nil {
			item["doorplate_image_url"] = *record.DoorplateImageURL
		}

		list = append(list, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  list,
			"total": total,
		},
		"message": "获取成功",
	})
}

// GetDeliveryRecordByIDForAdmin 获取配送记录详情（后台管理）
func GetDeliveryRecordByIDForAdmin(c *gin.Context) {
	idStr := c.Param("id")
	_, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的记录ID"})
		return
	}

	// TODO: 添加根据ID获取配送记录的函数
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
	})
}

// GetDeliveryRecordByOrderIDForAdmin 根据订单ID获取配送记录（后台管理）
func GetDeliveryRecordByOrderIDForAdmin(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的订单ID"})
		return
	}

	record, err := model.GetDeliveryRecordByOrderID(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取配送记录失败: " + err.Error()})
		return
	}

	if record == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"data":    nil,
			"message": "该订单暂无配送记录",
		})
		return
	}

	// 获取订单信息（包括配送费计算结果）
	var orderNumber string
	var customerDeliveryFee sql.NullFloat64
	var deliveryFeeCalcJSON sql.NullString
	var orderCreatedAt time.Time
	err = database.DB.QueryRow("SELECT order_number, delivery_fee, delivery_fee_calculation, created_at FROM orders WHERE id = ?", orderID).Scan(&orderNumber, &customerDeliveryFee, &deliveryFeeCalcJSON, &orderCreatedAt)
	if err != nil {
		orderNumber = ""
	}

	// 优先从 delivery_fee_calculation 中获取配送员实际所得（rider_payable_fee）
	// 如果不存在，则使用客户支付的配送费（delivery_fee）
	riderPayableFee := 0.0
	if deliveryFeeCalcJSON.Valid && deliveryFeeCalcJSON.String != "" {
		var calcResult model.DeliveryFeeCalculationResult
		if err := json.Unmarshal([]byte(deliveryFeeCalcJSON.String), &calcResult); err == nil {
			riderPayableFee = calcResult.RiderPayableFee
		} else {
			// 解析失败，使用客户支付的配送费
			if customerDeliveryFee.Valid {
				riderPayableFee = customerDeliveryFee.Float64
			}
		}
	} else {
		// 没有delivery_fee_calculation，使用客户支付的配送费
		if customerDeliveryFee.Valid {
			riderPayableFee = customerDeliveryFee.Float64
		}
	}

	deliveryFee := riderPayableFee // 使用配送员实际所得

	// 获取配送员信息
	employee, _ := model.GetEmployeeByEmployeeCode(record.DeliveryEmployeeCode)
	deliveryEmployeeName := ""
	if employee != nil {
		deliveryEmployeeName = employee.Name
	}

	// 获取配送流程日志
	deliveryLogs, _ := model.GetDeliveryLogsByOrderID(orderID)

	// 构建流程时间点
	processTimeline := map[string]interface{}{}
	for _, log := range deliveryLogs {
		switch log.Action {
		case model.DeliveryLogActionCreated:
			processTimeline["created_at"] = log.ActionTime
		case model.DeliveryLogActionAccepted:
			processTimeline["accepted_at"] = log.ActionTime
		case model.DeliveryLogActionPickupCompleted:
			processTimeline["pickup_completed_at"] = log.ActionTime
		case model.DeliveryLogActionDeliveringStarted:
			processTimeline["delivering_started_at"] = log.ActionTime
		case model.DeliveryLogActionDeliveringCompleted:
			processTimeline["delivering_completed_at"] = log.ActionTime
		}
	}

	data := map[string]interface{}{
		"id":                     record.ID,
		"order_id":               record.OrderID,
		"order_number":           orderNumber,
		"delivery_employee_code": record.DeliveryEmployeeCode,
		"delivery_employee_name": deliveryEmployeeName,
		"delivery_fee":           deliveryFee,
		"order_created_at":       orderCreatedAt,
		"completed_at":           record.CompletedAt,
		"created_at":             record.CreatedAt,
		"updated_at":             record.UpdatedAt,
		"process_timeline":       processTimeline,
		"delivery_logs":          deliveryLogs,
	}

	if record.ProductImageURL != nil {
		data["product_image_url"] = *record.ProductImageURL
	}
	if record.DoorplateImageURL != nil {
		data["doorplate_image_url"] = *record.DoorplateImageURL
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    data,
		"message": "获取成功",
	})
}
