package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// WeatherData 天气数据
type WeatherData struct {
	Temperature   float64 `json:"temperature"`   // 温度（摄氏度）
	Condition     string  `json:"condition"`     // 天气状况（如：晴、雨、雪等）
	Precipitation float64 `json:"precipitation"` // 降水量（毫米）
	Success       bool    `json:"success"`
	Message       string  `json:"message,omitempty"`
}

// GetWeatherByLocation 根据经纬度获取天气信息
// 优先使用高德地图天气API，如果没有配置则使用其他天气服务
func GetWeatherByLocation(latitude, longitude float64, amapKey string) (*WeatherData, error) {
	if amapKey == "" {
		// 如果没有配置API Key，返回默认数据（正常天气）
		return &WeatherData{
			Temperature:   25.0,
			Condition:     "晴",
			Precipitation: 0,
			Success:       false,
			Message:       "未配置天气API Key",
		}, nil
	}

	// 使用高德地图天气查询API
	return getWeatherByAmap(latitude, longitude, amapKey)
}

// getWeatherByAmap 使用高德地图API获取天气
func getWeatherByAmap(latitude, longitude float64, apiKey string) (*WeatherData, error) {
	// 高德地图天气查询API
	apiURL := "https://restapi.amap.com/v3/weather/weatherInfo"

	params := url.Values{}
	params.Set("key", apiKey)
	params.Set("location", fmt.Sprintf("%.6f,%.6f", longitude, latitude))
	params.Set("output", "json")
	params.Set("extensions", "all") // 获取详细信息

	reqURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := http.Get(reqURL)
	if err != nil {
		return &WeatherData{
			Success: false,
			Message: fmt.Sprintf("请求天气API失败: %v", err),
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &WeatherData{
			Success: false,
			Message: fmt.Sprintf("读取响应失败: %v", err),
		}, err
	}

	var result struct {
		Status string `json:"status"`
		Count  string `json:"count"`
		Info   string `json:"info"`
		Forecasts []struct {
			Casts []struct {
				DayTemp    string `json:"daytemp"`    // 白天温度
				NightTemp  string `json:"nighttemp"`  // 夜间温度
				DayWeather string `json:"dayweather"` // 白天天气
				DayWind    string `json:"daywind"`    // 白天风向
			} `json:"casts"`
		} `json:"forecasts"`
		Lives []struct {
			Temperature string `json:"temperature"` // 实时温度
			Weather     string `json:"weather"`     // 天气状况
			WindPower   string `json:"windpower"`   // 风力
		} `json:"lives"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return &WeatherData{
			Success: false,
			Message: fmt.Sprintf("解析响应失败: %v", err),
		}, err
	}

	if result.Status != "1" || result.Count == "0" {
		return &WeatherData{
			Success: false,
			Message: fmt.Sprintf("天气查询失败: %s", result.Info),
		}, nil
	}

	// 优先使用实时天气数据
	var temperature float64
	var condition string

	if len(result.Lives) > 0 {
		live := result.Lives[0]
		if temp, err := parseFloat(live.Temperature); err == nil {
			temperature = temp
		}
		condition = live.Weather
	} else if len(result.Forecasts) > 0 && len(result.Forecasts[0].Casts) > 0 {
		cast := result.Forecasts[0].Casts[0]
		// 使用白天温度
		if temp, err := parseFloat(cast.DayTemp); err == nil {
			temperature = temp
		}
		condition = cast.DayWeather
	}

	// 判断是否有降水（从天气状况字符串判断）
	precipitation := 0.0
	conditionLower := strings.ToLower(condition)
	if strings.Contains(conditionLower, "雨") || strings.Contains(conditionLower, "rain") {
		// 简单判断：小雨0.5，中雨1.0，大雨2.0
		if strings.Contains(conditionLower, "大") || strings.Contains(conditionLower, "heavy") {
			precipitation = 2.0
		} else if strings.Contains(conditionLower, "中") || strings.Contains(conditionLower, "moderate") {
			precipitation = 1.0
		} else {
			precipitation = 0.5
		}
	}
	if strings.Contains(conditionLower, "雪") || strings.Contains(conditionLower, "snow") {
		precipitation = 1.0
	}

	return &WeatherData{
		Temperature:   temperature,
		Condition:     condition,
		Precipitation: precipitation,
		Success:       true,
	}, nil
}

// parseFloat 解析字符串为浮点数
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// IsExtremeWeather 判断是否为极端天气
// 极端天气包括：雨雪（降水量>0.5mm）或高温（>extremeTemp）
func IsExtremeWeather(weather *WeatherData, extremeTemp float64) bool {
	if weather == nil || !weather.Success {
		return false
	}

	// 判断雨雪
	conditionLower := strings.ToLower(weather.Condition)
	if (strings.Contains(conditionLower, "雨") || strings.Contains(conditionLower, "rain") ||
		strings.Contains(conditionLower, "雪") || strings.Contains(conditionLower, "snow")) &&
		weather.Precipitation > 0.5 {
		return true
	}

	// 判断高温
	if weather.Temperature > extremeTemp {
		return true
	}

	return false
}

