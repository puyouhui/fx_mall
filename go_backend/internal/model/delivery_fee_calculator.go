package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"go_backend/internal/database"
	"go_backend/internal/utils"
)

// DeliveryFeeCalculationResult 配送费计算结果
type DeliveryFeeCalculationResult struct {
	BaseFee                  float64 `json:"base_fee"`                    // 基础配送费
	IsolatedFee              float64 `json:"isolated_fee"`                // 孤立订单补贴
	ItemFee                  float64 `json:"item_fee"`                    // 件数补贴
	UrgentFee                float64 `json:"urgent_fee"`                  // 加急订单补贴
	WeatherFee               float64 `json:"weather_fee"`                 // 极端天气补贴
	DeliveryFeeWithoutProfit float64 `json:"delivery_fee_without_profit"` // 不含利润分成的配送费
	ProfitShare              float64 `json:"profit_share"`                // 利润分成（仅管理员可见）
	RiderPayableFee          float64 `json:"rider_payable_fee"`           // 配送员实际所得
	TotalPlatformCost        float64 `json:"total_platform_cost"`         // 平台总成本
}

// DeliveryFeeCalculator 配送费计算器
type DeliveryFeeCalculator struct {
	orderID              int
	order                 *Order
	deliveryEmployeeCode *string // 可选：用于判断孤立时考虑该配送员的批次订单
}

// NewDeliveryFeeCalculator 创建配送费计算器
func NewDeliveryFeeCalculator(orderID int) (*DeliveryFeeCalculator, error) {
	order, err := GetOrderByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("订单不存在: %d", orderID)
	}

	return &DeliveryFeeCalculator{
		orderID:              orderID,
		order:                 order,
		deliveryEmployeeCode:  order.DeliveryEmployeeCode,
	}, nil
}

// NewDeliveryFeeCalculatorForEmployee 创建配送费计算器（基于指定配送员的批次判断）
func NewDeliveryFeeCalculatorForEmployee(orderID int, deliveryEmployeeCode string) (*DeliveryFeeCalculator, error) {
	order, err := GetOrderByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("订单不存在: %d", orderID)
	}

	return &DeliveryFeeCalculator{
		orderID:              orderID,
		order:                 order,
		deliveryEmployeeCode:  &deliveryEmployeeCode,
	}, nil
}

// Calculate 计算配送费
func (c *DeliveryFeeCalculator) Calculate(isAdminView bool) (*DeliveryFeeCalculationResult, error) {
	// 1. 基础配送费
	baseFee := c.getConfigFloat("delivery_base_fee", 4.0)
	baseFee = math.Max(0, baseFee)

	// 2. 孤立订单补贴
	isolatedFee := c.calculateIsolatedFee()

	// 3. 商品件数补贴
	itemFee := c.calculateItemFee()

	// 4. 加急订单补贴
	urgentFee := c.calculateUrgentFee()

	// 5. 极端天气补贴
	weatherFee := c.calculateWeatherFee()

	// 6. 汇总配送费（不含利润分成）
	deliveryFeeWithoutProfit := baseFee + isolatedFee + itemFee + urgentFee + weatherFee

	// 7. 利润分成（总是计算，但仅管理员可见明细）
	profitShare := c.calculateProfitShare(deliveryFeeWithoutProfit)

	// 配送员实际所得 = 基础配送费 + 各种补贴 + 利润分成
	riderPayableFee := deliveryFeeWithoutProfit + profitShare

	// 平台总成本 = 配送员实际所得（利润分成也是平台支付给配送员的）
	totalPlatformCost := riderPayableFee

	// 返回计算结果
	// 如果是配送员视图，隐藏利润分成明细（设为0），但实际金额已包含在rider_payable_fee中
	result := &DeliveryFeeCalculationResult{
		BaseFee:                  baseFee,
		IsolatedFee:              isolatedFee,
		ItemFee:                  itemFee,
		UrgentFee:                urgentFee,
		WeatherFee:               weatherFee,
		DeliveryFeeWithoutProfit: deliveryFeeWithoutProfit,
		ProfitShare:              profitShare,
		RiderPayableFee:          riderPayableFee,
		TotalPlatformCost:        totalPlatformCost,
	}

	// 配送员视图不显示利润分成明细
	if !isAdminView {
		result.ProfitShare = 0.0
	}

	return result, nil
}

// calculateIsolatedFee 计算孤立订单补贴
func (c *DeliveryFeeCalculator) calculateIsolatedFee() float64 {
	isolatedSubsidy := c.getConfigFloat("delivery_isolated_subsidy", 3.0)

	// 如果指定了配送员代码，必须基于该配送员的批次订单实时计算，忽略已存储的值
	// 因为不同配送员看到的孤立状态可能不同
	if c.deliveryEmployeeCode == nil || *c.deliveryEmployeeCode == "" {
		// 没有指定配送员，优先使用订单中已存储的 is_isolated 字段
		if c.order.IsIsolated {
			return isolatedSubsidy
		}
	}

	// 进行实时计算（基于配送员批次或全局未接单订单）
	isolatedDistance := c.getConfigFloat("delivery_isolated_distance", 8.0)

	// 获取订单地址
	address, err := GetAddressByID(c.order.AddressID)
	if err != nil || address == nil {
		return 0
	}

	// 检查地址是否有经纬度
	if address.Latitude == nil || address.Longitude == nil {
		return 0
	}

	// 获取邻近订单
	nearbyOrders, err := c.getNearbyOrders(*address.Latitude, *address.Longitude, isolatedDistance)
	if err != nil {
		return 0
	}

	// 过滤有效订单
	validNearby := c.filterNearbyOrders(nearbyOrders)

	// 判断是否孤立
	if len(validNearby) == 0 {
		return isolatedSubsidy
	}

	return 0
}

