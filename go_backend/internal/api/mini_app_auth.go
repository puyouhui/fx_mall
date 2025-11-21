package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go_backend/internal/config"
	"go_backend/internal/model"
	"go_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type miniAppLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

type weChatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// MiniAppLogin 小程序登录，仅记录用户唯一ID（openid）
func MiniAppLogin(c *gin.Context) {
	var req miniAppLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "登录参数错误"})
		return
	}

	sessionInfo, err := fetchWeChatSessionInfo(req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户凭证失败: " + err.Error()})
		return
	}

	if sessionInfo.OpenID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "未能获取到用户唯一ID"})
		return
	}

	user, err := model.GetMiniAppUserByUniqueID(sessionInfo.OpenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询用户信息失败: " + err.Error()})
		return
	}

	if user == nil {
		user, err = model.CreateMiniAppUser(sessionInfo.OpenID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败: " + err.Error()})
			return
		}
	}

	token, err := utils.GenerateMiniAppToken(user.UniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成登录凭证失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"unique_id": user.UniqueID,
			"token":     token,
			"user":      user,
		},
	})
}

// GetMiniAppUsers 获取小程序用户（后台管理使用）
func GetMiniAppUsers(c *gin.Context) {
	pageNum := parseQueryInt(c, "pageNum", 1)
	pageSize := parseQueryInt(c, "pageSize", 10)
	keyword := c.Query("keyword")

	users, total, err := model.GetMiniAppUsers(pageNum, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"data":  users,
		"total": total,
	})
}

func fetchWeChatSessionInfo(code string) (*weChatSessionResponse, error) {
	appID := config.Config.MiniApp.AppID
	appSecret := config.Config.MiniApp.AppSecret
	if appID == "" || appSecret == "" {
		return nil, fmt.Errorf("未配置小程序AppID或AppSecret")
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID,
		appSecret,
		code,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var session weChatSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, err
	}

	if session.ErrCode != 0 {
		return nil, fmt.Errorf("微信登录失败: %s", session.ErrMsg)
	}

	return &session, nil
}

type updateMiniUserTypeRequest struct {
	UserType string `json:"user_type" binding:"required"`
}

// UpdateMiniAppUserType 更新小程序用户类型
func UpdateMiniAppUserType(c *gin.Context) {
	token := extractBearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	claims, err := utils.ParseMiniAppToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "登录状态已失效，请重新登录"})
		return
	}

	var req updateMiniUserTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	req.UserType = strings.ToLower(strings.TrimSpace(req.UserType))
	if req.UserType != "retail" && req.UserType != "wholesale" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户类型仅支持 retail 或 wholesale"})
		return
	}

	if err := model.UpdateMiniAppUserType(claims.OpenID, req.UserType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户类型失败: " + err.Error()})
		return
	}

	user, err := model.GetMiniAppUserByUniqueID(claims.OpenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    user,
	})
}

func extractBearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	if len(header) > 7 && strings.EqualFold(header[0:6], "bearer") {
		return strings.TrimSpace(header[6:])
	}
	return header
}
