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

// MiniAppAuthMiddleware 小程序用户认证中间件
func MiniAppAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
			c.Abort()
			return
		}

		claims, err := utils.ParseMiniAppToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "登录状态已失效，请重新登录"})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("openID", claims.OpenID)
		c.Next()
	}
}

type updateMiniUserProfileRequest struct {
	Name      string  `json:"name"`
	Contact   string  `json:"contact"`
	Phone     string  `json:"phone"`
	Address   string  `json:"address"`
	StoreType string  `json:"storeType"`
	SalesCode string  `json:"salesCode"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

// UpdateMiniAppUserProfile 更新小程序用户资料，提交后自动设为零售身份并标记资料已完善
func UpdateMiniAppUserProfile(c *gin.Context) {
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

	var req updateMiniUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 验证必填字段
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "店铺名称不能为空"})
		return
	}
	if strings.TrimSpace(req.Contact) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "联系人不能为空"})
		return
	}
	if strings.TrimSpace(req.Phone) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "手机号码不能为空"})
		return
	}
	if strings.TrimSpace(req.Address) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "地址不能为空"})
		return
	}

	// 构建更新数据
	profileData := map[string]interface{}{
		"name":      strings.TrimSpace(req.Name),
		"contact":   strings.TrimSpace(req.Contact),
		"phone":     strings.TrimSpace(req.Phone),
		"address":   strings.TrimSpace(req.Address),
		"storeType": strings.TrimSpace(req.StoreType),
		"salesCode": strings.TrimSpace(req.SalesCode),
	}
	if req.Latitude != nil {
		profileData["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		profileData["longitude"] = *req.Longitude
	}

	// 更新用户资料（会自动设置为零售身份并标记资料已完善）
	if err := model.UpdateMiniAppUserProfile(claims.OpenID, profileData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户资料失败: " + err.Error()})
		return
	}

	// 返回更新后的用户信息
	user, err := model.GetMiniAppUserByUniqueID(claims.OpenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "资料更新成功",
		"data":    user,
	})
}

// UploadMiniAppUserAvatar 上传小程序用户头像
func UploadMiniAppUserAvatar(c *gin.Context) {
	// 从中间件获取用户信息
	openID, exists := c.Get("openID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少身份凭证"})
		return
	}

	// 检查是否有文件上传
	file, headers, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择要上传的图片: " + err.Error()})
		return
	}
	defer file.Close()
	
	if headers.Size > 5*1024*1024 { // 限制文件大小为5MB
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片大小不能超过5MB"})
		return
	}
	
	// 检查文件类型
	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
	}
	extension := ""
	for i := len(headers.Filename) - 1; i >= 0; i-- {
		if headers.Filename[i] == '.' {
			extension = headers.Filename[i:]
			break
		}
	}
	if !imageExtensions[strings.ToLower(extension)] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传JPG、PNG或GIF格式的图片"})
		return
	}

	// 上传图片到MinIO
	fileURL, err := utils.UploadFile("mini-user-avatar", c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "图片上传失败: " + err.Error()})
		return
	}

	// 更新用户头像
	uniqueID := openID.(string)
	if err := model.UpdateMiniAppUserAvatar(uniqueID, fileURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户头像失败: " + err.Error()})
		return
	}

	// 返回更新后的用户信息
	user, err := model.GetMiniAppUserByUniqueID(uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "头像上传成功",
		"data": gin.H{
			"avatar": fileURL,
			"user":   user,
		},
	})
}

