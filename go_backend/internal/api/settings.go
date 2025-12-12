package api

import (
	"fmt"
	"net/http"

	"go_backend/internal/config"
	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetSystemSettings 获取所有系统设置
func GetSystemSettings(c *gin.Context) {
	settings, err := model.GetAllSystemSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取系统设置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    settings,
	})
}

// UpdateSystemSettings 更新系统设置
type updateSystemSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

func UpdateSystemSettings(c *gin.Context) {
	var req updateSystemSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 更新每个设置项
	for key, value := range req.Settings {
		description := getSettingDescription(key)
		if err := model.SetSystemSetting(key, value, description); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新设置失败: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// GetMapSettings 获取地图相关设置
func GetMapSettings(c *gin.Context) {
	settings, err := model.GetMapSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取地图设置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    settings,
	})
}

// UpdateMapSettings 更新地图相关设置
type updateMapSettingsRequest struct {
	AmapKey    string `json:"amap_key"`
	TencentKey string `json:"tencent_key"`
}

func UpdateMapSettings(c *gin.Context) {
	var req updateMapSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 更新高德地图 Key
	if err := model.SetSystemSetting("map_amap_key", req.AmapKey, "高德地图API Key"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新高德地图Key失败: " + err.Error(),
		})
		return
	}

	// 更新腾讯地图 Key
	if err := model.SetSystemSetting("map_tencent_key", req.TencentKey, "腾讯地图API Key"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新腾讯地图Key失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// GetWebSocketConfig 获取WebSocket配置（供前端使用）
func GetWebSocketConfig(c *gin.Context) {
	// 获取请求的协议和主机
	scheme := "ws"
	// 检查是否使用HTTPS（支持代理场景）
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "wss"
	}
	host := c.Request.Host
	if host == "" {
		host = fmt.Sprintf("localhost:%d", config.Config.Server.Port)
	}

	// 构建完整的WebSocket URL
	employeeLocationURL := fmt.Sprintf("%s://%s%s", scheme, host, config.Config.WebSocket.EmployeeLocationURL)
	adminLocationURL := fmt.Sprintf("%s://%s%s", scheme, host, config.Config.WebSocket.AdminLocationURL)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": map[string]string{
			"employee_location_url": employeeLocationURL,
			"admin_location_url":    adminLocationURL,
		},
	})
}

// getSettingDescription 获取设置项的说明
func getSettingDescription(key string) string {
	descriptions := map[string]string{
		"map_amap_key":                 "高德地图API Key",
		"map_tencent_key":              "腾讯地图API Key",
		"order_urgent_fee":             "加急订单费用（元）",
		"delivery_base_fee":            "基础配送费（元）",
		"delivery_isolated_distance":   "孤立订单判断距离（公里）",
		"delivery_isolated_subsidy":    "孤立订单补贴（元）",
		"delivery_item_threshold_low":  "件数补贴低阈值（件）",
		"delivery_item_rate_low":       "件数补贴低档费率（元/件）",
		"delivery_item_threshold_high": "件数补贴高阈值（件）",
		"delivery_item_rate_high":      "件数补贴高档费率（元/件）",
		"delivery_item_max_count":      "件数补贴最大计件数",
		"delivery_urgent_subsidy":      "加急订单补贴（元）",
		"delivery_weather_subsidy":     "极端天气补贴（元）",
		"delivery_extreme_temp":        "极端高温阈值（摄氏度）",
		"delivery_profit_threshold":    "利润分成阈值（元）",
		"delivery_profit_share_rate":   "利润分成比例（8%）",
		"delivery_max_profit_share":    "利润分成上限（元）",
	}
	if desc, ok := descriptions[key]; ok {
		return desc
	}
	return ""
}
