package gateway

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/anhao/go-easy-sms/message"
)

// 火山引擎短信网关常量
const (
	// VolcengineEndpointAction 火山引擎短信 API 动作
	VolcengineEndpointAction = "SendSms"
	// VolcengineEndpointVersion 火山引擎短信 API 版本
	VolcengineEndpointVersion = "2020-01-01"
	// VolcengineEndpointContentType 火山引擎短信 API 内容类型
	VolcengineEndpointContentType = "application/json; charset=utf-8"
	// VolcengineEndpointAccept 火山引擎短信 API 接受类型
	VolcengineEndpointAccept = "application/json"
	// VolcengineEndpointUserAgent 火山引擎短信 API 用户代理
	VolcengineEndpointUserAgent = "go-easy-sms"
	// VolcengineEndpointService 火山引擎短信 API 服务名称
	VolcengineEndpointService = "volcSMS"
	// VolcengineAlgorithm 火山引擎短信 API 签名算法
	VolcengineAlgorithm = "HMAC-SHA256"
	// VolcengineEndpointDefaultRegionID 火山引擎短信 API 默认区域ID
	VolcengineEndpointDefaultRegionID = "cn-north-1"
)

// VolcengineEndpoints 火山引擎短信 API 端点
var VolcengineEndpoints = map[string]string{
	"cn-north-1":     "https://sms.volcengineapi.com",
	"ap-singapore-1": "https://sms.byteplusapi.com",
}

// VolcengineGateway 火山引擎短信网关
type VolcengineGateway struct {
	*BaseGateway
}

// NewVolcengineGateway 创建一个新的火山引擎短信网关
func NewVolcengineGateway(config map[string]any) *VolcengineGateway {
	return &VolcengineGateway{
		BaseGateway: NewBaseGateway("volcengine", config),
	}
}

// Send 发送短信
func (g *VolcengineGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 获取消息数据
	data := msg.GetData()
	signName := g.getSignName(data)
	smsAccount := g.getSmsAccount(data)
	templateID := msg.GetTemplate()
	phoneNumbers := g.getPhoneNumbers(to, data)
	templateParam := g.getTemplateParam(msg, data)
	tag := g.getTag(data)

	// 构建请求参数
	queries := map[string]string{
		"Action":  VolcengineEndpointAction,
		"Version": VolcengineEndpointVersion,
	}

	// 构建请求负载
	payload := map[string]any{
		"SmsAccount":    smsAccount,
		"Sign":          signName,
		"TemplateID":    templateID,
		"TemplateParam": templateParam,
		"PhoneNumbers":  phoneNumbers,
	}

	// 如果有标签，添加到负载中
	if tag != "" {
		payload["Tag"] = tag
	}

	// 发送请求
	result, err := g.request("POST", g.getEndpoint(), queries, payload)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if result["ResponseMetadata"] != nil {
		metadata := result["ResponseMetadata"].(map[string]any)
		if metadata["Error"] != nil {
			errorData := metadata["Error"].(map[string]any)
			return nil, fmt.Errorf("volcengine gateway error: %s", errorData["Message"])
		}
	}

	return result, nil
}

// getSignName 获取签名名称
func (g *VolcengineGateway) getSignName(data map[string]any) string {
	if signName, ok := data["sign_name"]; ok {
		return signName.(string)
	}
	return g.GetConfigString("sign_name")
}

// getSmsAccount 获取短信账号
func (g *VolcengineGateway) getSmsAccount(data map[string]any) string {
	if smsAccount, ok := data["sms_account"]; ok {
		return smsAccount.(string)
	}
	return g.GetConfigString("sms_account")
}

// getPhoneNumbers 获取电话号码
func (g *VolcengineGateway) getPhoneNumbers(to *message.PhoneNumber, data map[string]any) string {
	if phoneNumbers, ok := data["phone_numbers"]; ok {
		return phoneNumbers.(string)
	}
	return to.GetNumber()
}

// getTemplateParam 获取模板参数
func (g *VolcengineGateway) getTemplateParam(msg *message.Message, data map[string]any) map[string]any {
	if templateParam, ok := data["template_param"]; ok {
		return templateParam.(map[string]any)
	}
	return msg.GetData()
}

// getTag 获取标签
func (g *VolcengineGateway) getTag(data map[string]any) string {
	if tag, ok := data["tag"]; ok {
		return tag.(string)
	}
	return ""
}

// getEndpoint 获取端点
func (g *VolcengineGateway) getEndpoint() string {
	regionID := g.GetConfigString("region_id", VolcengineEndpointDefaultRegionID)
	if endpoint, ok := VolcengineEndpoints[regionID]; ok {
		return endpoint
	}
	return VolcengineEndpoints[VolcengineEndpointDefaultRegionID]
}

