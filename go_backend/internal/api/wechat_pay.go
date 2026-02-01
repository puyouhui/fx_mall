package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

// wechatPayConfig 微信支付配置
type wechatPayConfig struct {
	MchID             string
	AppID             string
	APIv3Key          string
	SerialNo          string
	PrivateKeyPEM     string
	NotifyURL         string
	PublicKeyID       string // 微信支付公钥ID（PUB_KEY_ID_开头，新商户必填）
	PublicKeyPEM      string // 微信支付公钥（新商户需在商户平台申请后下载 pub_key.pem）
}

// getWechatPayConfig 从系统设置获取微信支付配置
func getWechatPayConfig() (*wechatPayConfig, error) {
	mchID, _ := model.GetSystemSetting("wechat_pay_mch_id")
	appID, _ := model.GetSystemSetting("wechat_pay_app_id")
	apiV3Key, _ := model.GetSystemSetting("wechat_pay_api_v3_key")
	serialNo, _ := model.GetSystemSetting("wechat_pay_serial_no")
	privateKeyPEM, _ := model.GetSystemSetting("wechat_pay_private_key")
	notifyURL, _ := model.GetSystemSetting("wechat_pay_notify_url")
	publicKeyID, _ := model.GetSystemSetting("wechat_pay_public_key_id")
	publicKeyPEM, _ := model.GetSystemSetting("wechat_pay_public_key")

	if mchID == "" || appID == "" || apiV3Key == "" || serialNo == "" || privateKeyPEM == "" || notifyURL == "" {
		return nil, fmt.Errorf("微信支付配置不完整，请在后台【系统设置-微信支付】中配置")
	}

	return &wechatPayConfig{
		MchID:         mchID,
		AppID:         appID,
		APIv3Key:      apiV3Key,
		SerialNo:      serialNo,
		PrivateKeyPEM: privateKeyPEM,
		NotifyURL:     notifyURL,
		PublicKeyID:   strings.TrimSpace(publicKeyID),
		PublicKeyPEM:  strings.TrimSpace(publicKeyPEM),
	}, nil
}

