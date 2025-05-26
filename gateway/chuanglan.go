package gateway

import (
	"fmt"

	"github.com/anhao/go-easy-sms/message"
)

// 创蓝短信网关常量
const (
	// ChuanglanEndpointURLTemplate URL模板
	ChuanglanEndpointURLTemplate = "https://%s.253.com/msg/send/json"
	// ChuanglanIntURL 国际短信
	ChuanglanIntURL = "http://intapi.253.com/send/json"
	// ChuanglanChannelValidateCode 验证码渠道code
	ChuanglanChannelValidateCode = "smsbj1"
	// ChuanglanChannelPromotionCode 会员营销渠道code
	ChuanglanChannelPromotionCode = "smssh1"
)

// ChuanglanGateway 创蓝短信网关
type ChuanglanGateway struct {
	*BaseGateway
}

// NewChuanglanGateway 创建一个新的创蓝短信网关
func NewChuanglanGateway(config map[string]any) *ChuanglanGateway {
	return &ChuanglanGateway{
		BaseGateway: NewBaseGateway("chuanglan", config),
	}
}

// Send 发送短信
func (g *ChuanglanGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	iddCode := to.GetIDDCode()
	if iddCode == 0 {
		iddCode = 86
	}

	// 构建请求参数
	params := map[string]any{
		"account":  g.GetConfigString("account"),
		"password": g.GetConfigString("password"),
		"phone":    to.String(),
		"msg":      g.wrapChannelContent(msg.GetContent(), iddCode),
	}

	// 处理国际短信
	if iddCode != 86 {
		params["mobile"] = fmt.Sprintf("%d%s", to.GetIDDCode(), to.String())

		// 使用国际账号和密码，如果没有则使用普通账号和密码
		intelAccount := g.GetConfigString("intel_account")
		if intelAccount != "" {
			params["account"] = intelAccount
		}

		intelPassword := g.GetConfigString("intel_password")
		if intelPassword != "" {
			params["password"] = intelPassword
		}
	}

	// 发送请求
	endpoint := g.buildEndpoint(iddCode)
	result, err := g.postJSON(endpoint, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if code, ok := result["code"].(string); !ok || code != "0" {
		errorMsg := ""
		if msg, ok := result["errorMsg"].(string); ok {
			errorMsg = msg
		}

		errorCode := 0
		if code, ok := result["code"].(string); ok {
			// 尝试将 code 转换为整数
			_, _ = fmt.Sscanf(code, "%d", &errorCode)
		}

		return result, fmt.Errorf("创蓝短信发送失败: [%d] %s", errorCode, errorMsg)
	}

	return result, nil
}

// buildEndpoint 构建请求地址
func (g *ChuanglanGateway) buildEndpoint(iddCode int) string {
	channel := g.getChannel(iddCode)

	if channel == ChuanglanIntURL {
		return channel
	}

	return fmt.Sprintf(ChuanglanEndpointURLTemplate, channel)
}

// getChannel 获取通道
func (g *ChuanglanGateway) getChannel(iddCode int) string {
	if iddCode != 86 {
		return ChuanglanIntURL
	}

	channel := g.GetConfigString("channel", ChuanglanChannelValidateCode)

	if channel != ChuanglanChannelValidateCode && channel != ChuanglanChannelPromotionCode {
		// 使用默认通道
		return ChuanglanChannelValidateCode
	}

	return channel
}

// wrapChannelContent 包装通道内容
func (g *ChuanglanGateway) wrapChannelContent(content string, iddCode int) string {
	channel := g.getChannel(iddCode)

	if channel == ChuanglanChannelPromotionCode {
		sign := g.GetConfigString("sign", "")
		if sign == "" {
			// 使用默认内容，不添加签名
			return content
		}

		unsubscribe := g.GetConfigString("unsubscribe", "")
		if unsubscribe == "" {
			// 使用默认内容，不添加退订信息
			return sign + content
		}

		return sign + content + unsubscribe
	}

	return content
}

// postJSON 发送 JSON 请求
func (g *ChuanglanGateway) postJSON(url string, data map[string]any) (map[string]any, error) {
	// 使用 BaseGateway 的 PostJSON 方法发送请求
	return g.PostJSON(url, data, map[string]string{
		"Content-Type": "application/json",
	})
}
