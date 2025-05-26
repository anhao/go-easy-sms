package gateway

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/anhao/go-easy-sms/message"
)

// YunpianGateway 云片短信网关
type YunpianGateway struct {
	*BaseGateway
}

// NewYunpianGateway 创建一个新的云片短信网关
func NewYunpianGateway(config map[string]any) *YunpianGateway {
	return &YunpianGateway{
		BaseGateway: NewBaseGateway("yunpian", config),
	}
}

// Send 发送短信
func (g *YunpianGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	apiKey := g.GetConfigString("api_key")
	if apiKey == "" {
		return nil, errors.New("api_key is required")
	}

	// 获取短信内容
	content := msg.GetContent()
	if content == "" {
		return nil, errors.New("content is required")
	}

	// 检查是否有签名
	signature := g.GetConfigString("signature")
	if signature != "" && !strings.Contains(content, signature) {
		content = signature + content
	}

	// 构建请求参数
	data := url.Values{}
	data.Add("apikey", apiKey)
	data.Add("mobile", to.GetNumber())
	data.Add("text", content)

	// 构建请求URL
	endpoint := g.GetConfigString("endpoint", "https://sms.yunpian.com")
	requestURL := endpoint + "/v2/sms/single_send.json"

	// 将 url.Values 转换为 map[string]string
	params := make(map[string]string)
	for key, values := range data {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	// 使用 BaseGateway 的 Post 方法发送请求
	result, err := g.Post(requestURL, params, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	// 检查响应状态
	if code, ok := result["code"].(float64); !ok || code != 0 {
		message := "unknown error"
		if msg, ok := result["msg"].(string); ok {
			message = msg
		}
		return result, fmt.Errorf("yunpian gateway error: %s", message)
	}

	return result, nil
}
