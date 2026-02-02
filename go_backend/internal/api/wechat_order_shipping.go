package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go_backend/internal/config"
	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
)

var (
	mpAccessToken     string
	mpAccessTokenExp  time.Time
	mpAccessTokenLock sync.RWMutex
)

// getMiniProgramAccessToken 获取小程序 access_token（带缓存，有效期7200秒）
func getMiniProgramAccessToken() (string, error) {
	mpAccessTokenLock.RLock()
	if mpAccessToken != "" && time.Now().Before(mpAccessTokenExp) {
		tok := mpAccessToken
		mpAccessTokenLock.RUnlock()
		return tok, nil
	}
	mpAccessTokenLock.RUnlock()

	mpAccessTokenLock.Lock()
	defer mpAccessTokenLock.Unlock()

	if mpAccessToken != "" && time.Now().Before(mpAccessTokenExp) {
		return mpAccessToken, nil
	}

	appID := config.Config.MiniApp.AppID
	appSecret := config.Config.MiniApp.AppSecret
	if appID == "" || appSecret == "" {
		return "", fmt.Errorf("未配置小程序 AppID 或 AppSecret")
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Errcode     int    `json:"errcode"`
		Errmsg      string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Errcode != 0 {
		return "", fmt.Errorf("获取 access_token 失败: errcode=%d errmsg=%s", result.Errcode, result.Errmsg)
	}

	mpAccessToken = result.AccessToken
	mpAccessTokenExp = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second) // 提前5分钟过期
	return mpAccessToken, nil
}

// UploadWechatShippingInfo 向微信小程序发货信息管理服务录入发货信息
// 用于「小程序购物订单」展示物流、支持资金结算
// 仅对微信支付订单（有 wechat_transaction_id）且尚未录入过的订单生效
func UploadWechatShippingInfo(orderID int) error {
	order, err := model.GetOrderByID(orderID)
	if err != nil || order == nil {
		return fmt.Errorf("订单不存在")
	}
	if order.WechatTransactionID == nil || *order.WechatTransactionID == "" {
		return nil // 非微信支付订单，跳过
	}

	user, err := model.GetMiniAppUserByID(order.UserID)
	if err != nil || user == nil {
		return fmt.Errorf("用户不存在")
	}
	openid := user.UniqueID
	if openid == "" {
		return fmt.Errorf("用户 openid 为空")
	}

	items, err := model.GetOrderItemsByOrderID(orderID)
	if err != nil || len(items) == 0 {
		return fmt.Errorf("订单明细为空")
	}

	// 构建商品描述（限120字）
	var parts []string
	for _, it := range items {
		part := it.ProductName
		if it.SpecName != "" {
			part += " " + it.SpecName
		}
		part += "×" + fmt.Sprintf("%d", it.Quantity)
		parts = append(parts, part)
	}
	itemDesc := strings.Join(parts, ";")
	if len(itemDesc) > 120 {
		itemDesc = itemDesc[:117] + "..."
	}

	token, err := getMiniProgramAccessToken()
	if err != nil {
		return err
	}

	// 同城配送（自建配送）：logistics_type=2，不需要 tracking_no、express_company
	// delivery_mode=1 统一发货
	reqBody := map[string]interface{}{
		"order_key": map[string]interface{}{
			"order_number_type": 2,
			"transaction_id":    *order.WechatTransactionID,
		},
		"logistics_type": 2, // 同城配送
		"delivery_mode":  1, // 统一发货
		"shipping_list": []map[string]interface{}{
			{"item_desc": itemDesc},
		},
		"upload_time": time.Now().Format(time.RFC3339),
		"payer":       map[string]string{"openid": openid},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/sec/order/upload_shipping_info?access_token=%s", token)
	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if result.Errcode != 0 {
		// 10060002: 支付单已完成发货，无需重复录入
		if result.Errcode == 10060002 {
			return nil
		}
		return fmt.Errorf("微信发货录入失败: errcode=%d errmsg=%s", result.Errcode, result.Errmsg)
	}
	log.Printf("[WechatShipping] 订单 %d 发货信息录入成功", orderID)
	return nil
}

// AdminUpdateOrderDetailPath 配置「小程序购物订单」跳转路径（管理员）
// POST /api/admin/wechat/order-detail-path
// Body: {"path": "pages/order/detail?id=${商品订单号}"}
// path 必须包含 "${商品订单号}"，微信会替换为 out_trade_no（订单编号）
func AdminUpdateOrderDetailPath(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "path 必填"})
		return
	}
	if !strings.Contains(req.Path, "${商品订单号}") {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "path 必须包含 ${商品订单号}"})
		return
	}
	token, err := getMiniProgramAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取 access_token 失败: " + err.Error()})
		return
	}
	bodyBytes, _ := json.Marshal(map[string]string{"path": req.Path})
	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/sec/order/update_order_detail_path?access_token=%s", token)
	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "调用微信接口失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	var result struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "解析响应失败"})
		return
	}
	if result.Errcode != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": result.Errcode, "message": "微信接口错误: " + result.Errmsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "配置成功"})
}

// AdminUploadWechatShipping 手动录入订单发货信息（管理员，用于补录或修复）
// POST /api/admin/orders/:id/upload-wechat-shipping
func AdminUploadWechatShipping(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID无效"})
		return
	}
	if err := UploadWechatShippingInfo(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "录入成功"})
}
