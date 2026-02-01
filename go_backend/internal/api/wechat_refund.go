package api

import (
	"context"
	"fmt"
	"math"
	"time"

	"go_backend/internal/model"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

// RequestWechatRefund 对已支付订单发起微信退款（取消时调用）
// reason 可选，为空时使用默认原因
// 返回 (refundID, error)
func RequestWechatRefund(order *model.Order, reason string) (string, error) {
	if reason == "" {
		reason = "用户取消订单"
	}
	cfg, err := getWechatPayConfig()
	if err != nil {
		return "", fmt.Errorf("微信支付配置不完整: %w", err)
	}

	mchPrivateKey, err := utils.LoadPrivateKey(cfg.PrivateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("加载商户私钥失败: %w", err)
	}

	ctx := context.Background()
	var opts []core.ClientOption
	if cfg.PublicKeyID != "" && cfg.PublicKeyPEM != "" {
		wechatPubKey, err := utils.LoadPublicKey(cfg.PublicKeyPEM)
		if err != nil {
			return "", fmt.Errorf("加载微信支付公钥失败: %w", err)
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
		return "", fmt.Errorf("初始化微信支付客户端失败: %w", err)
	}

	// 金额转分，避免浮点精度问题
	amountFen := int64(math.Round(order.TotalAmount * 100))
	if amountFen < 1 {
		amountFen = 1
	}

	// 商户退款单号：订单号_refund，保证幂等
	outRefundNo := order.OrderNumber + "_refund"

	svc := refunddomestic.RefundsApiService{Client: client}
	req := refunddomestic.CreateRequest{
		OutTradeNo:  core.String(order.OrderNumber),
		OutRefundNo: core.String(outRefundNo),
		Reason:      core.String(reason),
		Amount: &refunddomestic.AmountReq{
			Refund:   core.Int64(amountFen),
			Total:    core.Int64(amountFen),
			Currency: core.String("CNY"),
		},
	}

	// 可选：配置退款结果回调
	refundNotifyURL, _ := model.GetSystemSetting("wechat_pay_refund_notify_url")
	if refundNotifyURL != "" {
		req.NotifyUrl = core.String(refundNotifyURL)
	}

	resp, _, err := svc.Create(ctx, req)
	if err != nil {
		return "", fmt.Errorf("调用微信退款失败: %w", err)
	}

	refundID := ""
	if resp != nil && resp.RefundId != nil {
		refundID = *resp.RefundId
	}
	return refundID, nil
}

// RefundOptions 售后退款选项
type RefundOptions struct {
	RefundAmount float64 // 退款金额（元），0 表示全额
	Reason       string  // 退款原因描述
}

// RequestWechatRefundWithOptions 售后退款：支持指定金额和自定义原因
// 用于部分退款、售后补偿等场景
func RequestWechatRefundWithOptions(order *model.Order, opts RefundOptions) (string, error) {
	reason := opts.Reason
	if reason == "" {
		reason = "售后退款"
	}
	refundAmount := opts.RefundAmount
	if refundAmount <= 0 {
		refundAmount = order.TotalAmount
	}
	if refundAmount > order.TotalAmount {
		return "", fmt.Errorf("退款金额不能超过订单实付金额 %.2f 元", order.TotalAmount)
	}

	cfg, err := getWechatPayConfig()
	if err != nil {
		return "", fmt.Errorf("微信支付配置不完整: %w", err)
	}

	mchPrivateKey, err := utils.LoadPrivateKey(cfg.PrivateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("加载商户私钥失败: %w", err)
	}

	ctx := context.Background()
	var clientOpts []core.ClientOption
	if cfg.PublicKeyID != "" && cfg.PublicKeyPEM != "" {
		wechatPubKey, err := utils.LoadPublicKey(cfg.PublicKeyPEM)
		if err != nil {
			return "", fmt.Errorf("加载微信支付公钥失败: %w", err)
		}
		clientOpts = []core.ClientOption{
			option.WithWechatPayPublicKeyAuthCipher(cfg.MchID, cfg.SerialNo, mchPrivateKey, cfg.PublicKeyID, wechatPubKey),
		}
	} else {
		clientOpts = []core.ClientOption{
			option.WithWechatPayAutoAuthCipher(cfg.MchID, cfg.SerialNo, mchPrivateKey, cfg.APIv3Key),
		}
	}
	client, err := core.NewClient(ctx, clientOpts...)
	if err != nil {
		return "", fmt.Errorf("初始化微信支付客户端失败: %w", err)
	}

	totalFen := int64(math.Round(order.TotalAmount * 100))
	if totalFen < 1 {
		totalFen = 1
	}
	refundFen := int64(math.Round(refundAmount * 100))
	if refundFen < 1 {
		refundFen = 1
	}

	// 售后退款使用唯一单号，支持同一订单多次部分退款
	outRefundNo := fmt.Sprintf("%s_refund_aftersale_%d", order.OrderNumber, time.Now().UnixNano()/1e6)

	svc := refunddomestic.RefundsApiService{Client: client}
	req := refunddomestic.CreateRequest{
		OutTradeNo:  core.String(order.OrderNumber),
		OutRefundNo: core.String(outRefundNo),
		Reason:      core.String(reason),
		Amount: &refunddomestic.AmountReq{
			Refund:   core.Int64(refundFen),
			Total:    core.Int64(totalFen),
			Currency: core.String("CNY"),
		},
	}

	refundNotifyURL, _ := model.GetSystemSetting("wechat_pay_refund_notify_url")
	if refundNotifyURL != "" {
		req.NotifyUrl = core.String(refundNotifyURL)
	}

	resp, _, err := svc.Create(ctx, req)
	if err != nil {
		return "", fmt.Errorf("调用微信退款失败: %w", err)
	}

	refundID := ""
	if resp != nil && resp.RefundId != nil {
		refundID = *resp.RefundId
	}
	return refundID, nil
}
