package utils

import (
	"fmt"
	"math"
	"sync"
)

// Location 位置信息
type Location struct {
	Name  string  // 位置名称（可以是订单ID或地址名称）
	Lat   float64 // 纬度
	Lng   float64 // 经度
	Index int     // 索引（内部使用）
}

// DeliveryRouteOptimizer 配送路线优化器
type DeliveryRouteOptimizer struct {
	locations     []Location
	nameToIndex   map[string]int
	distanceCache map[routeKey]float64
	mu            sync.RWMutex
}

type routeKey struct {
	Idx1 int
	Idx2 int
}

// NewDeliveryRouteOptimizer 创建新的路线优化器
func NewDeliveryRouteOptimizer() *DeliveryRouteOptimizer {
	return &DeliveryRouteOptimizer{
		locations:     make([]Location, 0),
		nameToIndex:   make(map[string]int),
		distanceCache: make(map[routeKey]float64),
	}
}

// AddLocation 添加位置；若已存在返回其索引
func (o *DeliveryRouteOptimizer) AddLocation(name string, lat, lng float64) int {
	o.mu.Lock()
	defer o.mu.Unlock()

	if idx, exists := o.nameToIndex[name]; exists {
		return idx
	}

	idx := len(o.locations)
	o.locations = append(o.locations, Location{
		Name:  name,
		Lat:   lat,
		Lng:   lng,
		Index: idx,
	})
	o.nameToIndex[name] = idx
	return idx
}

// CalculateDistance 计算两点球面距离（公里），带缓存（公开方法）
func (o *DeliveryRouteOptimizer) CalculateDistance(idx1, idx2 int) float64 {
	// 确保 idx1 < idx2 以统一缓存键
	if idx1 > idx2 {
		idx1, idx2 = idx2, idx1
	}
	key := routeKey{Idx1: idx1, Idx2: idx2}

	o.mu.RLock()
	if dist, exists := o.distanceCache[key]; exists {
		o.mu.RUnlock()
		return dist
	}
	o.mu.RUnlock()

	// 计算距离
	a := o.locations[idx1]
	b := o.locations[idx2]
	R := 6371.0 // 地球半径（公里）

	lat1 := a.Lat * math.Pi / 180.0
	lon1 := a.Lng * math.Pi / 180.0
	lat2 := b.Lat * math.Pi / 180.0
	lon2 := b.Lng * math.Pi / 180.0

	dlat := lat2 - lat1
	dlon := lon2 - lon1

	h := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	dist := R * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))

	o.mu.Lock()
	o.distanceCache[key] = dist
	o.mu.Unlock()

	return dist
}

// NearestNeighbor 最近邻生成初始路线
func (o *DeliveryRouteOptimizer) NearestNeighbor(startIdx int) ([]int, float64) {
	n := len(o.locations)
	if n <= 1 {
		return []int{startIdx}, 0.0
	}

	unvisited := make(map[int]bool)
	for i := 0; i < n; i++ {
		if i != startIdx {
			unvisited[i] = true
		}
	}

	route := []int{startIdx}
	total := 0.0
	current := startIdx

	for len(unvisited) > 0 {
		nearestIdx := -1
		nearestDist := math.MaxFloat64
		// 收集所有距离相同的候选点，用于稳定排序
		candidates := make([]int, 0)

		for next := range unvisited {
			d := o.CalculateDistance(current, next)
			if d < nearestDist {
				nearestDist = d
				nearestIdx = next
				// 重置候选列表，因为找到了更近的点
				candidates = []int{next}
			} else if d == nearestDist && nearestDist < math.MaxFloat64 {
				// 如果距离相同，添加到候选列表
				candidates = append(candidates, next)
			}
		}

		// 如果有多个距离相同的候选点，按索引排序选择最小的（稳定排序）
		if len(candidates) > 1 {
			// 对候选点按索引排序
			for i := 0; i < len(candidates)-1; i++ {
				for j := i + 1; j < len(candidates); j++ {
					if candidates[i] > candidates[j] {
						candidates[i], candidates[j] = candidates[j], candidates[i]
					}
				}
			}
			nearestIdx = candidates[0]
		}

		if nearestIdx == -1 {
			break
		}

		route = append(route, nearestIdx)
		total += nearestDist
		delete(unvisited, nearestIdx)
		current = nearestIdx
	}

	return route, total
}