// getNearbyOrders 获取邻近订单
// 如果 deliveryEmployeeCode 不为空，优先基于该配送员的批次订单判断
// 否则，查询未接单的订单（pending, pending_delivery）
func (c *DeliveryFeeCalculator) getNearbyOrders(lat, lng, distance float64) ([]*Order, error) {
	var orders []*Order
	
	// 如果指定了配送员，优先基于该配送员的批次订单判断
	if c.deliveryEmployeeCode != nil && *c.deliveryEmployeeCode != "" {
		// 获取该配送员当前批次的所有订单（pending_pickup 和 delivering）
		batchOrderIDs, err := GetOrderIDsByEmployee(*c.deliveryEmployeeCode)
		if err == nil && len(batchOrderIDs) > 0 {
			// 查询这些订单的地址信息
			placeholders := make([]string, len(batchOrderIDs))
			args := make([]interface{}, len(batchOrderIDs))
			for i, orderID := range batchOrderIDs {
				placeholders[i] = "?"
				args[i] = orderID
			}
			
			query := fmt.Sprintf(`
				SELECT o.id, o.order_number, o.user_id, o.address_id, o.status, o.created_at, o.delivery_employee_code
				FROM orders o
				JOIN mini_app_addresses a ON o.address_id = a.id
				WHERE o.id IN (%s)
				  AND o.id != ?
				  AND o.status IN ('pending_pickup', 'delivering')
				  AND a.latitude IS NOT NULL
				  AND a.longitude IS NOT NULL
			`, strings.Join(placeholders, ","))
			
			args = append(args, c.orderID)
			rows, err := database.DB.Query(query, args...)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var order Order
					var addressID int
					var deliveryEmployeeCode sql.NullString
					err := rows.Scan(&order.ID, &order.OrderNumber, &order.UserID, &addressID, &order.Status, &order.CreatedAt, &deliveryEmployeeCode)
					if err != nil {
						continue
					}
					
					if deliveryEmployeeCode.Valid {
						code := deliveryEmployeeCode.String
						order.DeliveryEmployeeCode = &code
					}

					// 获取地址经纬度
					var addrLat, addrLng sql.NullFloat64
					err = database.DB.QueryRow("SELECT latitude, longitude FROM mini_app_addresses WHERE id = ?", addressID).
						Scan(&addrLat, &addrLng)
					if err != nil || !addrLat.Valid || !addrLng.Valid {
						continue
					}

					// 计算距离
					dist := utils.CalculateDistance(lat, lng, addrLat.Float64, addrLng.Float64)
					if dist <= distance {
						orders = append(orders, &order)
					}
				}
			}
		}
		
		// 如果批次订单中没有找到邻近订单，继续查询未接单的订单
		if len(orders) == 0 {
			// 查询未接单的订单
			query := `
				SELECT o.id, o.order_number, o.user_id, o.address_id, o.status, o.created_at, o.delivery_employee_code
				FROM orders o
				JOIN mini_app_addresses a ON o.address_id = a.id
				WHERE o.status IN ('pending', 'pending_delivery')
				  AND o.id != ?
				  AND a.latitude IS NOT NULL
				  AND a.longitude IS NOT NULL
			`
			rows, err := database.DB.Query(query, c.orderID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var order Order
					var addressID int
					var deliveryEmployeeCode sql.NullString
					err := rows.Scan(&order.ID, &order.OrderNumber, &order.UserID, &addressID, &order.Status, &order.CreatedAt, &deliveryEmployeeCode)
					if err != nil {
						continue
					}
					
					if deliveryEmployeeCode.Valid {
						code := deliveryEmployeeCode.String
						order.DeliveryEmployeeCode = &code
					}

					// 获取地址经纬度
					var addrLat, addrLng sql.NullFloat64
					err = database.DB.QueryRow("SELECT latitude, longitude FROM mini_app_addresses WHERE id = ?", addressID).
						Scan(&addrLat, &addrLng)
					if err != nil || !addrLat.Valid || !addrLng.Valid {
						continue
					}

					// 计算距离
					dist := utils.CalculateDistance(lat, lng, addrLat.Float64, addrLng.Float64)
					if dist <= distance {
						orders = append(orders, &order)
					}
				}
			}
		}
	} else {
		// 没有指定配送员，只查询未接单的订单
		query := `
			SELECT o.id, o.order_number, o.user_id, o.address_id, o.status, o.created_at, o.delivery_employee_code
			FROM orders o
			JOIN mini_app_addresses a ON o.address_id = a.id
			WHERE o.status IN ('pending', 'pending_delivery')
			  AND o.id != ?
			  AND a.latitude IS NOT NULL
			  AND a.longitude IS NOT NULL
		`
		rows, err := database.DB.Query(query, c.orderID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var order Order
			var addressID int
			var deliveryEmployeeCode sql.NullString
			err := rows.Scan(&order.ID, &order.OrderNumber, &order.UserID, &addressID, &order.Status, &order.CreatedAt, &deliveryEmployeeCode)
			if err != nil {
				continue
			}
			
			if deliveryEmployeeCode.Valid {
				code := deliveryEmployeeCode.String
				order.DeliveryEmployeeCode = &code
			}

			// 获取地址经纬度
			var addrLat, addrLng sql.NullFloat64
			err = database.DB.QueryRow("SELECT latitude, longitude FROM mini_app_addresses WHERE id = ?", addressID).
				Scan(&addrLat, &addrLng)
			if err != nil || !addrLat.Valid || !addrLng.Valid {
				continue
			}

			// 计算距离
			dist := utils.CalculateDistance(lat, lng, addrLat.Float64, addrLng.Float64)
			if dist <= distance {
				orders = append(orders, &order)
			}
		}
	}

	return orders, nil
}

