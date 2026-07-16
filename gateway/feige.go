package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/anhao/go-easy-sms/message"
	"github.com/google/uuid"
)

// 飞鸽云短信网关常量
const (
	// FeigeEndpointTemplate 飞鸽云模板短信 API 地址（推荐）
	FeigeEndpointTemplate = "https://api.feige.cn/sendsms/template/send"
	// FeigeEndpointCustom 飞鸽云自定义短信 API 地址（人工审核）
	FeigeEndpointCustom = "https://api.feige.cn/sendsms/send"
	// FeigeSuccessCode 飞鸽云短信 API 成功状态码
	FeigeSuccessCode = 0
)

// feigeReservedKeys 模板变量中需要排除的保留字段
var feigeReservedKeys = map[string]struct{}{
	"sms_sign_id": {},
	"user_data":   {},
}

// FeigeGateway 飞鸽云短信网关
//
// 支持两种发送方式：
//   - 设置 template + data：使用模板短信接口（推荐）
//   - 仅设置 content：使用自定义短信接口（需人工审核）
type FeigeGateway struct {
	*BaseGateway
}

// NewFeigeGateway 创建一个新的飞鸽云短信网关
func NewFeigeGateway(config map[string]any) *FeigeGateway {
	return &FeigeGateway{
		BaseGateway: NewBaseGateway("feige", config),
	}
}

// Send 发送短信
func (g *FeigeGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	data := msg.GetData()
	template := msg.GetTemplate()

	var endpoint string
	var payload map[string]any

	if template != "" {
		// 模板短信（推荐）：模板 ID 必须为数字
		templateID, err := strconv.Atoi(template)
		if err != nil {
			return nil, fmt.Errorf("飞鸽云短信模板ID必须为数字: %s", template)
		}

		// 构建模板变量，排除保留字段
		params := make(map[string]any)
		for k, v := range data {
			if _, reserved := feigeReservedKeys[k]; reserved {
				continue
			}
			params[k] = v
		}

		endpoint = FeigeEndpointTemplate
		payload = map[string]any{
			"template_id": templateID,
			"messages": []map[string]any{
				{
					"phone_number": to.GetNumber(),
					"params":       params,
				},
			},
		}
	} else {
		// 自定义短信（人工审核）
		endpoint = FeigeEndpointCustom
		payload = map[string]any{
			"content":       msg.GetContent(),
			"phone_numbers": []string{to.GetNumber()},
		}
	}

	// 签名 ID：模板短信可选，自定义短信必填。优先取 data，其次取配置
	if signID := g.getSignID(data); signID > 0 {
		payload["sms_sign_id"] = signID
	}

	// 自定义业务数据（报告推送时原样返回）
	if userData := g.getUserData(data); userData != "" {
		payload["user_data"] = userData
	}

	// 发送请求
	result, err := g.request(endpoint, payload)
	if err != nil {
		return nil, err
	}

	// 检查响应，code 为 0 表示成功
	code, ok := result["code"].(float64)
	if !ok || int(code) != FeigeSuccessCode {
		errorMsg := ""
		if m, ok := result["msg"].(string); ok {
			errorMsg = m
		}
		return result, fmt.Errorf("飞鸽云短信发送失败: [%d] %s", int(code), errorMsg)
	}

	return result, nil
}

// getSignID 获取签名 ID（优先使用 data 中的值，其次使用配置）
func (g *FeigeGateway) getSignID(data map[string]any) int {
	if v, ok := data["sms_sign_id"]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		case string:
			if n, err := strconv.Atoi(val); err == nil {
				return n
			}
		}
	}
	return g.GetConfigInt("sms_sign_id")
}

// getUserData 获取自定义业务数据
func (g *FeigeGateway) getUserData(data map[string]any) string {
	if v, ok := data["user_data"].(string); ok {
		return v
	}
	return ""
}

// request 发送 HTTP 请求（携带签名请求头）
func (g *FeigeGateway) request(endpoint string, payload map[string]any) (map[string]any, error) {
	apiKey := g.GetConfigString("api_key")
	secret := g.GetConfigString("api_secret")
	timestamp := time.Now().Unix()
	nonce := uuid.New().String()

	headers := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
		"X-Api-Key":    apiKey,
		"X-Timestamp":  strconv.FormatInt(timestamp, 10),
		"X-Nonce":      nonce,
		"X-Sign":       g.generateSign(apiKey, secret, timestamp, nonce),
	}

	return g.PostJSON(endpoint, payload, headers)
}

// generateSign 生成签名：SHA256(apiKey + secret + timestamp + nonce)，结果为小写十六进制
func (g *FeigeGateway) generateSign(apiKey, secret string, timestamp int64, nonce string) string {
	raw := fmt.Sprintf("%s%s%d%s", apiKey, secret, timestamp, nonce)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
