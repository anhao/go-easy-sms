package gateway

import (
	"encoding/base64"
	"fmt"

	"github.com/anhao/go-easy-sms/message"
)

// 螺丝帽短信网关常量
const (
	// LuosimaoEndpointTemplate 螺丝帽短信 API 地址模板
	LuosimaoEndpointTemplate = "https://%s.luosimao.com/%s/%s.%s"
	// LuosimaoEndpointVersion 螺丝帽短信 API 版本
	LuosimaoEndpointVersion = "v1"
	// LuosimaoEndpointFormat 螺丝帽短信 API 格式
	LuosimaoEndpointFormat = "json"
)

// LuosimaoGateway 螺丝帽短信网关
type LuosimaoGateway struct {
	*BaseGateway
}

// NewLuosimaoGateway 创建一个新的螺丝帽短信网关
func NewLuosimaoGateway(config map[string]any) *LuosimaoGateway {
	return &LuosimaoGateway{
		BaseGateway: NewBaseGateway("luosimao", config),
	}
}

// Send 发送短信
func (g *LuosimaoGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 构建请求地址
	endpoint := g.buildEndpoint("sms-api", "send")

	// 构建请求参数
	params := map[string]string{
		"mobile":  to.GetNumber(),
		"message": msg.GetContent(),
	}

	// 构建请求头
	headers := map[string]string{
		"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("api:key-"+g.GetConfigString("api_key"))),
	}

	// 发送请求
	result, err := g.post(endpoint, params, headers)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if errorCode, ok := result["error"].(float64); ok && errorCode != 0 {
		errorMsg := ""
		if msg, ok := result["msg"].(string); ok {
			errorMsg = msg
		}

		return result, fmt.Errorf("螺丝帽短信发送失败: [%d] %s", int(errorCode), errorMsg)
	}

	return result, nil
}

// buildEndpoint 构建请求地址
func (g *LuosimaoGateway) buildEndpoint(typeStr, function string) string {
	return fmt.Sprintf(LuosimaoEndpointTemplate, typeStr, LuosimaoEndpointVersion, function, LuosimaoEndpointFormat)
}

// post 发送 POST 请求
func (g *LuosimaoGateway) post(endpoint string, params map[string]string, headers map[string]string) (map[string]any, error) {
	// 使用 BaseGateway 的 Post 方法发送请求
	// 合并请求头
	mergedHeaders := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	for k, v := range headers {
		mergedHeaders[k] = v
	}
	return g.Post(endpoint, params, mergedHeaders)
}