// filterNearbyOrders 过滤邻近订单
func (c *DeliveryFeeCalculator) filterNearbyOrders(orders []*Order) []*Order {
	var validOrders []*Order
	
	// 获取配送员代码：优先使用传入的 deliveryEmployeeCode，如果没有则使用订单的
	deliveryEmployeeCode := ""
	if c.deliveryEmployeeCode != nil && *c.deliveryEmployeeCode != "" {
		deliveryEmployeeCode = *c.deliveryEmployeeCode
	} else if c.order != nil && c.order.DeliveryEmployeeCode != nil {
		deliveryEmployeeCode = *c.order.DeliveryEmployeeCode
	}
	
	for _, order := range orders {
		// 排除自己
		if order.ID == c.orderID {
			continue
		}

		// 只保留未接单的订单，或者当前配送员接的订单
		if order.Status == "pending" || order.Status == "pending_delivery" {
			validOrders = append(validOrders, order)
		} else if order.Status == "pending_pickup" || order.Status == "delivering" {
			// 如果是 pending_pickup 或 delivering 状态，只保留当前配送员接的订单
			if order.DeliveryEmployeeCode != nil && 
			   *order.DeliveryEmployeeCode == deliveryEmployeeCode && 
			   deliveryEmployeeCode != "" {
				validOrders = append(validOrders, order)
			}
		}

		// 注意：不再排除同用户短期订单
		// 因为孤立订单的判断应该基于地理位置，而不是用户
		// 即使是同一用户的订单，如果地理位置相邻，也不应该都算孤立
	}
	return validOrders
}

// calculateItemFee 计算商品件数补贴
func (c *DeliveryFeeCalculator) calculateItemFee() float64 {
	thresholdLow := c.getConfigInt("delivery_item_threshold_low", 5)
	rateLow := c.getConfigFloat("delivery_item_rate_low", 0.5)
	thresholdHigh := c.getConfigInt("delivery_item_threshold_high", 10)
	rateHigh := c.getConfigFloat("delivery_item_rate_high", 0.6)
	maxItems := c.getConfigInt("delivery_item_max_count", 50)

	// 获取订单商品数量
	items, err := GetOrderItemsByOrderID(c.orderID)
	if err != nil {
		return 0
	}

	// 使用 float64 累加，支持小数计件数
	var itemCount float64 = 0
	for _, item := range items {
		// 获取该订单项的配送计件数
		deliveryCount := c.getDeliveryCountForItem(&item)
		// 累加：配送计件数 × 数量
		itemCount += deliveryCount * float64(item.Quantity)
	}

	// 上限保护（转换为整数比较）
	itemCountInt := int(math.Ceil(itemCount)) // 向上取整
	if itemCountInt > maxItems {
		itemCount = float64(maxItems)
	} else {
		itemCount = float64(itemCountInt) // 保持整数，但保留小数累加的逻辑
	}

	// 注意：阈值比较时使用整数
	if itemCountInt < thresholdLow {
		return 0
	} else if itemCountInt < thresholdHigh {
		return itemCount * rateLow
	} else {
		return itemCount * rateHigh
	}
}

// getDeliveryCountForItem 获取订单项的配送计件数
func (c *DeliveryFeeCalculator) getDeliveryCountForItem(item *OrderItem) float64 {
	// 从商品规格中获取 delivery_count
	product, err := GetProductByID(item.ProductID)
	if err != nil || product == nil {
		// 如果获取失败，默认按1件计算
		return 1.0
	}

	// 查找匹配的规格
	for _, spec := range product.Specs {
		if spec.Name == item.SpecName {
			// 如果规格有 delivery_count 字段且大于0，使用它；否则默认为1.0
			if spec.DeliveryCount > 0 {
				return spec.DeliveryCount
			}
			return 1.0
		}
	}

	// 如果找不到匹配的规格，默认按1件计算
	return 1.0
}

// calculateUrgentFee 计算加急订单补贴
func (c *DeliveryFeeCalculator) calculateUrgentFee() float64 {
	if !c.order.IsUrgent {
		return 0
	}

	urgentSubsidy := c.getConfigFloat("delivery_urgent_subsidy", 10.0)
	return math.Max(0, urgentSubsidy)
}

