package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// RoutePlanningResult 路线规划结果
type RoutePlanningResult struct {
	Success   bool            `json:"success"`
	Message   string          `json:"message,omitempty"`
	Route     *RouteInfo      `json:"route,omitempty"`
	Routes    []*RouteInfo    `json:"routes,omitempty"`
	Polyline  string          `json:"polyline,omitempty"`  // 路线坐标点串
	Distance  int             `json:"distance,omitempty"`  // 路线距离（米）
	Duration  int             `json:"duration,omitempty"`  // 路线耗时（秒）
	Steps     []*RouteStep    `json:"steps,omitempty"`     // 路线步骤
	Waypoints []*WaypointInfo `json:"waypoints,omitempty"` // 途经点信息
}

// RouteInfo 路线信息
type RouteInfo struct {
	Distance int          `json:"distance"` // 路线距离（米）
	Duration int          `json:"duration"` // 路线耗时（秒）
	Polyline string       `json:"polyline"` // 路线坐标点串
	Steps    []*RouteStep `json:"steps"`    // 路线步骤
}

// RouteStep 路线步骤
type RouteStep struct {
	Instruction string `json:"instruction"` // 导航指令
	Road        string `json:"road"`        // 道路名称
	Distance    int    `json:"distance"`    // 步骤距离（米）
	Duration    int    `json:"duration"`    // 步骤耗时（秒）
	Polyline    string `json:"polyline"`    // 步骤坐标点串
}

// WaypointInfo 途经点信息
type WaypointInfo struct {
	Location string `json:"location"` // 坐标（经度,纬度）
	Name     string `json:"name"`     // 名称
}

// Coordinate 坐标点
type Coordinate struct {
	Longitude float64 `json:"longitude"` // 经度
	Latitude  float64 `json:"latitude"`  // 纬度
}