// routeDistance 计算路线总距离
func (o *DeliveryRouteOptimizer) routeDistance(indices []int) float64 {
	if len(indices) <= 1 {
		return 0.0
	}
	total := 0.0
	for i := 0; i < len(indices)-1; i++ {
		total += o.CalculateDistance(indices[i], indices[i+1])
	}
	return total
}

// TwoOptOptimization 2-opt 改善路线
func (o *DeliveryRouteOptimizer) TwoOptOptimization(route []int, iterations int) ([]int, float64) {
	bestRoute := make([]int, len(route))
	copy(bestRoute, route)
	bestDistance := o.routeDistance(bestRoute)
	n := len(bestRoute)

	for iter := 0; iter < iterations; iter++ {
		improved := false
		for i := 1; i < n-2; i++ {
			for j := i + 1; j < n; j++ {
				if j-i == 1 {
					continue
				}

				// 创建新路线：反转 i 到 j 之间的部分
				newRoute := make([]int, 0, n)
				newRoute = append(newRoute, bestRoute[:i]...)
				// 反转 i 到 j 之间的部分
				for k := j - 1; k >= i; k-- {
					newRoute = append(newRoute, bestRoute[k])
				}
				newRoute = append(newRoute, bestRoute[j:]...)

				newDistance := o.routeDistance(newRoute)
				if newDistance < bestDistance {
					bestRoute = newRoute
					bestDistance = newDistance
					improved = true
				}
			}
		}
		if !improved {
			break
		}
	}

	return bestRoute, bestDistance
}

// OptimizeRoute 执行完整优化流程
// startPointName: 起点名称（必须在 locations 中存在）
// iterations: 2-opt 迭代次数
// 返回优化后的路线索引列表和总距离
func (o *DeliveryRouteOptimizer) OptimizeRoute(startPointName string, iterations int) ([]int, float64, error) {
	o.mu.RLock()
	startIdx, exists := o.nameToIndex[startPointName]
	o.mu.RUnlock()

	if !exists {
		return nil, 0, fmt.Errorf("起点 '%s' 不存在", startPointName)
	}

	// 最近邻生成初始路线
	initialRoute, _ := o.NearestNeighbor(startIdx)

	// 2-opt 优化
	optimizedRoute, optimizedDistance := o.TwoOptOptimization(initialRoute, iterations)

	return optimizedRoute, optimizedDistance, nil
}

// ClearData 清空存储
func (o *DeliveryRouteOptimizer) ClearData() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.locations = o.locations[:0]
	o.nameToIndex = make(map[string]int)
	o.distanceCache = make(map[routeKey]float64)
}

// GetLocationByIndex 根据索引获取位置信息
func (o *DeliveryRouteOptimizer) GetLocationByIndex(idx int) *Location {
	o.mu.RLock()
	defer o.mu.RUnlock()
	if idx >= 0 && idx < len(o.locations) {
		return &o.locations[idx]
	}
	return nil
}

// OptimizeDeliveryRoute 外部直接调用的简化接口
// locations: 位置列表，每个位置包含 name, lat, lng
// startPointName: 起点名称，需在 locations 中存在
// iterations: 2-opt 迭代次数
// 返回优化后的路线索引列表和总距离（公里）
func OptimizeDeliveryRoute(
	locations []struct {
		Name string
		Lat  float64
		Lng  float64
	},
	startPointName string,
	iterations int,
) ([]int, float64, error) {
	optimizer := NewDeliveryRouteOptimizer()
	for _, loc := range locations {
		optimizer.AddLocation(loc.Name, loc.Lat, loc.Lng)
	}
	return optimizer.OptimizeRoute(startPointName, iterations)
}