// calculateWeatherFee 计算极端天气补贴
func (c *DeliveryFeeCalculator) calculateWeatherFee() float64 {
	weatherSubsidy := c.getConfigFloat("delivery_weather_subsidy", 1.0)
	extremeTemp := c.getConfigFloat("delivery_extreme_temp", 37.0)

	// 优先使用订单中已存储的天气信息（避免重复调用外部API）
	var weather *utils.WeatherData
	if c.order.WeatherInfo != nil && *c.order.WeatherInfo != "" {
		// 解析订单中存储的天气信息JSON
		var weatherData map[string]interface{}
		if err := json.Unmarshal([]byte(*c.order.WeatherInfo), &weatherData); err == nil {
			weather = &utils.WeatherData{
				Success: true,
			}
			if temp, ok := weatherData["temperature"].(float64); ok {
				weather.Temperature = temp
			}
			if cond, ok := weatherData["condition"].(string); ok {
				weather.Condition = cond
			}
			if prec, ok := weatherData["precipitation"].(float64); ok {
				weather.Precipitation = prec
			}
		}
	}

	// 如果订单中没有天气信息，才调用外部API（这种情况应该很少，因为订单创建时会更新天气信息）
	if weather == nil || !weather.Success {
		// 获取订单地址
		address, addrErr := GetAddressByID(c.order.AddressID)
		if addrErr != nil || address == nil {
			return 0
		}

		// 检查地址是否有经纬度
		if address.Latitude == nil || address.Longitude == nil {
			return 0
		}

		// 获取天气信息（调用外部API）
		amapKey, _ := GetSystemSetting("map_amap_key")
		var weatherErr error
		weather, weatherErr = utils.GetWeatherByLocation(*address.Latitude, *address.Longitude, amapKey)
		if weatherErr != nil || !weather.Success {
			// 数据缺失按正常天气处理
			return 0
		}
	}

	// 判断是否为极端天气
	if utils.IsExtremeWeather(weather, extremeTemp) {
		return weatherSubsidy
	}

	return 0
}

// calculateProfitShare 计算利润分成
func (c *DeliveryFeeCalculator) calculateProfitShare(deliveryFeeWithoutProfit float64) float64 {
	profitThreshold := c.getConfigFloat("delivery_profit_threshold", 25.0)
	profitShareRate := c.getConfigFloat("delivery_profit_share_rate", 0.08)
	maxProfitShare := c.getConfigFloat("delivery_max_profit_share", 50.0)

	// 获取订单利润
	orderProfit := c.calculateOrderProfit()
	if orderProfit <= profitThreshold {
		return 0.0
	}

	// 计算分成基数（避免循环依赖）
	profitExcess := orderProfit - deliveryFeeWithoutProfit
	if profitExcess <= 0 {
		return 0.0
	}

	// 计算分成
	profitShare := profitExcess * profitShareRate

	// 边界保护
	profitShare = math.Min(math.Max(0, profitShare), maxProfitShare)
	return math.Round(profitShare*100) / 100
}

// CalculateOrderProfit 计算订单利润（公开方法）
func (c *DeliveryFeeCalculator) CalculateOrderProfit() float64 {
	return c.calculateOrderProfit()
}

// calculateOrderProfit 计算订单利润
func (c *DeliveryFeeCalculator) calculateOrderProfit() float64 {
	// 获取订单商品明细
	items, err := GetOrderItemsByOrderID(c.orderID)
	if err != nil {
		return 0
	}

	// 计算总成本
	totalCost := 0.0
	for _, item := range items {
		// 从商品规格JSON中获取成本
		var specsJSON sql.NullString
		err := database.DB.QueryRow(`SELECT specs FROM products WHERE id = ?`, item.ProductID).Scan(&specsJSON)
		if err != nil || !specsJSON.Valid {
			continue
		}

		// 解析规格JSON
		var specs []struct {
			Name           string  `json:"name"`
			WholesalePrice float64 `json:"wholesale_price"`
			RetailPrice    float64 `json:"retail_price"`
			Cost           float64 `json:"cost"`
		}
		if err := json.Unmarshal([]byte(specsJSON.String), &specs); err != nil {
			continue
		}

		// 查找匹配的规格
		var cost float64
		for _, spec := range specs {
			if spec.Name == item.SpecName {
				cost = spec.Cost
				break
			}
		}
		if cost < 0 {
			cost = 0
		}
		totalCost += cost * float64(item.Quantity)
	}

	// 利润 = 商品总金额 - 总成本
	profit := c.order.GoodsAmount - totalCost
	return math.Max(0, profit)
}

