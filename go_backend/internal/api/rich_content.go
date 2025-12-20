package api

import (
	"go_backend/internal/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateRichContentRequest 创建富文本内容请求
type CreateRichContentRequest struct {
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
	ContentType string `json:"content_type"` // notice, activity, other
}

// UpdateRichContentRequest 更新富文本内容请求
type UpdateRichContentRequest struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	Status      string `json:"status"` // draft, published, archived
}

// CreateRichContent 创建富文本内容
func CreateRichContent(c *gin.Context) {
	var req CreateRichContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文获取管理员信息
	adminUsername, _ := c.Get("username")

	richContent := &model.RichContent{
		Title:       req.Title,
		Content:     req.Content,
		ContentType: req.ContentType,
		CreatedBy:   adminUsername.(string),
		UpdatedBy:   adminUsername.(string),
	}

	if err := model.CreateRichContent(richContent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建富文本内容失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    richContent,
	})
}

// GetRichContent 获取富文本内容详情
func GetRichContent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	content, err := model.GetRichContentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "富文本内容不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": content,
	})
}

// GetRichContentList 获取富文本内容列表
func GetRichContentList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	contentType := c.Query("content_type")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	contents, total, err := model.GetAllRichContents(page, pageSize, contentType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取富文本内容列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": contents,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// UpdateRichContent 更新富文本内容
func UpdateRichContent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req UpdateRichContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文获取管理员信息
	adminUsername, _ := c.Get("username")

	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.ContentType != "" {
		updates["content_type"] = req.ContentType
	}
	if req.Status != "" {
		updates["status"] = req.Status
		if req.Status == "published" {
			now := time.Now()
			updates["published_at"] = now
		}
	}
	updates["updated_by"] = adminUsername.(string)

	if err := model.UpdateRichContent(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新富文本内容失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// PublishRichContent 发布富文本内容
func PublishRichContent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	// 从上下文获取管理员信息
	adminUsername, _ := c.Get("username")

	if err := model.PublishRichContent(id, adminUsername.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布富文本内容失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "发布成功",
	})
}

// ArchiveRichContent 归档富文本内容
func ArchiveRichContent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	// 从上下文获取管理员信息
	adminUsername, _ := c.Get("username")

	if err := model.ArchiveRichContent(id, adminUsername.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "归档富文本内容失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "归档成功",
	})
}

// DeleteRichContent 删除富文本内容
func DeleteRichContent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := model.DeleteRichContent(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除富文本内容失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// ========== 小程序端API ==========

// GetPublishedRichContentList 获取已发布的富文本内容列表（小程序端）
func GetPublishedRichContentList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	contentType := c.Query("content_type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	contents, total, err := model.GetPublishedRichContents(page, pageSize, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取富文本内容列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": contents,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetPublishedRichContentDetail 获取已发布的富文本内容详情（小程序端）
func GetPublishedRichContentDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	// 获取内容并增加浏览次数
	content, err := model.GetRichContentByIDAndIncrementView(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "富文本内容不存在"})
		return
	}

	// 只返回已发布的内容
	if content.Status != "published" {
		c.JSON(http.StatusNotFound, gin.H{"error": "富文本内容不存在或未发布"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": content,
	})
}
