package gateway

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"time"

	"github.com/anhao/go-easy-sms/message"
)

// Ue35 短信网关常量
const (
	// Ue35EndpointHost Ue35 短信 API 主机
	Ue35EndpointHost = "sms.ue35.net:8443"
	// Ue35EndpointURI Ue35 短信 API URI
	Ue35EndpointURI = "/sms/interface/sendmess.htm"
	// Ue35SuccessCode Ue35 短信 API 成功状态码
	Ue35SuccessCode = 1
)

// Ue35Gateway Ue35 短信网关
type Ue35Gateway struct {
	*BaseGateway
}

// NewUe35Gateway 创建一个新的 Ue35 短信网关
func NewUe35Gateway(config map[string]any) *Ue35Gateway {
	return &Ue35Gateway{
		BaseGateway: NewBaseGateway("ue35", config),
	}
}

// Send 发送短信
func (g *Ue35Gateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 构建请求参数
	params := map[string]string{
		"username": g.GetConfigString("username"),
		"userpwd":  g.GetConfigString("userpwd"),
		"mobiles":  to.GetNumber(),
		"content":  msg.GetContent(),
	}

	// 发送请求
	result, err := g.request(g.getEndpointURI(), params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	errorCode, ok := result["errorcode"].(float64)
	if !ok || int(errorCode) != Ue35SuccessCode {
		errMsg := "未知错误"
		if msg, ok := result["message"].(string); ok {
			errMsg = msg
		}
		return nil, fmt.Errorf("ue35 gateway error: %s (code: %v)", errMsg, errorCode)
	}

	return result, nil
}

// getEndpointURI 获取端点 URI
func (g *Ue35Gateway) getEndpointURI() string {
	return "https://" + Ue35EndpointHost + Ue35EndpointURI
}

// request 发送 HTTP 请求
func (g *Ue35Gateway) request(endpoint string, params map[string]string) (map[string]any, error) {
	// 构建 URL 查询参数
	query := url.Values{}
	for k, v := range params {
		query.Add(k, v)
	}

	// 构建请求 URL
	requestURL := fmt.Sprintf("%s?%s", endpoint, query.Encode())

	// 设置请求头
	headers := map[string]string{
		"Host":         Ue35EndpointHost,
		"Content-Type": "application/json",
		"User-Agent":   "Go EasySms Client",
	}

	// 创建上下文，设置超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(g.GetConfigFloat("timeout", 5.0))*time.Second)
	defer cancel()

	// 使用 BaseGateway 的 HTTP 客户端发送请求
	httpClient := g.GetHTTPClient()
	body, err := httpClient.Get(ctx, requestURL, nil, headers)
	if err != nil {
		return nil, err
	}

	// 尝试解析 XML 响应
	var xmlData struct {
		XMLName   xml.Name `xml:"returnsms"`
		ErrorCode int      `xml:"errorcode"`
		Message   string   `xml:"message"`
	}

	if err := xml.Unmarshal(body, &xmlData); err == nil {
		// 成功解析 XML
		return map[string]any{
			"errorcode": float64(xmlData.ErrorCode),
			"message":   xmlData.Message,
		}, nil
	}

	// 尝试解析 JSON 响应
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		// 如果 JSON 解析也失败，返回原始响应
		return map[string]any{
			"errorcode": float64(0),
			"message":   string(body),
		}, nil
	}

	return result, nil
}
