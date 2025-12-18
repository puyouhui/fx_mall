package api

import (
	"net/http"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetHotSearchKeywords 获取热门搜索关键词（小程序用）
func GetHotSearchKeywords(c *gin.Context) {
	keywords, err := model.GetActiveHotSearchKeywords()
	if err != nil {
		internalErrorResponse(c, "获取热门搜索关键词失败")
		return
	}
	if keywords == nil {
		keywords = []string{}
	}
	successResponse(c, keywords, "")
}

// GetAllHotSearchKeywordsForAdmin 获取所有热门搜索关键词（管理后台用）
func GetAllHotSearchKeywordsForAdmin(c *gin.Context) {
	keywords, err := model.GetAllHotSearchKeywords()
	if err != nil {
		internalErrorResponse(c, "获取热门搜索关键词失败")
		return
	}
	if keywords == nil {
		keywords = []model.HotSearchKeyword{}
	}
	successResponse(c, keywords, "")
}

// CreateHotSearchKeyword 创建热门搜索关键词
func CreateHotSearchKeyword(c *gin.Context) {
	var req struct {
		Keyword string `json:"keyword" binding:"required"`
		Sort    int    `json:"sort"`
		Status  int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "参数错误")
		return
	}

	keyword, err := model.CreateHotSearchKeyword(req.Keyword, req.Sort, req.Status)
	if err != nil {
		internalErrorResponse(c, "创建热门搜索关键词失败")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 200, "data": keyword, "message": "创建成功"})
}

// UpdateHotSearchKeyword 更新热门搜索关键词
func UpdateHotSearchKeyword(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var req struct {
		Keyword string `json:"keyword" binding:"required"`
		Sort    int    `json:"sort"`
		Status  int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "参数错误")
		return
	}

	if err := model.UpdateHotSearchKeyword(id, req.Keyword, req.Sort, req.Status); err != nil {
		internalErrorResponse(c, "更新热门搜索关键词失败")
		return
	}

	successResponse(c, nil, "更新成功")
}

// DeleteHotSearchKeyword 删除热门搜索关键词
func DeleteHotSearchKeyword(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	if err := model.DeleteHotSearchKeyword(id); err != nil {
		internalErrorResponse(c, "删除热门搜索关键词失败")
		return
	}

	successResponse(c, nil, "删除成功")
}