// WeChatPayPrepay 微信支付预支付接口（小程序调起支付）
// POST /mini-app/users/orders/:id/wechat-pay/prepay
func WeChatPayPrepay(c *gin.Context) {
	user, ok := getMiniUserFromContext(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供订单ID"})
		return
	}
	orderID, err := strconv.Atoi(idStr)
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单ID格式错误"})
		return
	}

	// 获取订单
	order, err := model.GetOrderByID(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取订单失败"})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
		return
	}
	if order.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权操作此订单"})
		return
	}
	if order.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单已取消，无法支付"})
		return
	}
	// 已支付（paid_at 不为空或 status 为 paid）
	if order.PaidAt != nil || order.Status == "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单已支付"})
		return
	}
	if order.TotalAmount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单金额异常，无法支付"})
		return
	}

	cfg, err := getWechatPayConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	mchPrivateKey, err := utils.LoadPrivateKey(cfg.PrivateKeyPEM)
	if err != nil {
		log.Printf("[WeChatPayPrepay] 加载私钥失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "微信支付私钥配置错误，请检查 apiclient_key.pem 格式是否正确"})
		return
	}

	ctx := c.Request.Context()
	var opts []core.ClientOption
	// 新商户需配置微信支付公钥（无可用的平台证书时使用）
	if cfg.PublicKeyID != "" && cfg.PublicKeyPEM != "" {
		wechatPubKey, err := utils.LoadPublicKey(cfg.PublicKeyPEM)
		if err != nil {
			log.Printf("[WeChatPayPrepay] 加载微信支付公钥失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "微信支付公钥配置错误，请检查 pub_key.pem 格式"})
			return
		}
		opts = []core.ClientOption{
			option.WithWechatPayPublicKeyAuthCipher(cfg.MchID, cfg.SerialNo, mchPrivateKey, cfg.PublicKeyID, wechatPubKey),
		}
	} else {
		opts = []core.ClientOption{
			option.WithWechatPayAutoAuthCipher(cfg.MchID, cfg.SerialNo, mchPrivateKey, cfg.APIv3Key),
		}
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		log.Printf("[WeChatPayPrepay] 初始化微信支付客户端失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "微信支付服务异常: " + err.Error()})
		return
	}

	// 金额转为分
	amountFen := int64(order.TotalAmount * 100)
	if amountFen < 1 {
		amountFen = 1
	}

	svc := jsapi.JsapiApiService{Client: client}
	resp, _, err := svc.PrepayWithRequestPayment(ctx, jsapi.PrepayRequest{
		Appid:       core.String(cfg.AppID),
		Mchid:       core.String(cfg.MchID),
		Description: core.String("订单支付-" + order.OrderNumber),
		OutTradeNo:  core.String(order.OrderNumber),
		NotifyUrl:   core.String(cfg.NotifyURL),
		Amount: &jsapi.Amount{
			Total: core.Int64(amountFen),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(user.UniqueID), // mini_app_users.unique_id 即 openid
		},
	})

	if err != nil {
		log.Printf("[WeChatPayPrepay] 调用微信支付预下单失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "发起支付失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"timeStamp": resp.TimeStamp,
			"nonceStr":  resp.NonceStr,
			"package":   resp.Package,
			"signType":  resp.SignType,
			"paySign":   resp.PaySign,
		},
	})
}

// WeChatPayNotify 微信支付结果回调（微信服务器调用，无需鉴权）
// POST /api/mini/wechat-pay/notify
func WeChatPayNotify(c *gin.Context) {
	cfg, err := getWechatPayConfig()
	if err != nil {
		log.Printf("[WeChatPayNotify] 获取配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "配置错误"})
		return
	}

	mchPrivateKey, err := utils.LoadPrivateKey(cfg.PrivateKeyPEM)
	if err != nil {
		log.Printf("[WeChatPayNotify] 加载私钥失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "配置错误"})
		return
	}

	ctx := context.Background()
	var handler *notify.Handler
	if cfg.PublicKeyID != "" && cfg.PublicKeyPEM != "" {
		// 使用微信支付公钥验签（新商户）
		wechatPubKey, err := utils.LoadPublicKey(cfg.PublicKeyPEM)
		if err != nil {
			log.Printf("[WeChatPayNotify] 加载微信支付公钥失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "公钥配置错误"})
			return
		}
		handler = notify.NewNotifyHandler(cfg.APIv3Key, verifiers.NewSHA256WithRSAPubkeyVerifier(cfg.PublicKeyID, *wechatPubKey))
	} else {
		// 使用平台证书验签（老商户）
		err = downloader.MgrInstance().RegisterDownloaderWithPrivateKey(ctx, mchPrivateKey, cfg.SerialNo, cfg.MchID, cfg.APIv3Key)
		if err != nil {
			log.Printf("[WeChatPayNotify] 注册下载器失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "初始化失败"})
			return
		}
		certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(cfg.MchID)
		handler = notify.NewNotifyHandler(cfg.APIv3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))
	}

	transaction := new(payments.Transaction)
	_, err = handler.ParseNotifyRequest(ctx, c.Request, transaction)
	if err != nil {
		log.Printf("[WeChatPayNotify] 解析回调失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	// 仅处理支付成功
	if transaction.TradeState == nil || *transaction.TradeState != "SUCCESS" {
		log.Printf("[WeChatPayNotify] 非支付成功状态: %v", transaction.TradeState)
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
		return
	}

	outTradeNo := ""
	if transaction.OutTradeNo != nil {
		outTradeNo = *transaction.OutTradeNo
	}
	transactionID := ""
	if transaction.TransactionId != nil {
		transactionID = *transaction.TransactionId
	}

	order, err := model.GetOrderByOrderNumber(outTradeNo)
	if err != nil || order == nil {
		log.Printf("[WeChatPayNotify] 订单不存在: out_trade_no=%s", outTradeNo)
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
		return
	}

	if order.PaidAt != nil || order.Status == "paid" {
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
		return
	}

	if err := model.MarkOrderPaidByWechatPay(order.ID, transactionID); err != nil {
		log.Printf("[WeChatPayNotify] 更新订单失败: orderID=%d, err=%v", order.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "处理失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}
