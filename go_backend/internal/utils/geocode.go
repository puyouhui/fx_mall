package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// GeocodeResult 地址解析结果
type GeocodeResult struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"` // 逆地理编码时返回地址
	Success   bool    `json:"success"`
	Message   string  `json:"message,omitempty"`
}

// GeocodeAddress 地址解析（使用高德地图API或腾讯地图API）
// address: 地址文本，如 "北京市朝阳区xxx路xxx号"
// amapKey: 高德地图API Key（可选）
// tencentKey: 腾讯地图API Key（可选）
// 返回经纬度坐标
func GeocodeAddress(address, amapKey, tencentKey string) (*GeocodeResult, error) {
	// 如果地址为空，直接返回失败
	address = strings.TrimSpace(address)
	if address == "" {
		return &GeocodeResult{
			Success: false,
			Message: "地址不能为空",
		}, nil
	}

	// 优先使用高德地图API，如果没有配置则使用腾讯地图API
	if amapKey != "" {
		return geocodeByAmap(address, amapKey)
	}

	// 如果没有配置高德地图Key，尝试使用腾讯地图
	if tencentKey != "" {
		return geocodeByTencent(address, tencentKey)
	}

	// 如果都没有配置，返回错误
	return &GeocodeResult{
		Success: false,
		Message: "未配置地图API Key，无法进行地址解析。请在系统设置中配置高德地图或腾讯地图API Key",
	}, nil
}

// geocodeByAmap 使用高德地图API进行地址解析
func geocodeByAmap(address, apiKey string) (*GeocodeResult, error) {
	// 高德地图地理编码API
	apiURL := "https://restapi.amap.com/v3/geocode/geo"

	params := url.Values{}
	params.Set("key", apiKey)
	params.Set("address", address)
	params.Set("output", "json")

	reqURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := http.Get(reqURL)
	if err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("请求高德地图API失败: %v", err),
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("读取响应失败: %v", err),
		}, err
	}

	var result struct {
		Status   string `json:"status"`
		Count    string `json:"count"`
		Geocodes []struct {
			Location string `json:"location"`
		} `json:"geocodes"`
		Info     string `json:"info"`
		Infocode string `json:"infocode"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("解析响应失败: %v", err),
		}, err
	}

	if result.Status != "1" || result.Count == "0" || len(result.Geocodes) == 0 {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("地址解析失败: %s", result.Info),
		}, nil
	}

	// 解析经纬度（高德返回格式：经度,纬度）
	location := result.Geocodes[0].Location
	parts := strings.Split(location, ",")
	if len(parts) != 2 {
		return &GeocodeResult{
			Success: false,
			Message: "解析经纬度格式错误",
		}, nil
	}

	var longitude, latitude float64
	if _, err := fmt.Sscanf(parts[0], "%f", &longitude); err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("解析经度失败: %v", err),
		}, nil
	}
	if _, err := fmt.Sscanf(parts[1], "%f", &latitude); err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("解析纬度失败: %v", err),
		}, nil
	}

	return &GeocodeResult{
		Latitude:  latitude,
		Longitude: longitude,
		Success:   true,
	}, nil
}

// geocodeByTencent 使用腾讯地图API进行地址解析
func geocodeByTencent(address, tencentKey string) (*GeocodeResult, error) {

	apiURL := "https://apis.map.qq.com/ws/geocoder/v1/"

	params := url.Values{}
	params.Set("key", tencentKey)
	params.Set("address", address)
	params.Set("output", "json")

	reqURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := http.Get(reqURL)
	if err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("请求腾讯地图API失败: %v", err),
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("读取响应失败: %v", err),
		}, err
	}

	var result struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("解析响应失败: %v", err),
		}, err
	}

	if result.Status != 0 {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("地址解析失败: %s", result.Message),
		}, nil
	}

	return &GeocodeResult{
		Latitude:  result.Result.Location.Lat,
		Longitude: result.Result.Location.Lng,
		Success:   true,
	}, nil
}

// ReverseGeocode 逆地理编码（根据经纬度获取地址）
// longitude: 经度
// latitude: 纬度
// amapKey: 高德地图API Key（可选）
// tencentKey: 腾讯地图API Key（可选）
// 返回地址字符串
func ReverseGeocode(longitude, latitude float64, amapKey, tencentKey string) (*GeocodeResult, error) {
	// 优先使用高德地图API，如果没有配置则使用腾讯地图API
	if amapKey != "" {
		return reverseGeocodeByAmap(longitude, latitude, amapKey)
	}

	// 如果没有配置高德地图Key，尝试使用腾讯地图
	if tencentKey != "" {
		return reverseGeocodeByTencent(longitude, latitude, tencentKey)
	}

	// 如果都没有配置，返回错误
	return &GeocodeResult{
		Success: false,
		Message: "未配置地图API Key，无法进行逆地理编码。请在系统设置中配置高德地图或腾讯地图API Key",
	}, nil
}

// reverseGeocodeByAmap 使用高德地图API进行逆地理编码
func reverseGeocodeByAmap(longitude, latitude float64, apiKey string) (*GeocodeResult, error) {
	// 高德地图逆地理编码API
	apiURL := "https://restapi.amap.com/v3/geocode/regeo"

	params := url.Values{}
	params.Set("key", apiKey)
	params.Set("location", fmt.Sprintf("%.6f,%.6f", longitude, latitude))
	params.Set("output", "json")
	params.Set("radius", "1000")
	params.Set("extensions", "all")

	reqURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := http.Get(reqURL)
	if err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("请求高德地图逆地理编码API失败: %v", err),
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("读取高德地图逆地理编码API响应失败: %v", err),
		}, err
	}

	var result struct {
		Status    string `json:"status"`
		Info      string `json:"info"`
		Regeocode struct {
			FormattedAddress string `json:"formatted_address"`
		} `json:"regeocode"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("解析高德地图逆地理编码API响应失败: %v", err),
		}, err
	}

	if result.Status != "1" {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("高德地图逆地理编码API返回错误: %s", result.Info),
		}, nil
	}

	return &GeocodeResult{
		Latitude:  latitude,
		Longitude: longitude,
		Address:   result.Regeocode.FormattedAddress,
		Success:   true,
	}, nil
}

// reverseGeocodeByTencent 使用腾讯地图API进行逆地理编码
func reverseGeocodeByTencent(longitude, latitude float64, apiKey string) (*GeocodeResult, error) {
	// 腾讯地图逆地理编码API
	apiURL := "https://apis.map.qq.com/ws/geocoder/v1/"

	params := url.Values{}
	params.Set("key", apiKey)
	params.Set("location", fmt.Sprintf("%.6f,%.6f", latitude, longitude))
	params.Set("get_poi", "0")

	reqURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := http.Get(reqURL)
	if err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("请求腾讯地图逆地理编码API失败: %v", err),
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("读取腾讯地图逆地理编码API响应失败: %v", err),
		}, err
	}

	var result struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Address string `json:"address"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("解析腾讯地图逆地理编码API响应失败: %v", err),
		}, err
	}

	if result.Status != 0 {
		return &GeocodeResult{
			Success: false,
			Message: fmt.Sprintf("腾讯地图逆地理编码API返回错误: %s", result.Message),
		}, nil
	}

	return &GeocodeResult{
		Latitude:  latitude,
		Longitude: longitude,
		Address:   result.Result.Address,
		Success:   true,
	}, nil
}
