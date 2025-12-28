package api

import (
	"net/http"
	"strconv"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetPointsLogs 获取用户积分明细列表
func GetPointsLogs(c *gin.Context) {
	// 从上下文获取用户ID（由认证中间件设置）
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	userID, ok := userIDInterface.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户ID格式错误"})
		return
	}

	pageNumStr := c.DefaultQuery("page_num", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	logs, total, err := model.GetPointsLogs(userID, pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取积分明细失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":      logs,
			"total":     total,
			"page_num":  pageNum,
			"page_size": pageSize,
		},
	})
}

