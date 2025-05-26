package gateway

import (
	"fmt"

	"github.com/anhao/go-easy-sms/message"
)

// 现在云短信网关常量
const (
	// NowcnEndpointURL 现在云短信 API 地址
	NowcnEndpointURL = "http://ad1200.now.net.cn:2003/sms/sendSMS"
	// NowcnSuccessCode 现在云短信 API 成功状态码
	NowcnSuccessCode = 0
)

// NowcnGateway 现在云短信网关
type NowcnGateway struct {
	*BaseGateway
}

// NewNowcnGateway 创建一个新的现在云短信网关
func NewNowcnGateway(config map[string]any) *NowcnGateway {
	return &NowcnGateway{
		BaseGateway: NewBaseGateway("nowcn", config),
	}
}

// Send 发送短信
func (g *NowcnGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 检查配置
	if g.GetConfigString("key") == "" {
		return nil, fmt.Errorf("key not found")
	}

	// 构建请求参数
	params := map[string]string{
		"mobile":   to.GetNumber(),
		"content":  msg.GetContent(),
		"userId":   g.GetConfigString("key"),
		"password": g.GetConfigString("secret"),
		"apiType":  g.GetConfigString("api_type"),
	}

	// 发送请求
	result, err := g.get(NowcnEndpointURL, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if resultCode, ok := result["code"].(float64); ok && int(resultCode) != NowcnSuccessCode {
		errorMsg := ""
		if msg, ok := result["msg"].(string); ok {
			errorMsg = msg
		}

		return result, fmt.Errorf("现在云短信发送失败: [%d] %s", int(resultCode), errorMsg)
	}

	return result, nil
}

// get 发送 GET 请求
func (g *NowcnGateway) get(endpoint string, params map[string]string) (map[string]any, error) {
	// 使用 BaseGateway 的 Get 方法发送请求
	return g.Get(endpoint, params, nil)
}