// request 发送 HTTP 请求
func (g *VolcengineGateway) request(method, endpoint string, queries map[string]string, payload map[string]any) (map[string]any, error) {
	// 构建请求 URL
	requestURL := endpoint
	if len(queries) > 0 {
		values := url.Values{}
		for k, v := range queries {
			values.Add(k, v)
		}
		requestURL = fmt.Sprintf("%s?%s", endpoint, values.Encode())
	}

	// 将参数转换为 JSON 以计算签名
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 准备请求头
	headers := map[string]string{
		"Content-Type": VolcengineEndpointContentType,
		"Accept":       VolcengineEndpointAccept,
		"User-Agent":   VolcengineEndpointUserAgent,
	}

	// 添加日期头
	requestDate := time.Now().UTC().Format("20060102T150405Z")
	headers["X-Date"] = requestDate

	// 计算并添加签名
	authHeader, err := g.generateAuthHeader(method, requestURL, headers, requestDate, jsonPayload)
	if err != nil {
		return nil, err
	}
	headers["Authorization"] = authHeader

	// 使用 BaseGateway 的 PostJSON 方法发送请求
	return g.PostJSON(requestURL, payload, headers)
}

// generateAuthHeader 生成授权头
func (g *VolcengineGateway) generateAuthHeader(method, requestURL string, headers map[string]string, requestDate string, payload []byte) (string, error) {
	// 获取凭证
	accessKeyID := g.GetConfigString("access_key_id")
	accessKeySecret := g.GetConfigString("access_key_secret")
	regionID := g.GetConfigString("region_id", VolcengineEndpointDefaultRegionID)

	// 计算负载哈希
	payloadHash := g.sha256Hex(string(payload))

	// 解析 URL 获取 host 和查询参数
	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return "", err
	}

	// 添加 host 到 headers
	headers["host"] = parsedURL.Host

	// 获取规范头部
	canonicalHeaders, signedHeaders := g.getCanonicalHeadersFromMap(headers)

	// 构建规范请求
	canonicalRequest := method + "\n" +
		g.getCanonicalURI() + "\n" +
		g.getCanonicalQueryString(parsedURL.Query()) + "\n" +
		canonicalHeaders + "\n" +
		signedHeaders + "\n" +
		payloadHash

	// 构建签名字符串
	credentialScope := g.getCredentialScope(requestDate, regionID)
	stringToSign := VolcengineAlgorithm + "\n" +
		requestDate + "\n" +
		credentialScope + "\n" +
		g.sha256Hex(canonicalRequest)

	// 计算签名
	signingKey := g.getSigningKey(accessKeySecret, requestDate, regionID)
	signature := g.hmacSha256Hex(signingKey, stringToSign)

	// 构建授权头
	return fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		VolcengineAlgorithm,
		accessKeyID,
		credentialScope,
		signedHeaders,
		signature), nil
}

// getCanonicalHeadersFromMap 获取规范头部（从 map[string]string 对象）
func (g *VolcengineGateway) getCanonicalHeadersFromMap(headers map[string]string) (string, string) {
	// 转换所有键为小写
	lowercaseHeaders := make(map[string]string)
	for name, value := range headers {
		lowercaseHeaders[strings.ToLower(name)] = strings.TrimSpace(value)
	}

	// 按照键排序
	var sortedKeys []string
	for key := range lowercaseHeaders {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	// 构建规范头部
	var canonicalHeaders strings.Builder
	var signedHeaders []string

	for _, key := range sortedKeys {
		canonicalHeaders.WriteString(key)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(lowercaseHeaders[key])
		canonicalHeaders.WriteString("\n")
		signedHeaders = append(signedHeaders, key)
	}

	return canonicalHeaders.String(), strings.Join(signedHeaders, ";")
}

// getCanonicalQueryString 获取规范查询字符串
func (g *VolcengineGateway) getCanonicalQueryString(query url.Values) string {
	if len(query) == 0 {
		return ""
	}

	// 按照键排序
	var keys []string
	for key := range query {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// 构建规范查询字符串
	var result strings.Builder
	for i, key := range keys {
		if i > 0 {
			result.WriteString("&")
		}
		result.WriteString(url.QueryEscape(key))
		result.WriteString("=")
		result.WriteString(url.QueryEscape(query.Get(key)))
	}

	return result.String()
}

// getCanonicalURI 获取规范 URI
func (g *VolcengineGateway) getCanonicalURI() string {
	return "/"
}

// getCredentialScope 获取凭证范围
func (g *VolcengineGateway) getCredentialScope(requestDate, regionID string) string {
	date := requestDate[:8] // 取日期部分 YYYYMMDD
	return fmt.Sprintf("%s/%s/%s/request", date, regionID, VolcengineEndpointService)
}

// getSigningKey 获取签名密钥
func (g *VolcengineGateway) getSigningKey(secretKey, requestDate, regionID string) []byte {
	date := requestDate[:8] // 取日期部分 YYYYMMDD
	kDate := g.hmacSha256([]byte(secretKey), date)
	kRegion := g.hmacSha256(kDate, regionID)
	kService := g.hmacSha256(kRegion, VolcengineEndpointService)
	kSigning := g.hmacSha256(kService, "request")
	return kSigning
}

// sha256Hex 计算 SHA256 哈希并返回十六进制字符串
func (g *VolcengineGateway) sha256Hex(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// hmacSha256 计算 HMAC-SHA256
func (g *VolcengineGateway) hmacSha256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// hmacSha256Hex 计算 HMAC-SHA256 并返回十六进制字符串
func (g *VolcengineGateway) hmacSha256Hex(key []byte, data string) string {
	return hex.EncodeToString(g.hmacSha256(key, data))
}
