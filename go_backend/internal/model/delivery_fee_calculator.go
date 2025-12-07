package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

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
	orderID int
	order   *Order
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
		orderID: orderID,
		order:   order,
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

	// 优先使用订单中已存储的 is_isolated 字段（避免重复查询数据库）
	// 订单创建或更新时会计算并存储这个值
	if c.order.IsIsolated {
		return isolatedSubsidy
	}

	// 如果订单中没有存储孤立状态，才进行实时计算（这种情况应该很少）
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
func (c *DeliveryFeeCalculator) getNearbyOrders(lat, lng, distance float64) ([]*Order, error) {
	// 查询所有待配送订单（排除当前订单）
	query := `
		SELECT o.id, o.order_number, o.user_id, o.address_id, o.status, o.created_at
		FROM orders o
		JOIN mini_app_addresses a ON o.address_id = a.id
		WHERE o.status IN ('pending_delivery', 'pending')
		  AND o.id != ?
		  AND a.latitude IS NOT NULL
		  AND a.longitude IS NOT NULL
	`
	rows, err := database.DB.Query(query, c.orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		var addressID int
		err := rows.Scan(&order.ID, &order.OrderNumber, &order.UserID, &addressID, &order.Status, &order.CreatedAt)
		if err != nil {
			continue
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

	return orders, nil
}

// filterNearbyOrders 过滤邻近订单
func (c *DeliveryFeeCalculator) filterNearbyOrders(orders []*Order) []*Order {
	var validOrders []*Order
	for _, order := range orders {
		// 排除自己
		if order.ID == c.orderID {
			continue
		}

		// 排除已完成订单
		if order.Status != "pending_delivery" && order.Status != "pending" {
			continue
		}

		// 排除同用户短期订单（1小时内）
		if order.UserID == c.order.UserID {
			timeDiff := order.CreatedAt.Sub(c.order.CreatedAt)
			if timeDiff < 0 {
				timeDiff = -timeDiff
			}
			if timeDiff < time.Hour {
				continue
			}
		}

		validOrders = append(validOrders, order)
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

	itemCount := 0
	for _, item := range items {
		itemCount += item.Quantity
	}

	// 上限保护
	if itemCount > maxItems {
		itemCount = maxItems
	}

	if itemCount < thresholdLow {
		return 0
	} else if itemCount < thresholdHigh {
		return float64(itemCount) * rateLow
	} else {
		return float64(itemCount) * rateHigh
	}
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
	var affectedOrderIDs []int // 可能受影响的订单ID列表
	if address.Latitude != nil && address.Longitude != nil {
		nearbyOrders, err := calculator.getNearbyOrders(*address.Latitude, *address.Longitude, isolatedDistance)
		if err == nil {
			validNearby := calculator.filterNearbyOrders(nearbyOrders)
			isIsolated = len(validNearby) == 0

			// 收集可能受影响的订单ID（这些订单可能因为新订单的创建而改变孤立状态）
			for _, nearbyOrder := range validNearby {
				affectedOrderIDs = append(affectedOrderIDs, nearbyOrder.ID)
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
	for _, affectedOrderID := range affectedOrderIDs {
		_ = updateOrderIsolatedStatus(affectedOrderID, isolatedDistance)
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

// CalculateAndStoreOrderProfit 计算并存储订单的配送费计算结果和利润信息
func CalculateAndStoreOrderProfit(orderID int) error {
	calculator, err := NewDeliveryFeeCalculator(orderID)
	if err != nil {
		return err
	}

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
