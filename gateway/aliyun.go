package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"

	"github.com/anhao/go-easy-sms/message"
)

// AliyunGateway 阿里云短信网关
type AliyunGateway struct {
	*BaseGateway
}

// NewAliyunGateway 创建一个新的阿里云短信网关
func NewAliyunGateway(config map[string]any) *AliyunGateway {
	return &AliyunGateway{
		BaseGateway: NewBaseGateway("aliyun", config),
	}
}

// Send 发送短信
func (g *AliyunGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	accessKeyID := g.GetConfigString("access_key_id")
	accessKeySecret := g.GetConfigString("access_key_secret")
	signName := g.GetConfigString("sign_name")

	if accessKeyID == "" || accessKeySecret == "" || signName == "" {
		return nil, errors.New("access_key_id, access_key_secret and sign_name are required")
	}

	// 获取模板ID和模板参数
	templateCode := msg.GetTemplate()
	if templateCode == "" {
		return nil, errors.New("template is required")
	}

	// 构建请求参数
	params := map[string]string{
		"AccessKeyId":      accessKeyID,
		"Action":           "SendSms",
		"Format":           "JSON",
		"RegionId":         "cn-hangzhou",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
		"PhoneNumbers":     to.GetNumber(),
		"SignName":         signName,
		"TemplateCode":     templateCode,
	}

	// 添加模板参数
	if len(msg.GetData()) > 0 {
		templateParamJSON, err := json.Marshal(msg.GetData())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal template params: %v", err)
		}
		params["TemplateParam"] = string(templateParamJSON)
	}

	// 计算签名
	signature := g.computeSignature(accessKeySecret, params)
	params["Signature"] = signature

	// 构建请求URL
	endpoint := g.GetConfigString("endpoint", "http://dysmsapi.aliyuncs.com")

	// 构建 URL 查询参数
	query := url.Values{}
	for k, v := range params {
		query.Add(k, v)
	}

	// 构建请求 URL
	requestURL := fmt.Sprintf("%s/?%s", endpoint, query.Encode())

	// 发送 GET 请求
	// 使用 BaseGateway 的 Get 方法，但传递完整的 URL 而不是分开的 endpoint 和 params
	// 这样可以确保 URL 格式完全符合阿里云 API 的要求
	result, err := g.Get(requestURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	// 检查响应状态
	if code, ok := result["Code"].(string); !ok || code != "OK" {
		message := "unknown error"
		if msg, ok := result["Message"].(string); ok {
			message = msg
		}
		return result, fmt.Errorf("aliyun gateway error: %s", message)
	}

	return result, nil
}

// computeSignature 计算签名
func (g *AliyunGateway) computeSignature(accessKeySecret string, params map[string]string) string {
	// 按照参数名称的字母顺序排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建规范化请求字符串
	var canonicalizedQueryString strings.Builder
	for _, k := range keys {
		canonicalizedQueryString.WriteString("&")
		canonicalizedQueryString.WriteString(url.QueryEscape(k))
		canonicalizedQueryString.WriteString("=")
		canonicalizedQueryString.WriteString(url.QueryEscape(params[k]))
	}

	// 构建待签名字符串
	stringToSign := "GET&%2F&" + url.QueryEscape(canonicalizedQueryString.String()[1:])

	// 计算HMAC-SHA1签名
	key := accessKeySecret + "&"
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature
}
