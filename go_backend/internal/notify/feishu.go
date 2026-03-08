package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go_backend/internal/model"
)

const defaultFeishuWebhook = "https://open.feishu.cn/open-apis/bot/v2/hook/6cf4c4d5-b73a-4105-88ff-75d5f905628a"

// NotifyOrderNew 新订单通知
// isSalesOrder: true=销售员代下单, false=用户自助下单
func NotifyOrderNew(order *model.Order, orderItems []model.OrderItem, user *model.MiniAppUser, address *model.Address, isSalesOrder bool) {
	webhook, _ := model.GetSystemSetting("feishu_webhook_url")
	if webhook == "" {
		webhook = defaultFeishuWebhook
	}
	createTime := order.CreatedAt.Format("2006-01-02 15:04:05")
	productList := formatProductList(orderItems)
	paymentMethod := formatPaymentMethod(order.PaymentMethod)
	orderSource := "用户自助下单"
	if isSalesOrder {
		orderSource = "销售员代下单"
	}
	text := fmt.Sprintf("🛒 新订单通知\n——————————\n订单状态：已下单\n下单方式：%s\n订单编号：%s\n下单时间：%s\n——————————\n👤 用户信息\n用户ID：%d\n昵称：%s\n联系电话：%s\n——————————\n📦 收货信息\n收货人：%s\n地址：%s\n——————————\n📝 商品明细\n%s\n——————————\n💰 订单金额：￥%.2f\n💳 支付方式：%s\n——————————",
		orderSource, order.OrderNumber, createTime,
		user.ID, orEmpty(user.Name), orEmpty(user.Phone),
		orEmpty(address.Contact), orEmpty(address.Address),
		productList,
		order.TotalAmount, paymentMethod,
	)
	sendFeishuText(webhook, text)
}

// NotifyOrderCancelled 订单取消通知
func NotifyOrderCancelled(order *model.Order, user *model.MiniAppUser, address *model.Address, cancelReason string) {
	webhook, _ := model.GetSystemSetting("feishu_webhook_url")
	if webhook == "" {
		webhook = defaultFeishuWebhook
	}
	cancelTime := time.Now().Format("2006-01-02 15:04:05")
	paymentMethod := formatPaymentMethod(order.PaymentMethod)
	reason := cancelReason
	if reason == "" {
		reason = "未填写"
	}
	text := fmt.Sprintf("❌ 订单取消通知\n——————————\n订单状态：已取消\n订单编号：%s\n取消时间：%s\n——————————\n👤 用户信息\n用户ID：%d\n昵称：%s\n联系电话：%s\n——————————\n📦 收货信息\n收货人：%s\n地址：%s\n——————————\n💰 订单金额：￥%.2f\n💳 支付方式：%s\n——————————\n⚠️ 取消原因：%s",
		order.OrderNumber, cancelTime,
		user.ID, orEmpty(user.Name), orEmpty(user.Phone),
		orEmpty(address.Contact), orEmpty(address.Address),
		order.TotalAmount, paymentMethod,
		reason,
	)
	sendFeishuText(webhook, text)
}

// NotifyOrderDelivered 订单送达通知
func NotifyOrderDelivered(order *model.Order, orderItems []model.OrderItem, user *model.MiniAppUser, address *model.Address) {
	webhook, _ := model.GetSystemSetting("feishu_webhook_url")
	if webhook == "" {
		webhook = defaultFeishuWebhook
	}
	deliveredTime := time.Now().Format("2006-01-02 15:04:05")
	productList := formatProductList(orderItems)
	paymentMethod := formatPaymentMethod(order.PaymentMethod)
	text := fmt.Sprintf("🚚 订单送达通知\n——————————\n订单状态：已送达\n订单编号：%s\n送达时间：%s\n——————————\n👤 用户信息\n用户ID：%d\n昵称：%s\n联系电话：%s\n——————————\n📦 收货信息\n收货人：%s\n地址：%s\n——————————\n📝 商品明细\n%s\n——————————\n💰 订单金额：￥%.2f\n💳 支付方式：%s",
		order.OrderNumber, deliveredTime,
		user.ID, orEmpty(user.Name), orEmpty(user.Phone),
		orEmpty(address.Contact), orEmpty(address.Address),
		productList,
		order.TotalAmount, paymentMethod,
	)
	sendFeishuText(webhook, text)
}

// NotifyOrderPaid 订单收款通知
func NotifyOrderPaid(order *model.Order, orderItems []model.OrderItem, user *model.MiniAppUser, transactionID string) {
	webhook, _ := model.GetSystemSetting("feishu_webhook_url")
	if webhook == "" {
		webhook = defaultFeishuWebhook
	}
	paidTime := ""
	if order.PaidAt != nil {
		paidTime = order.PaidAt.Format("2006-01-02 15:04:05")
	} else {
		paidTime = time.Now().Format("2006-01-02 15:04:05")
	}
	productList := formatProductList(orderItems)
	paymentMethod := formatPaymentMethod(order.PaymentMethod)
	txID := transactionID
	if txID == "" && order.WechatTransactionID != nil && *order.WechatTransactionID != "" {
		txID = *order.WechatTransactionID
	}
	if txID == "" {
		txID = "-"
	}
	text := fmt.Sprintf("💰 订单收款通知\n——————————\n订单状态：已收款\n订单编号：%s\n支付时间：%s\n——————————\n👤 用户信息\n用户ID：%d\n昵称：%s\n——————————\n📝 商品明细\n%s\n——————————\n💵 实付金额：￥%.2f\n💳 支付方式：%s\n📊 交易号：%s",
		order.OrderNumber, paidTime,
		user.ID, orEmpty(user.Name),
		productList,
		order.TotalAmount, paymentMethod, txID,
	)
	sendFeishuText(webhook, text)
}

func formatProductList(items []model.OrderItem) string {
	if len(items) == 0 {
		return "（无明细）"
	}
	var sb strings.Builder
	for _, it := range items {
		sb.WriteString(fmt.Sprintf("• %s %s x%d ￥%.2f\n", it.ProductName, it.SpecName, it.Quantity, it.Subtotal))
	}
	return strings.TrimSuffix(sb.String(), "\n")
}

func formatPaymentMethod(pm string) string {
	switch pm {
	case "online":
		return "在线支付"
	case "cod":
		return "货到付款"
	default:
		if pm == "" {
			return "货到付款"
		}
		return pm
	}
}

func orEmpty(s string) string {
	if s != "" {
		return s
	}
	return "-"
}

// SendTestNotification 发送测试推送（用于配置校验）
// webhookURL 为空时使用系统配置或默认地址，返回错误以便 API 反馈
func SendTestNotification(webhookURL string) error {
	if webhookURL == "" {
		webhookURL, _ = model.GetSystemSetting("feishu_webhook_url")
	}
	if webhookURL == "" {
		webhookURL = defaultFeishuWebhook
	}
	text := "🔔 飞书通知测试\n——————————\n这是一条测试消息，说明飞书推送配置正确。\n——————————\n时间：" + time.Now().Format("2006-01-02 15:04:05")
	return sendFeishuTextWithError(webhookURL, text)
}

func sendFeishuTextWithError(webhookURL, text string) error {
	if webhookURL == "" {
		return nil
	}
	payload := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": text,
		},
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("飞书返回状态码 %d", resp.StatusCode)
	}
	return nil
}

func sendFeishuText(webhookURL, text string) {
	if err := sendFeishuTextWithError(webhookURL, text); err != nil {
		log.Printf("[Feishu] 发送通知失败: %v", err)
	}
}
