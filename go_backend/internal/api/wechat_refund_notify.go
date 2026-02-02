package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"go_backend/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

// refundNotifyResource 退款结果通知解密后的资源（与微信文档一致）
type refundNotifyResource struct {
	OutTradeNo   string `json:"out_trade_no"`
	OutRefundNo  string `json:"out_refund_no"`
	RefundID     string `json:"refund_id"`
	RefundStatus string `json:"refund_status"` // SUCCESS-成功, CLOSED-关闭, ABNORMAL-异常
}

// WeChatRefundNotify 微信退款结果回调
// POST /api/mini/wechat-pay/refund-notify
// 需在系统设置中配置 wechat_pay_refund_notify_url 为完整地址，如 https://域名/api/mini/wechat-pay/refund-notify
func WeChatRefundNotify(c *gin.Context) {
	cfg, err := getWechatPayConfig()
	if err != nil {
		log.Printf("[WeChatRefundNotify] 获取配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "配置错误"})
		return
	}

	mchPrivateKey, err := utils.LoadPrivateKey(cfg.PrivateKeyPEM)
	if err != nil {
		log.Printf("[WeChatRefundNotify] 加载私钥失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "配置错误"})
		return
	}

	ctx := context.Background()
	var handler *notify.Handler
	if cfg.PublicKeyID != "" && cfg.PublicKeyPEM != "" {
		wechatPubKey, err := utils.LoadPublicKey(cfg.PublicKeyPEM)
		if err != nil {
			log.Printf("[WeChatRefundNotify] 加载微信支付公钥失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "公钥配置错误"})
			return
		}
		handler = notify.NewNotifyHandler(cfg.APIv3Key, verifiers.NewSHA256WithRSAPubkeyVerifier(cfg.PublicKeyID, *wechatPubKey))
	} else {
		err = downloader.MgrInstance().RegisterDownloaderWithPrivateKey(ctx, mchPrivateKey, cfg.SerialNo, cfg.MchID, cfg.APIv3Key)
		if err != nil {
			log.Printf("[WeChatRefundNotify] 注册下载器失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "初始化失败"})
			return
		}
		certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(cfg.MchID)
		handler = notify.NewNotifyHandler(cfg.APIv3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))
	}

	// 退款通知解密后为 refund 类型，解析到 map 再提取字段
	content := make(map[string]interface{})
	_, err = handler.ParseNotifyRequest(ctx, c.Request, &content)
	if err != nil {
		log.Printf("[WeChatRefundNotify] 解析回调失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	// content 为解密后的退款对象，提取 out_trade_no、refund_id、refund_status
	var res refundNotifyResource
	if b, je := json.Marshal(content); je == nil {
		_ = json.Unmarshal(b, &res)
	}
	if res.OutTradeNo == "" {
		if v, ok := content["out_trade_no"].(string); ok {
			res.OutTradeNo = v
		}
	}
	if res.RefundStatus == "" {
		if v, ok := content["refund_status"].(string); ok {
			res.RefundStatus = v
		}
	}

	outTradeNo := strings.TrimSpace(res.OutTradeNo)
	if outTradeNo == "" {
		log.Printf("[WeChatRefundNotify] 回调中无 out_trade_no")
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
		return
	}

	order, err := model.GetOrderByOrderNumber(outTradeNo)
	if err != nil || order == nil {
		log.Printf("[WeChatRefundNotify] 订单不存在: out_trade_no=%s", outTradeNo)
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
		return
	}

	// 根据微信退款状态更新订单
	switch strings.ToUpper(res.RefundStatus) {
	case "SUCCESS":
		if err := model.MarkOrderRefundSuccess(order.ID); err != nil {
			log.Printf("[WeChatRefundNotify] 更新退款成功失败: orderID=%d err=%v", order.ID, err)
		} else {
			log.Printf("[WeChatRefundNotify] 退款成功: orderID=%d out_trade_no=%s refund_id=%s", order.ID, outTradeNo, res.RefundID)
		}
	case "CLOSED", "ABNORMAL":
		if err := model.MarkOrderRefundFailed(order.ID); err != nil {
			log.Printf("[WeChatRefundNotify] 更新退款失败状态失败: orderID=%d err=%v", order.ID, err)
		} else {
			log.Printf("[WeChatRefundNotify] 退款失败/关闭: orderID=%d out_trade_no=%s status=%s", order.ID, outTradeNo, res.RefundStatus)
		}
	default:
		log.Printf("[WeChatRefundNotify] 未知退款状态: %s out_trade_no=%s", res.RefundStatus, outTradeNo)
	}

	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}