// PlanRoute 规划路线（使用高德地图API）
// origin: 起点坐标
// destination: 终点坐标
// waypoints: 途经点坐标列表（可选）
// apiKey: 高德地图API Key
// 返回路线规划结果
// 注意：使用高德地图API的途经点功能，确保按顺序经过所有途经点且路线顺路
func PlanRoute(origin Coordinate, destination Coordinate, waypoints []Coordinate, apiKey string) (*RoutePlanningResult, error) {
	if apiKey == "" {
		return &RoutePlanningResult{
			Success: false,
			Message: "未配置高德地图API Key，无法进行路线规划",
		}, nil
	}

	// 高德地图路线规划2.0 API
	apiURL := "https://restapi.amap.com/v5/direction/driving"

	params := url.Values{}
	params.Set("key", apiKey)
	params.Set("origin", fmt.Sprintf("%.6f,%.6f", origin.Longitude, origin.Latitude))
	params.Set("destination", fmt.Sprintf("%.6f,%.6f", destination.Longitude, destination.Latitude))

	// 算路策略：1=费用优先（只返回一条路线）
	params.Set("strategy", "1")

	// 如果有途经点，添加途经点参数
	if len(waypoints) > 0 {
		waypointStrs := make([]string, len(waypoints))
		for i, wp := range waypoints {
			waypointStrs[i] = fmt.Sprintf("%.6f,%.6f", wp.Longitude, wp.Latitude)
		}
		// 途经点格式：经度,纬度;经度,纬度;...（使用分号分隔，不是管道符）
		params.Set("waypoints", strings.Join(waypointStrs, ";"))
	}

	// 返回详细导航指令和路线坐标
	// show_fields: cost(费用),navi(导航),polyline(路线坐标)
	params.Set("show_fields", "cost,navi,polyline")

	reqURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// 记录请求URL（隐藏API Key）
	safeURL := strings.Replace(reqURL, "key="+apiKey, "key=***", 1)
	log.Printf("[PlanRoute] 请求高德地图路线规划API - URL: %s\n", safeURL)
	log.Printf("[PlanRoute] 起点: %.6f,%.6f, 终点: %.6f,%.6f, 途经点数量: %d\n",
		origin.Latitude, origin.Longitude, destination.Latitude, destination.Longitude, len(waypoints))

	resp, err := http.Get(reqURL)
	if err != nil {
		log.Printf("[PlanRoute] HTTP请求失败: %v\n", err)
		return &RoutePlanningResult{
			Success: false,
			Message: fmt.Sprintf("请求高德地图路线规划API失败: %v", err),
		}, err
	}
	defer resp.Body.Close()

	log.Printf("[PlanRoute] HTTP响应状态码: %d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[PlanRoute] 读取响应体失败: %v\n", err)
		return &RoutePlanningResult{
			Success: false,
			Message: fmt.Sprintf("读取高德地图路线规划API响应失败: %v", err),
		}, err
	}

	// 记录完整原始响应（用于调试）
	responseStr := string(body)
	// 先打印前500字符，避免日志过长
	if len(responseStr) > 500 {
		log.Printf("[PlanRoute] API完整响应内容（前500字符）: %s...\n", responseStr[:500])
	} else {
		log.Printf("[PlanRoute] API完整响应内容: %s\n", responseStr)
	}
	// 打印完整响应内容
	log.Printf("[PlanRoute] API完整响应内容（全部，长度: %d）: %s\n", len(responseStr), responseStr)

	// 先尝试解析v5格式，如果失败再尝试v3格式
	// 注意：taxi_cost 可能是字符串或对象，使用 json.RawMessage 来灵活处理
	var result struct {
		Status string `json:"status"`
		Info   string `json:"info"`
		Count  string `json:"count"`
		Route  struct {
			Origin      string          `json:"origin"`
			Destination string          `json:"destination"`
			TaxiCost    json.RawMessage `json:"taxi_cost"` // 可能是字符串或对象
			Waypoints   []struct {
				Location string `json:"location"` // 途经点坐标（经度,纬度）
			} `json:"waypoints,omitempty"` // 途经点信息（优化后的顺序）
			Paths []struct {
				Distance string `json:"distance"`
				Duration string `json:"duration"`
				Steps    []struct {
					Instruction     string          `json:"instruction"`
					Road            string          `json:"road"`
					Distance        string          `json:"distance"`
					Duration        string          `json:"duration"`
					Polyline        string          `json:"polyline"`
					Action          string          `json:"action"`
					AssistantAction string          `json:"assistant_action"`
					Navi            json.RawMessage `json:"navi,omitempty"` // navi字段可能包含polyline
				} `json:"steps"`
				Polyline string `json:"polyline"`
			} `json:"paths"`
		} `json:"route"`
		// v3格式兼容
		Routes []struct {
			Distance string `json:"distance"`
			Duration string `json:"duration"`
			Steps    []struct {
				Instruction string `json:"instruction"`
				Road        string `json:"road"`
				Distance    string `json:"distance"`
				Duration    string `json:"duration"`
				Polyline    string `json:"polyline"`
			} `json:"steps"`
			Polyline string `json:"polyline"`
		} `json:"routes"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[PlanRoute] JSON解析失败: %v, 响应内容: %s\n", err, string(body))
		return &RoutePlanningResult{
			Success: false,
			Message: fmt.Sprintf("解析高德地图路线规划API响应失败: %v", err),
		}, err
	}

	log.Printf("[PlanRoute] API返回状态: status=%s, info=%s, count=%s\n", result.Status, result.Info, result.Count)

	// 检查API是否返回了途经点信息
	if len(result.Route.Waypoints) > 0 {
		log.Printf("[PlanRoute] API返回了 %d 个途经点（优化后的顺序）\n", len(result.Route.Waypoints))
		for i, wp := range result.Route.Waypoints {
			log.Printf("[PlanRoute] API途经点[%d]: %s\n", i+1, wp.Location)
		}
	} else {
		log.Printf("[PlanRoute] API未返回途经点信息，可能API不支持返回途经点顺序，或者途经点参数有问题\n")
	}

	if result.Status != "1" {
		log.Printf("[PlanRoute] API返回错误状态: status=%s, info=%s\n", result.Status, result.Info)
		return &RoutePlanningResult{
			Success: false,
			Message: fmt.Sprintf("高德地图路线规划API返回错误: %s", result.Info),
		}, nil
	}

	// 解析路线结果（优先使用v5格式，如果为空则使用v3格式）
	var distanceStr, durationStr, polylineStr string
	var stepList []struct {
		Instruction string `json:"instruction"`
		Road        string `json:"road"`
		Distance    string `json:"distance"`
		Duration    string `json:"duration"`
		Polyline    string `json:"polyline"`
	}

	if len(result.Route.Paths) > 0 {
		// v5格式
		log.Printf("[PlanRoute] 使用v5格式，找到 %d 条路线\n", len(result.Route.Paths))
		path := result.Route.Paths[0]
		distanceStr = path.Distance
		durationStr = path.Duration
		polylineStr = path.Polyline
		// 检查polyline是否包含途经点
		polylinePointCount := 0
		if polylineStr != "" {
			polylinePointCount = len(strings.Split(polylineStr, ";"))
		}
		log.Printf("[PlanRoute] 路线信息 - 距离: %s米, 耗时: %s秒, 步骤数: %d, path.polyline长度: %d, 坐标点数: %d\n",
			distanceStr, durationStr, len(path.Steps), len(polylineStr), polylinePointCount)
		if len(waypoints) > 0 {
			log.Printf("[PlanRoute] 途经点数量: %d, polyline应该包含经过所有途经点的完整路线\n", len(waypoints))
		}

		// 转换步骤格式
		stepList = make([]struct {
			Instruction string `json:"instruction"`
			Road        string `json:"road"`
			Distance    string `json:"distance"`
			Duration    string `json:"duration"`
			Polyline    string `json:"polyline"`
		}, len(path.Steps))
		for i, step := range path.Steps {
			stepPolyline := step.Polyline
			// 如果step的polyline为空，尝试从navi字段中提取
			if stepPolyline == "" && len(step.Navi) > 0 {
				var naviData struct {
					Polyline string `json:"polyline"`
				}
				if err := json.Unmarshal(step.Navi, &naviData); err == nil && naviData.Polyline != "" {
					stepPolyline = naviData.Polyline
					log.Printf("[PlanRoute] step[%d] 从navi字段提取polyline，长度: %d\n", i, len(stepPolyline))
				} else if err != nil {
					log.Printf("[PlanRoute] step[%d] 解析navi字段失败: %v\n", i, err)
				}
			}
			stepList[i] = struct {
				Instruction string `json:"instruction"`
				Road        string `json:"road"`
				Distance    string `json:"distance"`
				Duration    string `json:"duration"`
				Polyline    string `json:"polyline"`
			}{
				Instruction: step.Instruction,
				Road:        step.Road,
				Distance:    step.Distance,
				Duration:    step.Duration,
				Polyline:    stepPolyline,
			}
			if stepPolyline != "" {
				preview := stepPolyline
				if len(preview) > 50 {
					preview = preview[:50] + "..."
				}
				log.Printf("[PlanRoute] step[%d] polyline长度: %d, 预览: %s\n", i, len(stepPolyline), preview)
			} else {
				log.Printf("[PlanRoute] step[%d] polyline为空\n", i)
			}
		}

		// 如果path级别的polyline为空，从steps中拼接
		if polylineStr == "" && len(stepList) > 0 {
			log.Printf("[PlanRoute] path.polyline为空，从steps中拼接polyline\n")
			polylineParts := make([]string, 0)
			for i, step := range stepList {
				if step.Polyline != "" {
					polylineParts = append(polylineParts, step.Polyline)
					log.Printf("[PlanRoute] step[%d] polyline长度: %d, 坐标点数: %d\n", i, len(step.Polyline), len(strings.Split(step.Polyline, ";")))
				}
			}
			if len(polylineParts) > 0 {
				// 拼接时，每个step的polyline之间用分号连接
				// 注意：每个step的polyline本身已经是"经度,纬度;经度,纬度;..."格式
				// 所以拼接后仍然是正确的格式
				polylineStr = strings.Join(polylineParts, ";")
				polylinePointCount := len(strings.Split(polylineStr, ";"))
				log.Printf("[PlanRoute] 从steps拼接的polyline长度: %d, 分段数: %d, 坐标点数: %d\n", len(polylineStr), len(polylineParts), polylinePointCount)
				if len(waypoints) > 0 {
					log.Printf("[PlanRoute] 途经点数量: %d, 期望polyline应该经过所有途经点\n", len(waypoints))
					// 检查polyline是否经过途经点（通过检查坐标是否接近）
					for i, wp := range waypoints {
						wpStr := fmt.Sprintf("%.6f,%.6f", wp.Longitude, wp.Latitude)
						if strings.Contains(polylineStr, wpStr) {
							log.Printf("[PlanRoute] 途经点[%d]坐标 %s 在polyline中找到\n", i+1, wpStr)
						} else {
							log.Printf("[PlanRoute] 警告: 途经点[%d]坐标 %s 在polyline中未找到（可能坐标精度不同）\n", i+1, wpStr)
						}
					}
				}
			}
		} else if polylineStr != "" {
			// path级别的polyline不为空，检查是否包含所有途经点
			polylinePointCount := len(strings.Split(polylineStr, ";"))
			log.Printf("[PlanRoute] path.polyline包含 %d 个坐标点\n", polylinePointCount)
			if len(waypoints) > 0 {
				log.Printf("[PlanRoute] 途经点数量: %d, 期望polyline应该经过所有途经点\n", len(waypoints))
				// 检查polyline是否经过途经点
				for i, wp := range waypoints {
					wpStr := fmt.Sprintf("%.6f,%.6f", wp.Longitude, wp.Latitude)
					if strings.Contains(polylineStr, wpStr) {
						log.Printf("[PlanRoute] 途经点[%d]坐标 %s 在polyline中找到\n", i+1, wpStr)
					} else {
						log.Printf("[PlanRoute] 警告: 途经点[%d]坐标 %s 在polyline中未找到（可能坐标精度不同）\n", i+1, wpStr)
					}
				}
			}
		}
	} else if len(result.Routes) > 0 {
		// v3格式兼容
		log.Printf("[PlanRoute] 使用v3格式，找到 %d 条路线\n", len(result.Routes))
		route := result.Routes[0]
		distanceStr = route.Distance
		durationStr = route.Duration
		polylineStr = route.Polyline
		stepList = route.Steps
		log.Printf("[PlanRoute] 路线信息 - 距离: %s米, 耗时: %s秒, 步骤数: %d, route.polyline长度: %d\n",
			distanceStr, durationStr, len(stepList), len(polylineStr))

		// 如果route级别的polyline为空，从steps中拼接
		if polylineStr == "" && len(stepList) > 0 {
			log.Printf("[PlanRoute] route.polyline为空，从steps中拼接polyline\n")
			log.Printf("[PlanRoute] steps数量: %d\n", len(stepList))
			polylineParts := make([]string, 0)
			for i, step := range stepList {
				log.Printf("[PlanRoute] step[%d] polyline长度: %d, 内容预览: %s\n",
					i, len(step.Polyline),
					func() string {
						if len(step.Polyline) > 100 {
							return step.Polyline[:100] + "..."
						}
						return step.Polyline
					}())
				if step.Polyline != "" {
					polylineParts = append(polylineParts, step.Polyline)
				}
			}
			if len(polylineParts) > 0 {
				polylineStr = strings.Join(polylineParts, ";")
				log.Printf("[PlanRoute] 从steps拼接的polyline长度: %d, 分段数: %d\n", len(polylineStr), len(polylineParts))
			} else {
				log.Printf("[PlanRoute] 警告: 所有steps的polyline都为空，无法拼接\n")
			}
		}
	} else {
		log.Printf("[PlanRoute] 未找到可用路线 - Route.Paths: %d, Routes: %d\n", len(result.Route.Paths), len(result.Routes))
		return &RoutePlanningResult{
			Success: false,
			Message: "未找到可用路线",
		}, nil
	}

	var distance, duration int
	fmt.Sscanf(distanceStr, "%d", &distance)
	fmt.Sscanf(durationStr, "%d", &duration)

	// 解析步骤
	steps := make([]*RouteStep, 0, len(stepList))
	for _, step := range stepList {
		var stepDistance, stepDuration int
		fmt.Sscanf(step.Distance, "%d", &stepDistance)
		fmt.Sscanf(step.Duration, "%d", &stepDuration)

		steps = append(steps, &RouteStep{
			Instruction: step.Instruction,
			Road:        step.Road,
			Distance:    stepDistance,
			Duration:    stepDuration,
			Polyline:    step.Polyline,
		})
	}

	// 提取途经点信息（如果API返回了优化后的途经点顺序）
	waypointInfos := make([]*WaypointInfo, 0)
	if len(result.Route.Waypoints) > 0 {
		log.Printf("[PlanRoute] API返回了 %d 个途经点（优化后的顺序）\n", len(result.Route.Waypoints))
		for i, wp := range result.Route.Waypoints {
			waypointInfos = append(waypointInfos, &WaypointInfo{
				Location: wp.Location,
				Name:     fmt.Sprintf("途经点%d", i+1),
			})
			log.Printf("[PlanRoute] 途经点[%d]: %s\n", i+1, wp.Location)
		}
	} else if len(waypoints) > 0 {
		// 如果API没有返回途经点信息，使用原始途经点列表
		log.Printf("[PlanRoute] API未返回途经点信息，使用原始途经点列表\n")
		for i, wp := range waypoints {
			waypointInfos = append(waypointInfos, &WaypointInfo{
				Location: fmt.Sprintf("%.6f,%.6f", wp.Longitude, wp.Latitude),
				Name:     fmt.Sprintf("途经点%d", i+1),
			})
		}
	}

	routeInfo := &RouteInfo{
		Distance: distance,
		Duration: duration,
		Polyline: polylineStr,
		Steps:    steps,
	}

	return &RoutePlanningResult{
		Success:   true,
		Route:     routeInfo,
		Routes:    []*RouteInfo{routeInfo},
		Polyline:  polylineStr,
		Distance:  distance,
		Duration:  duration,
		Steps:     steps,
		Waypoints: waypointInfos,
	}, nil
}
