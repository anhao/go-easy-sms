package gateway

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/anhao/go-easy-sms/message"
)

// 百度云短信网关常量
const (
	// BaiduEndpointHost 百度云短信 API 主机地址
	BaiduEndpointHost = "smsv3.bj.baidubce.com"
	// BaiduEndpointURI 百度云短信 API 路径
	BaiduEndpointURI = "/api/v3/sendSms"
	// BaiduAuthVersion 百度云认证版本
	BaiduAuthVersion = "bce-auth-v1"
	// BaiduDefaultExpirationInSeconds 签名有效期默认1800秒
	BaiduDefaultExpirationInSeconds = 1800
	// BaiduSuccessCode 百度云短信 API 成功状态码
	BaiduSuccessCode = 1000
)

// BaiduGateway 百度云短信网关
type BaiduGateway struct {
	*BaseGateway
}

// NewBaiduGateway 创建一个新的百度云短信网关
func NewBaiduGateway(config map[string]any) *BaiduGateway {
	return &BaiduGateway{
		BaseGateway: NewBaseGateway("baidu", config),
	}
}

// Send 发送短信
func (g *BaiduGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 构建请求参数
	params := map[string]any{
		"signatureId": g.GetConfigString("invoke_id"),
		"mobile":      to.GetNumber(),
		"template":    msg.GetTemplate(),
		"contentVar":  msg.GetData(),
	}

	// 处理自定义参数
	if contentVar, ok := params["contentVar"].(map[string]any); ok {
		if custom, ok := contentVar["custom"]; ok {
			params["custom"] = custom
			delete(contentVar, "custom")
		}
		if userExtId, ok := contentVar["userExtId"]; ok {
			params["userExtId"] = userExtId
			delete(contentVar, "userExtId")
		}
	}

	// 获取当前 UTC 时间
	datetime := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// 构建请求头
	headers := map[string]string{
		"host":         BaiduEndpointHost,
		"content-type": "application/json",
		"x-bce-date":   datetime,
	}

	// 获取需要签名的头部
	signHeaders := g.getHeadersToSign(headers, []string{"host", "x-bce-date"})

	// 生成签名
	headers["Authorization"] = g.generateSign(signHeaders, datetime)

	// 发送请求
	endpoint := g.buildEndpoint()
	result, err := g.request("POST", endpoint, headers, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if code, ok := result["code"].(float64); !ok || int(code) != BaiduSuccessCode {
		errorMsg := ""
		if msg, ok := result["message"].(string); ok {
			errorMsg = msg
		}

		return result, fmt.Errorf("百度云短信发送失败: [%d] %s", int(code), errorMsg)
	}

	return result, nil
}

// buildEndpoint 构建请求地址
func (g *BaiduGateway) buildEndpoint() string {
	domain := g.GetConfigString("domain", BaiduEndpointHost)
	return fmt.Sprintf("http://%s%s", domain, BaiduEndpointURI)
}

// generateSign 生成签名
func (g *BaiduGateway) generateSign(signHeaders map[string]string, datetime string) string {
	// 获取配置
	ak := g.GetConfigString("ak")
	sk := g.GetConfigString("sk")

	// 生成 authString
	authString := fmt.Sprintf("%s/%s/%s/%d", BaiduAuthVersion, ak, datetime, BaiduDefaultExpirationInSeconds)

	// 使用 sk 和 authString 生成 signKey
	signingKey := g.hmacSha256(sk, authString)

	// 生成标准化 URI
	canonicalURI := strings.ReplaceAll(url.QueryEscape(BaiduEndpointURI), "%2F", "/")

	// 生成标准化 QueryString
	canonicalQueryString := "" // 此 API 不需要此项，返回空字符串

	// 整理 headersToSign，以 ';' 号连接
	var signedHeaderKeys []string
	for k := range signHeaders {
		signedHeaderKeys = append(signedHeaderKeys, k)
	}
	sort.Strings(signedHeaderKeys)
	signedHeaders := strings.ToLower(strings.Join(signedHeaderKeys, ";"))

	// 生成标准化 header
	canonicalHeader := g.getCanonicalHeaders(signHeaders)

	// 组成标准请求串
	canonicalRequest := fmt.Sprintf("POST\n%s\n%s\n%s", canonicalURI, canonicalQueryString, canonicalHeader)

	// 使用 signKey 和标准请求串完成签名
	signature := g.hmacSha256(signingKey, canonicalRequest)

	// 组成最终签名串
	return fmt.Sprintf("%s/%s/%s", authString, signedHeaders, signature)
}

// getCanonicalHeaders 生成标准化 HTTP 请求头串
func (g *BaiduGateway) getCanonicalHeaders(headers map[string]string) string {
	var headerStrings []string
	for name, value := range headers {
		// trim 后再 encode，之后使用 ':' 号连接起来
		headerStrings = append(headerStrings, fmt.Sprintf("%s:%s",
			url.QueryEscape(strings.ToLower(strings.TrimSpace(name))),
			url.QueryEscape(strings.TrimSpace(value))))
	}

	sort.Strings(headerStrings)

	return strings.Join(headerStrings, "\n")
}

// getHeadersToSign 根据指定的 keys 过滤应该参与签名的 header
func (g *BaiduGateway) getHeadersToSign(headers map[string]string, keys []string) map[string]string {
	result := make(map[string]string)
	for _, key := range keys {
		if value, ok := headers[key]; ok {
			result[key] = value
		}
	}
	return result
}

// hmacSha256 计算 HMAC-SHA256
func (g *BaiduGateway) hmacSha256(key, data string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// request 发送 HTTP 请求
func (g *BaiduGateway) request(_, endpoint string, headers map[string]string, params map[string]any) (map[string]any, error) {
	// 使用 BaseGateway 的 PostJSON 方法发送请求
	// 百度云 API 只使用 POST 方法
	return g.PostJSON(endpoint, params, headers)
}
