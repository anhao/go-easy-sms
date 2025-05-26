package gateway

import (
	"fmt"

	"github.com/anhao/go-easy-sms/message"
)

// 凯信通短信网关常量
const (
	// KingttoEndpointURL 凯信通短信 API 地址
	KingttoEndpointURL = "http://101.201.41.194:9999/sms.aspx"
	// KingttoEndpointMethod 凯信通短信 API 方法
	KingttoEndpointMethod = "send"
	// KingttoSuccessStatus 凯信通短信 API 成功状态
	KingttoSuccessStatus = "Success"
)

// KingttoGateway 凯信通短信网关
type KingttoGateway struct {
	*BaseGateway
}

// NewKingttoGateway 创建一个新的金坷垃短信网关
func NewKingttoGateway(config map[string]any) *KingttoGateway {
	return &KingttoGateway{
		BaseGateway: NewBaseGateway("kingtto", config),
	}
}

// Send 发送短信
func (g *KingttoGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 构建请求参数
	params := map[string]string{
		"action":   KingttoEndpointMethod,
		"userid":   g.GetConfigString("userid"),
		"account":  g.GetConfigString("account"),
		"password": g.GetConfigString("password"),
		"mobile":   to.GetNumber(),
		"content":  msg.GetContent(),
	}

	// 发送请求
	result, err := g.post(KingttoEndpointURL, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if status, ok := result["returnstatus"].(string); !ok || status != KingttoSuccessStatus {
		errorMsg := ""
		if msg, ok := result["message"].(string); ok {
			errorMsg = msg
		}

		remainPoint := 0
		if point, ok := result["remainpoint"].(string); ok {
			// 尝试将 point 转换为整数
			_, _ = fmt.Sscanf(point, "%d", &remainPoint)
		}

		return result, fmt.Errorf("金坷垃短信发送失败: %s", errorMsg)
	}

	return result, nil
}

// post 发送 POST 请求
func (g *KingttoGateway) post(endpoint string, params map[string]string) (map[string]any, error) {
	// 使用 BaseGateway 的 Post 方法发送请求
	return g.Post(endpoint, params, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
}
