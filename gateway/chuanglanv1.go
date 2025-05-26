package gateway

import (
	"fmt"

	"github.com/anhao/go-easy-sms/message"
)

// 创蓝 v1 版本 API 短信网关常量
const (
	// Chuanglanv1IntURL 创蓝 v1 版本 API 国际短信 URL
	Chuanglanv1IntURL = "http://intapi.253.com/send/json"
	// Chuanglanv1EndpointURLTemplate 创蓝 v1 版本 API URL 模板
	Chuanglanv1EndpointURLTemplate = "https://smssh1.253.com/msg/%s/json"
	// Chuanglanv1ChannelNormalCode 创蓝 v1 版本 API 普通通道
	Chuanglanv1ChannelNormalCode = "v1/send"
	// Chuanglanv1ChannelVariableCode 创蓝 v1 版本 API 变量通道
	Chuanglanv1ChannelVariableCode = "variable"
)

// Chuanglanv1Gateway 创蓝 v1 版本 API 短信网关
type Chuanglanv1Gateway struct {
	*BaseGateway
}

// NewChuanglanv1Gateway 创建一个新的创蓝 v1 版本 API 短信网关
func NewChuanglanv1Gateway(config map[string]any) *Chuanglanv1Gateway {
	return &Chuanglanv1Gateway{
		BaseGateway: NewBaseGateway("chuanglanv1", config),
	}
}

// Send 发送短信
func (g *Chuanglanv1Gateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 获取国际区号
	iddCode := to.GetIDDCode()
	if iddCode == 0 {
		iddCode = 86
	}

	// 构建请求参数
	params := map[string]any{
		"account":  g.GetConfigString("account"),
		"password": g.GetConfigString("password"),
		"report":   g.GetConfigBool("needstatus", false),
	}

	// 处理国际短信
	if iddCode != 86 {
		params["mobile"] = fmt.Sprintf("%d%s", iddCode, to.GetNumber())

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

	// 获取通道
	channel := g.getChannel(iddCode)

	// 处理不同通道的参数
	if channel == Chuanglanv1ChannelVariableCode {
		params["params"] = msg.GetData()
		params["msg"] = g.wrapChannelContent(msg.GetTemplate(), iddCode)
	} else {
		params["phone"] = to.GetNumber()
		params["msg"] = g.wrapChannelContent(msg.GetContent(), iddCode)
	}

	// 构建请求地址
	endpoint := g.buildEndpoint(iddCode)

	// 发送请求
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

		return result, fmt.Errorf("创蓝 v1 版本 API 短信发送失败: %s", errorMsg)
	}

	return result, nil
}

// buildEndpoint 构建请求地址
func (g *Chuanglanv1Gateway) buildEndpoint(iddCode int) string {
	channel := g.getChannel(iddCode)

	if channel == Chuanglanv1IntURL {
		return channel
	}

	return fmt.Sprintf(Chuanglanv1EndpointURLTemplate, channel)
}

// getChannel 获取通道
func (g *Chuanglanv1Gateway) getChannel(iddCode int) string {
	if iddCode != 86 {
		return Chuanglanv1IntURL
	}

	channel := g.GetConfigString("channel", Chuanglanv1ChannelNormalCode)

	if channel != Chuanglanv1ChannelNormalCode && channel != Chuanglanv1ChannelVariableCode {
		// 使用默认通道
		return Chuanglanv1ChannelNormalCode
	}

	return channel
}

// wrapChannelContent 包装通道内容
func (g *Chuanglanv1Gateway) wrapChannelContent(content string, _ int) string {
	return content
}

// postJSON 发送 JSON 请求
func (g *Chuanglanv1Gateway) postJSON(url string, data map[string]any) (map[string]any, error) {
	// 使用 BaseGateway 的 PostJSON 方法发送请求
	return g.PostJSON(url, data, map[string]string{
		"Content-Type": "application/json",
	})
}