// getConfigFloat 获取配置值（浮点数）
func (c *DeliveryFeeCalculator) getConfigFloat(key string, defaultValue float64) float64 {
	valueStr, err := GetSystemSetting(key)
	if err != nil || valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

// getConfigInt 获取配置值（整数）
func (c *DeliveryFeeCalculator) getConfigInt(key string, defaultValue int) int {
	valueStr, err := GetSystemSetting(key)
	if err != nil || valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// UpdateOrderDeliveryInfo 更新订单的配送相关信息（孤立状态、天气信息等）
// 同时会更新可能受影响的邻近订单的孤立状态
func UpdateOrderDeliveryInfo(orderID int) error {
	calculator, err := NewDeliveryFeeCalculator(orderID)
	if err != nil {
		return err
	}

	// 获取订单地址
	order, err := GetOrderByID(orderID)
	if err != nil || order == nil {
		return fmt.Errorf("订单不存在: %d", orderID)
	}

	address, err := GetAddressByID(order.AddressID)
	if err != nil || address == nil {
		return fmt.Errorf("地址不存在")
	}

	isolatedDistance := calculator.getConfigFloat("delivery_isolated_distance", 8.0)

	// 更新当前订单的孤立状态
	isIsolated := false
	affectedOrderIDsMap := make(map[int]bool) // 使用map去重，避免重复更新同一订单
	if address.Latitude != nil && address.Longitude != nil {
		nearbyOrders, err := calculator.getNearbyOrders(*address.Latitude, *address.Longitude, isolatedDistance)
		if err == nil {
			validNearby := calculator.filterNearbyOrders(nearbyOrders)
			isIsolated = len(validNearby) == 0

			// 如果新订单找到了相邻订单，这些相邻订单需要重新计算孤立状态
			// 因为新订单的出现可能使它们不再孤立
			for _, nearbyOrder := range validNearby {
				affectedOrderIDsMap[nearbyOrder.ID] = true
			}

			// 还需要找到所有以当前订单为相邻订单的订单（反向查找）
			// 这些订单也可能因为新订单的出现而改变孤立状态
			// 查询所有在范围内的未接单订单，检查它们是否以当前订单为相邻订单
			reverseQuery := `
				SELECT o.id, a.latitude, a.longitude
				FROM orders o
				JOIN mini_app_addresses a ON o.address_id = a.id
				WHERE o.status IN ('pending', 'pending_delivery')
				  AND o.id != ?
				  AND a.latitude IS NOT NULL
				  AND a.longitude IS NOT NULL
			`
			reverseRows, err := database.DB.Query(reverseQuery, orderID)
			if err == nil {
				defer reverseRows.Close()
				for reverseRows.Next() {
					var otherOrderID int
					var otherLat, otherLng sql.NullFloat64
					if err := reverseRows.Scan(&otherOrderID, &otherLat, &otherLng); err == nil && otherLat.Valid && otherLng.Valid {
						// 计算距离
						dist := utils.CalculateDistance(*address.Latitude, *address.Longitude, otherLat.Float64, otherLng.Float64)
						if dist <= isolatedDistance {
							// 这个订单以当前订单为相邻订单，需要重新计算孤立状态
							affectedOrderIDsMap[otherOrderID] = true
						}
					}
				}
			}
		}
	}

	// 更新天气信息
	var weatherInfoJSON *string
	if address.Latitude != nil && address.Longitude != nil {
		amapKey, _ := GetSystemSetting("map_amap_key")
		weather, err := utils.GetWeatherByLocation(*address.Latitude, *address.Longitude, amapKey)
		if err == nil && weather.Success {
			weatherData := map[string]interface{}{
				"temperature":   weather.Temperature,
				"condition":     weather.Condition,
				"precipitation": weather.Precipitation,
			}
			jsonBytes, err := json.Marshal(weatherData)
			if err == nil {
				jsonStr := string(jsonBytes)
				weatherInfoJSON = &jsonStr
			}
		}
	}

	// 更新当前订单
	_, err = database.DB.Exec(`
		UPDATE orders 
		SET is_isolated = ?, weather_info = ?, updated_at = NOW()
		WHERE id = ?
	`, isIsolated, weatherInfoJSON, orderID)
	if err != nil {
		return err
	}

	// 更新可能受影响的邻近订单的孤立状态
	// 这些订单可能因为新订单的创建而不再孤立
	// 将map转换为slice
	var affectedOrderIDs []int
	for affectedID := range affectedOrderIDsMap {
		affectedOrderIDs = append(affectedOrderIDs, affectedID)
	}
	// 同步更新受影响订单的孤立状态和配送费，确保状态一致性
	for _, affectedOrderID := range affectedOrderIDs {
		// 重新计算受影响订单的孤立状态
		// 注意：这里需要重新查询，因为可能已经有新订单了
		if err := updateOrderIsolatedStatus(affectedOrderID, isolatedDistance); err != nil {
			// 记录错误但不中断流程
			log.Printf("[UpdateOrderDeliveryInfo] 更新订单 %d 的孤立状态失败: %v", affectedOrderID, err)
		} else {
			// 孤立状态更新成功后，重新计算并存储配送费
			// 因为孤立状态改变会影响配送费金额（孤立补贴）
			if err := CalculateAndStoreOrderProfit(affectedOrderID); err != nil {
				// 记录错误但不中断流程
				log.Printf("[UpdateOrderDeliveryInfo] 重新计算订单 %d 的配送费失败: %v", affectedOrderID, err)
			}
		}
	}

	return nil
}

// updateOrderIsolatedStatus 更新单个订单的孤立状态
func updateOrderIsolatedStatus(orderID int, isolatedDistance float64) error {
	order, err := GetOrderByID(orderID)
	if err != nil || order == nil {
		return err
	}

	address, err := GetAddressByID(order.AddressID)
	if err != nil || address == nil {
		return err
	}

	// 如果地址没有经纬度，无法判断孤立状态
	if address.Latitude == nil || address.Longitude == nil {
		return nil
	}

	// 创建临时计算器用于查询邻近订单
	calculator, err := NewDeliveryFeeCalculator(orderID)
	if err != nil {
		return err
	}

	// 获取邻近订单
	nearbyOrders, err := calculator.getNearbyOrders(*address.Latitude, *address.Longitude, isolatedDistance)
	if err != nil {
		return err
	}

	// 过滤有效订单
	validNearby := calculator.filterNearbyOrders(nearbyOrders)
	isIsolated := len(validNearby) == 0

	// 更新孤立状态
	_, err = database.DB.Exec(`
		UPDATE orders 
		SET is_isolated = ?, updated_at = NOW()
		WHERE id = ?
	`, isIsolated, orderID)

	return err
}

// updateAffectedOrdersIsolatedStatus 更新受影响的订单的孤立状态
// 当订单状态变化可能影响其他订单的孤立判断时调用
func updateAffectedOrdersIsolatedStatus(orderID int) error {
	order, err := GetOrderByID(orderID)
	if err != nil || order == nil {
		return err
	}

	address, err := GetAddressByID(order.AddressID)
	if err != nil || address == nil {
		return nil
	}

	if address.Latitude == nil || address.Longitude == nil {
		return nil
	}

	isolatedDistanceStr, _ := GetSystemSetting("delivery_isolated_distance")
	isolatedDistance := 8.0
	if isolatedDistanceStr != "" {
		if val, err := strconv.ParseFloat(isolatedDistanceStr, 64); err == nil {
			isolatedDistance = val
		}
	}

	// 查找所有以当前订单为邻近订单的订单（反向查找）
	// 只查询未接单的订单（pending, pending_delivery）
	query := `
		SELECT o.id, a.latitude, a.longitude
		FROM orders o
		JOIN mini_app_addresses a ON o.address_id = a.id
		WHERE o.status IN ('pending', 'pending_delivery')
		  AND o.id != ?
		  AND a.latitude IS NOT NULL
		  AND a.longitude IS NOT NULL
	`
	rows, err := database.DB.Query(query, orderID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var affectedOrderIDs []int
	for rows.Next() {
		var otherOrderID int
		var otherLat, otherLng sql.NullFloat64
		if err := rows.Scan(&otherOrderID, &otherLat, &otherLng); err == nil && otherLat.Valid && otherLng.Valid {
			dist := utils.CalculateDistance(*address.Latitude, *address.Longitude, otherLat.Float64, otherLng.Float64)
			if dist <= isolatedDistance {
				affectedOrderIDs = append(affectedOrderIDs, otherOrderID)
			}
		}
	}

	// 更新所有受影响订单的孤立状态和配送费
	for _, affectedOrderID := range affectedOrderIDs {
		if err := updateOrderIsolatedStatus(affectedOrderID, isolatedDistance); err != nil {
			log.Printf("[updateAffectedOrdersIsolatedStatus] 更新订单 %d 的孤立状态失败: %v", affectedOrderID, err)
		} else {
			// 重新计算配送费
			if err := CalculateAndStoreOrderProfit(affectedOrderID); err != nil {
				log.Printf("[updateAffectedOrdersIsolatedStatus] 重新计算订单 %d 的配送费失败: %v", affectedOrderID, err)
			}
		}
	}

	return nil
}

// CalculateAndStoreOrderProfit 计算并存储订单的配送费计算结果和利润信息
func CalculateAndStoreOrderProfit(orderID int) error {
	order, err := GetOrderByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("订单不存在: %d", orderID)
	}

	// 如果订单已经被接单，使用已存储的配送费计算利润，而不是重新计算配送费
	if order.DeliveryEmployeeCode != nil && *order.DeliveryEmployeeCode != "" {
		// 从数据库读取已存储的配送费计算结果
		var deliveryFeeCalcJSON sql.NullString
		err := database.DB.QueryRow(`
			SELECT delivery_fee_calculation
			FROM orders WHERE id = ?
		`, orderID).Scan(&deliveryFeeCalcJSON)

		if err == nil && deliveryFeeCalcJSON.Valid && deliveryFeeCalcJSON.String != "" {
			// 解析已存储的配送费计算结果
			var deliveryFeeResult DeliveryFeeCalculationResult
			if json.Unmarshal([]byte(deliveryFeeCalcJSON.String), &deliveryFeeResult) == nil {
				// 创建计算器用于计算订单利润（不重新计算配送费）
				calculator, err := NewDeliveryFeeCalculator(orderID)
				if err != nil {
					return err
				}

				// 计算订单利润
				orderProfit := calculator.CalculateOrderProfit()

				// 计算净利润（使用已存储的配送费）
				netProfit := orderProfit - deliveryFeeResult.TotalPlatformCost

				// 更新订单表（只更新利润，不更新配送费）
				_, err = database.DB.Exec(`
					UPDATE orders 
					SET order_profit = ?, net_profit = ?, updated_at = NOW()
					WHERE id = ?
				`, orderProfit, netProfit, orderID)

				return err
			}
		}
	}

	// 如果订单未接单，或者没有已存储的配送费，则重新计算
	calculator, err := NewDeliveryFeeCalculator(orderID)
	if err != nil {
		return err
	}

	return CalculateAndStoreOrderProfitWithCalculator(calculator, orderID)
}

// CalculateAndStoreOrderProfitWithCalculator 使用指定的计算器计算并存储订单的配送费
func CalculateAndStoreOrderProfitWithCalculator(calculator *DeliveryFeeCalculator, orderID int) error {
	// 计算配送费（管理员视图，包含利润分成）
	deliveryFeeResult, err := calculator.Calculate(true)
	if err != nil {
		return err
	}

	// 计算订单利润
	orderProfit := calculator.CalculateOrderProfit()

	// 计算净利润
	netProfit := orderProfit - deliveryFeeResult.TotalPlatformCost

	// 将配送费计算结果序列化为JSON
	deliveryFeeJSON, err := json.Marshal(deliveryFeeResult)
	if err != nil {
		return err
	}

	// 更新订单表
	_, err = database.DB.Exec(`
		UPDATE orders 
		SET delivery_fee_calculation = ?, order_profit = ?, net_profit = ?, updated_at = NOW()
		WHERE id = ?
	`, string(deliveryFeeJSON), orderProfit, netProfit, orderID)

	return err
}

// CalculateRiderDeliveryFeePreview 计算配送员配送费预览（基于采购单和地址，不创建订单）
func CalculateRiderDeliveryFeePreview(items []PurchaseListItem, addressID int, isUrgent bool, userType string) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"base_fee":          0.0,
		"isolated_fee":      0.0,
		"item_fee":          0.0,
		"urgent_fee":        0.0,
		"weather_fee":       0.0,
		"profit_share":      0.0,
		"rider_payable_fee": 0.0,
	}

	// 1. 基础配送费
	baseFeeStr, _ := GetSystemSetting("delivery_base_fee")
	baseFee := 4.0
	if baseFeeStr != "" {
		if val, err := strconv.ParseFloat(baseFeeStr, 64); err == nil {
			baseFee = math.Max(0, val)
		}
	}
	result["base_fee"] = baseFee

	// 2. 商品件数补贴
	thresholdLowStr, _ := GetSystemSetting("delivery_item_threshold_low")
	thresholdLow := 5
	if thresholdLowStr != "" {
		if val, err := strconv.Atoi(thresholdLowStr); err == nil {
			thresholdLow = val
		}
	}

	rateLowStr, _ := GetSystemSetting("delivery_item_rate_low")
	rateLow := 0.5
	if rateLowStr != "" {
		if val, err := strconv.ParseFloat(rateLowStr, 64); err == nil {
			rateLow = val
		}
	}

	thresholdHighStr, _ := GetSystemSetting("delivery_item_threshold_high")
	thresholdHigh := 10
	if thresholdHighStr != "" {
		if val, err := strconv.Atoi(thresholdHighStr); err == nil {
			thresholdHigh = val
		}
	}

	rateHighStr, _ := GetSystemSetting("delivery_item_rate_high")
	rateHigh := 0.6
	if rateHighStr != "" {
		if val, err := strconv.ParseFloat(rateHighStr, 64); err == nil {
			rateHigh = val
		}
	}

	maxItemsStr, _ := GetSystemSetting("delivery_item_max_count")
	maxItems := 50
	if maxItemsStr != "" {
		if val, err := strconv.Atoi(maxItemsStr); err == nil {
			maxItems = val
		}
	}

	// 辅助函数：从规格快照获取配送计件数
	getDeliveryCountFromSpecSnapshot := func(snapshot PurchaseSpecSnapshot) float64 {
		// 如果规格快照有 delivery_count 字段且大于0，使用它；否则默认为1.0
		if snapshot.DeliveryCount > 0 {
			return snapshot.DeliveryCount
		}
		return 1.0
	}

	// 使用 float64 累加，支持小数计件数
	var itemCount float64 = 0
	for _, item := range items {
		// 从规格快照中获取配送计件数
		deliveryCount := getDeliveryCountFromSpecSnapshot(item.SpecSnapshot)
		// 累加：配送计件数 × 数量
		itemCount += deliveryCount * float64(item.Quantity)
	}

	// 上限保护（转换为整数比较）
	itemCountInt := int(math.Ceil(itemCount)) // 向上取整
	if itemCountInt > maxItems {
		itemCount = float64(maxItems)
	} else {
		itemCount = float64(itemCountInt) // 保持整数，但保留小数累加的逻辑
	}

	itemFee := 0.0
	if itemCountInt >= thresholdLow {
		if itemCountInt < thresholdHigh {
			itemFee = itemCount * rateLow
		} else {
			itemFee = itemCount * rateHigh
		}
	}
	result["item_fee"] = itemFee

	// 3. 加急订单补贴
	urgentFee := 0.0
	if isUrgent {
		urgentSubsidyStr, _ := GetSystemSetting("delivery_urgent_subsidy")
		urgentSubsidy := 10.0
		if urgentSubsidyStr != "" {
			if val, err := strconv.ParseFloat(urgentSubsidyStr, 64); err == nil {
				urgentSubsidy = math.Max(0, val)
			}
		}
		urgentFee = urgentSubsidy
	}
	result["urgent_fee"] = urgentFee

	// 4. 孤立订单补贴（需要地址信息）
	isolatedFee := 0.0
	if addressID > 0 {
		isolatedSubsidyStr, _ := GetSystemSetting("delivery_isolated_subsidy")
		isolatedSubsidy := 3.0
		if isolatedSubsidyStr != "" {
			if val, err := strconv.ParseFloat(isolatedSubsidyStr, 64); err == nil {
				isolatedSubsidy = val
			}
		}

		isolatedDistanceStr, _ := GetSystemSetting("delivery_isolated_distance")
		isolatedDistance := 8.0
		if isolatedDistanceStr != "" {
			if val, err := strconv.ParseFloat(isolatedDistanceStr, 64); err == nil {
				isolatedDistance = val
			}
		}

		// 获取地址
		address, err := GetAddressByID(addressID)
		if err == nil && address != nil && address.Latitude != nil && address.Longitude != nil {
			// 查询附近未接单的订单（只查询pending和pending_delivery）
			rows, err := database.DB.Query(`
				SELECT a.latitude, a.longitude
				FROM orders o
				JOIN mini_app_addresses a ON o.address_id = a.id
				WHERE o.status IN ('pending', 'pending_delivery')
				  AND a.latitude IS NOT NULL
				  AND a.longitude IS NOT NULL
			`)
			if err == nil {
				defer rows.Close()
				hasNearby := false
				for rows.Next() {
					var lat, lng sql.NullFloat64
					if err := rows.Scan(&lat, &lng); err == nil && lat.Valid && lng.Valid {
						dist := utils.CalculateDistance(*address.Latitude, *address.Longitude, lat.Float64, lng.Float64)
						if dist <= isolatedDistance {
							hasNearby = true
							break
						}
					}
				}
				if !hasNearby {
					isolatedFee = isolatedSubsidy
				}
			}
		}
	}
	result["isolated_fee"] = isolatedFee

	// 5. 极端天气补贴（需要地址信息）
	weatherFee := 0.0
	if addressID > 0 {
		weatherSubsidyStr, _ := GetSystemSetting("delivery_weather_subsidy")
		weatherSubsidy := 1.0
		if weatherSubsidyStr != "" {
			if val, err := strconv.ParseFloat(weatherSubsidyStr, 64); err == nil {
				weatherSubsidy = val
			}
		}

		extremeTempStr, _ := GetSystemSetting("delivery_extreme_temp")
		extremeTemp := 37.0
		if extremeTempStr != "" {
			if val, err := strconv.ParseFloat(extremeTempStr, 64); err == nil {
				extremeTemp = val
			}
		}

		// 获取地址
		address, err := GetAddressByID(addressID)
		if err == nil && address != nil && address.Latitude != nil && address.Longitude != nil {
			amapKey, _ := GetSystemSetting("map_amap_key")
			weather, err := utils.GetWeatherByLocation(*address.Latitude, *address.Longitude, amapKey)
			if err == nil && weather.Success && utils.IsExtremeWeather(weather, extremeTemp) {
				weatherFee = weatherSubsidy
			}
		}
	}
	result["weather_fee"] = weatherFee

	// 6. 汇总配送费（不含利润分成）
	deliveryFeeWithoutProfit := baseFee + isolatedFee + itemFee + urgentFee + weatherFee

	// 7. 利润分成（预览时也计算，确保与实际订单一致）
	profitShare := 0.0
	profitThresholdStr, _ := GetSystemSetting("delivery_profit_threshold")
	profitThreshold := 25.0
	if profitThresholdStr != "" {
		if val, err := strconv.ParseFloat(profitThresholdStr, 64); err == nil {
			profitThreshold = val
		}
	}

	profitShareRateStr, _ := GetSystemSetting("delivery_profit_share_rate")
	profitShareRate := 0.08
	if profitShareRateStr != "" {
		if val, err := strconv.ParseFloat(profitShareRateStr, 64); err == nil {
			profitShareRate = val
		}
	}

	maxProfitShareStr, _ := GetSystemSetting("delivery_max_profit_share")
	maxProfitShare := 50.0
	if maxProfitShareStr != "" {
		if val, err := strconv.ParseFloat(maxProfitShareStr, 64); err == nil {
			maxProfitShare = val
		}
	}

	// 计算订单利润
	// 商品总金额
	goodsAmount := 0.0
	totalCost := 0.0
	for _, item := range items {
		// 根据用户类型计算商品金额
		var price float64
		if userType == "wholesale" {
			price = item.SpecSnapshot.WholesalePrice
			if price <= 0 {
				price = item.SpecSnapshot.RetailPrice
			}
		} else {
			price = item.SpecSnapshot.RetailPrice
			if price <= 0 {
				price = item.SpecSnapshot.WholesalePrice
			}
		}
		if price <= 0 {
			price = item.SpecSnapshot.Cost
		}
		if price < 0 {
			price = 0
		}
		goodsAmount += price * float64(item.Quantity)

		// 计算成本
		cost := item.SpecSnapshot.Cost
		if cost < 0 {
			cost = 0
		}
		totalCost += cost * float64(item.Quantity)
	}

	// 订单利润 = 商品总金额 - 总成本
	orderProfit := goodsAmount - totalCost
	if orderProfit < 0 {
		orderProfit = 0
	}

	// 计算利润分成
	if orderProfit > profitThreshold {
		// 计算分成基数（避免循环依赖）
		profitExcess := orderProfit - deliveryFeeWithoutProfit
		if profitExcess > 0 {
			profitShare = profitExcess * profitShareRate
			// 边界保护
			profitShare = math.Min(math.Max(0, profitShare), maxProfitShare)
			profitShare = math.Round(profitShare*100) / 100
		}
	}
	result["profit_share"] = profitShare

	// 配送员实际所得 = 基础配送费 + 各种补贴 + 利润分成
	riderPayableFee := deliveryFeeWithoutProfit + profitShare
	result["rider_payable_fee"] = riderPayableFee

	return result, nil
}
