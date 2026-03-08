package api

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetUomDefaultCategory 获取默认「件」单位类别ID（用于老数据兼容）
func GetUomDefaultCategory(c *gin.Context) {
	id, err := model.GetDefaultUomCategoryID()
	if err != nil {
		log.Printf("获取默认单位类别失败: %v", err)
		internalErrorResponse(c, "获取默认单位类别失败")
		return
	}
	successResponse(c, gin.H{"id": id}, "")
}

// GetUomCategories 获取所有单位类别（含单位列表）
func GetUomCategories(c *gin.Context) {
	list, err := model.GetAllUomCategories()
	if err != nil {
		log.Printf("获取单位类别失败: %v", err)
		internalErrorResponse(c, "获取单位类别失败: "+err.Error())
		return
	}
	successResponse(c, list, "")
}

// CreateUomCategory 创建单位类别
func CreateUomCategory(c *gin.Context) {
	var req model.UomCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		badRequestResponse(c, "类别名称不能为空")
		return
	}
	exists, err := model.CheckUomCategoryNameExists(req.Name, 0)
	if err != nil {
		internalErrorResponse(c, err.Error())
		return
	}
	if exists {
		badRequestResponse(c, "类别名称已存在")
		return
	}
	if err := model.CreateUomCategory(&req); err != nil {
		log.Printf("创建单位类别失败: %v", err)
		internalErrorResponse(c, "创建单位类别失败: "+err.Error())
		return
	}
	successResponse(c, req, "创建成功")
}

// UpdateUomCategory 更新单位类别
func UpdateUomCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		badRequestResponse(c, "无效的ID")
		return
	}
	var req model.UomCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}
	req.ID = id
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		badRequestResponse(c, "类别名称不能为空")
		return
	}
	exists, _ := model.CheckUomCategoryNameExists(req.Name, id)
	if exists {
		badRequestResponse(c, "类别名称已存在")
		return
	}
	if err := model.UpdateUomCategory(&req); err != nil {
		log.Printf("更新单位类别失败: %v", err)
		internalErrorResponse(c, "更新单位类别失败: "+err.Error())
		return
	}
	successResponse(c, nil, "更新成功")
}

// DeleteUomCategory 删除单位类别
func DeleteUomCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		badRequestResponse(c, "无效的ID")
		return
	}
	if err := model.DeleteUomCategory(id); err != nil {
		log.Printf("删除单位类别失败: %v", err)
		internalErrorResponse(c, "删除单位类别失败: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// GetUomUnits 获取某类别的单位列表
func GetUomUnits(c *gin.Context) {
	categoryID, err := strconv.Atoi(c.Query("category_id"))
	if err != nil || categoryID <= 0 {
		badRequestResponse(c, "无效的category_id")
		return
	}
	list, err := model.GetUomUnitsByCategoryID(categoryID)
	if err != nil {
		log.Printf("获取单位列表失败: %v", err)
		internalErrorResponse(c, "获取单位列表失败: "+err.Error())
		return
	}
	successResponse(c, list, "")
}

// CreateUomUnit 创建单位
func CreateUomUnit(c *gin.Context) {
	var req model.UomUnit
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		badRequestResponse(c, "单位名称不能为空")
		return
	}
	if req.CategoryID <= 0 {
		badRequestResponse(c, "请选择单位类别")
		return
	}
	if req.Ratio <= 0 {
		badRequestResponse(c, "换算比例必须大于0")
		return
	}
	// 若设为基准单位，强制 ratio=1
	if req.IsBase == 1 {
		req.Ratio = 1
	}
	exists, _ := model.CheckUomUnitNameExists(req.CategoryID, req.Name, 0)
	if exists {
		badRequestResponse(c, "该类别下已存在同名单位")
		return
	}
	if req.IsBase == 1 {
		cnt, _ := model.CountBaseUnitsInCategory(req.CategoryID)
		if cnt > 0 {
			badRequestResponse(c, "该类别已存在基准单位，一个类别只能有一个基准单位")
			return
		}
	}
	if err := model.CreateUomUnit(&req); err != nil {
		log.Printf("创建单位失败: %v", err)
		internalErrorResponse(c, "创建单位失败: "+err.Error())
		return
	}
	successResponse(c, req, "创建成功")
}

// UpdateUomUnit 更新单位
func UpdateUomUnit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		badRequestResponse(c, "无效的ID")
		return
	}
	unit, err := model.GetUomUnitByID(id)
	if err != nil || unit == nil {
		notFoundResponse(c, "单位不存在")
		return
	}
	var req model.UomUnit
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}
	req.ID = id
	req.CategoryID = unit.CategoryID
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		badRequestResponse(c, "单位名称不能为空")
		return
	}
	if req.Ratio <= 0 {
		badRequestResponse(c, "换算比例必须大于0")
		return
	}
	if req.IsBase == 1 {
		req.Ratio = 1
	}
	exists, _ := model.CheckUomUnitNameExists(req.CategoryID, req.Name, id)
	if exists {
		badRequestResponse(c, "该类别下已存在同名单位")
		return
	}
	if req.IsBase == 1 && unit.IsBase != 1 {
		cnt, _ := model.CountBaseUnitsInCategory(req.CategoryID)
		if cnt > 0 {
			badRequestResponse(c, "该类别已存在基准单位")
			return
		}
	}
	if err := model.UpdateUomUnit(&req); err != nil {
		log.Printf("更新单位失败: %v", err)
		internalErrorResponse(c, "更新单位失败: "+err.Error())
		return
	}
	successResponse(c, nil, "更新成功")
}

// DeleteUomUnit 删除单位
func DeleteUomUnit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		badRequestResponse(c, "无效的ID")
		return
	}
	if err := model.DeleteUomUnit(id); err != nil {
		log.Printf("删除单位失败: %v", err)
		internalErrorResponse(c, "删除单位失败: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

