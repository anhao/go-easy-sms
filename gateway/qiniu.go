package gateway

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/anhao/go-easy-sms/message"
)

// 七牛云短信网关常量
const (
	// QiniuEndpointTemplate 七牛云短信 API 地址模板
	QiniuEndpointTemplate = "https://%s.qiniuapi.com/%s/%s"
	// QiniuEndpointVersion 七牛云短信 API 版本
	QiniuEndpointVersion = "v1"
)

// QiniuGateway 七牛云短信网关
type QiniuGateway struct {
	*BaseGateway
}

// NewQiniuGateway 创建一个新的七牛云短信网关
func NewQiniuGateway(config map[string]any) *QiniuGateway {
	return &QiniuGateway{
		BaseGateway: NewBaseGateway("qiniu", config),
	}
}

// Send 发送短信
func (g *QiniuGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 构建请求地址
	endpoint := g.buildEndpoint("sms", "message/single")

	// 获取消息数据
	data := msg.GetData()

	// 构建请求参数
	params := map[string]any{
		"template_id": msg.GetTemplate(),
		"mobile":      to.GetNumber(),
	}

	// 添加模板参数
	if len(data) > 0 {
		params["parameters"] = data
	}

	// 构建请求头
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// 将参数转换为 JSON
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	// 生成签名
	headers["Authorization"] = g.generateSign(endpoint, "POST", string(jsonParams), headers["Content-Type"])

	// 发送请求
	result, err := g.postJSON(endpoint, params, headers)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if errorCode, ok := result["error"].(string); ok && errorCode != "" {
		errorMsg := ""
		if msg, ok := result["message"].(string); ok {
			errorMsg = msg
		}

		return result, fmt.Errorf("七牛云短信发送失败: %s", errorMsg)
	}

	return result, nil
}

// buildEndpoint 构建请求地址
func (g *QiniuGateway) buildEndpoint(typeStr, function string) string {
	return fmt.Sprintf(QiniuEndpointTemplate, typeStr, QiniuEndpointVersion, function)
}

// generateSign 生成签名
func (g *QiniuGateway) generateSign(urlStr, method, body, contentType string) string {
	// 解析 URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	host := parsedURL.Host
	path := parsedURL.Path
	query := parsedURL.RawQuery

	// 构建待签名字符串
	var toSignStr strings.Builder
	toSignStr.WriteString(method)
	toSignStr.WriteString(" ")
	toSignStr.WriteString(path)
	if query != "" {
		toSignStr.WriteString("?")
		toSignStr.WriteString(query)
	}
	toSignStr.WriteString("\nHost: ")
	toSignStr.WriteString(host)
	if contentType != "" {
		toSignStr.WriteString("\nContent-Type: ")
		toSignStr.WriteString(contentType)
	}
	toSignStr.WriteString("\n\n")
	if body != "" {
		toSignStr.WriteString(body)
	}

	// 计算 HMAC-SHA1 签名
	h := hmac.New(sha1.New, []byte(g.GetConfigString("secret_key")))
	h.Write([]byte(toSignStr.String()))
	hmacResult := h.Sum(nil)

	// Base64 URL 安全编码
	encodedSign := g.base64UrlSafeEncode(hmacResult)

	return fmt.Sprintf("Qiniu %s:%s", g.GetConfigString("access_key"), encodedSign)
}

// base64UrlSafeEncode Base64 URL 安全编码
func (g *QiniuGateway) base64UrlSafeEncode(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	encoded = strings.ReplaceAll(encoded, "+", "-")
	encoded = strings.ReplaceAll(encoded, "/", "_")
	return encoded
}

// postJSON 发送 JSON 请求
func (g *QiniuGateway) postJSON(endpoint string, params map[string]any, headers map[string]string) (map[string]any, error) {
	// 使用 BaseGateway 的 PostJSON 方法发送请求
	return g.PostJSON(endpoint, params, headers)
}
