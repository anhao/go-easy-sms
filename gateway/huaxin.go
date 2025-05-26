package gateway

import (
	"fmt"

	"github.com/anhao/go-easy-sms/message"
)

// 华信短信网关常量
const (
	// HuaxinEndpointTemplate 华信短信 API 地址模板
	HuaxinEndpointTemplate = "http://%s/smsJson.aspx"
	// HuaxinSuccessStatus 华信短信 API 成功状态
	HuaxinSuccessStatus = "Success"
)

// HuaxinGateway 华信短信网关
type HuaxinGateway struct {
	*BaseGateway
}

// NewHuaxinGateway 创建一个新的华信短信网关
func NewHuaxinGateway(config map[string]any) *HuaxinGateway {
	return &HuaxinGateway{
		BaseGateway: NewBaseGateway("huaxin", config),
	}
}

// Send 发送短信
func (g *HuaxinGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 构建请求地址
	endpoint := g.buildEndpoint(g.GetConfigString("ip"))

	// 构建请求参数
	params := map[string]string{
		"userid":   g.GetConfigString("user_id"),
		"account":  g.GetConfigString("account"),
		"password": g.GetConfigString("password"),
		"mobile":   to.GetNumber(),
		"content":  msg.GetContent(),
		"sendTime": "",
		"action":   "send",
		"extno":    g.GetConfigString("ext_no", ""),
	}

	// 发送请求
	result, err := g.request(endpoint, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if status, ok := result["returnstatus"].(string); !ok || status != HuaxinSuccessStatus {
		errorMsg := ""
		if msg, ok := result["message"].(string); ok {
			errorMsg = msg
		}

		return result, fmt.Errorf("华信短信发送失败: %s", errorMsg)
	}

	return result, nil
}

// buildEndpoint 构建请求地址
func (g *HuaxinGateway) buildEndpoint(ip string) string {
	return fmt.Sprintf(HuaxinEndpointTemplate, ip)
}

// request 发送 HTTP 请求
func (g *HuaxinGateway) request(endpoint string, params map[string]string) (map[string]any, error) {
	// 使用 BaseGateway 的 Post 方法发送请求
	return g.Post(endpoint, params, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
}
