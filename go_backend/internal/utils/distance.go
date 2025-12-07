package utils

import "math"

// CalculateDistance 计算两个经纬度坐标之间的距离（公里）
// 使用 Haversine 公式
// lat1, lng1: 第一个点的纬度和经度
// lat2, lng2: 第二个点的纬度和经度
// 返回距离（公里）
func CalculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	// 地球半径（公里）
	const earthRadius = 6371.0

	// 转换为弧度
	lat1Rad := lat1 * math.Pi / 180.0
	lng1Rad := lng1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	lng2Rad := lng2 * math.Pi / 180.0

	// 计算差值
	deltaLat := lat2Rad - lat1Rad
	deltaLng := lng2Rad - lng1Rad

	// Haversine 公式
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// 计算距离
	distance := earthRadius * c

	return distance
}

