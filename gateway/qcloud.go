package gateway

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/anhao/go-easy-sms/message"
)

// QcloudGateway 腾讯云短信网关
// 参考文档: https://cloud.tencent.com/document/api/382/55981
type QcloudGateway struct {
	*BaseGateway
}

// 腾讯云短信 API 常量
const (
	QcloudEndpointURL     = "https://sms.tencentcloudapi.com"
	QcloudEndpointService = "sms"
	QcloudEndpointMethod  = "SendSms"
	QcloudEndpointVersion = "2021-01-11"
	QcloudEndpointRegion  = "ap-guangzhou"
	QcloudEndpointFormat  = "json"
)

// NewQcloudGateway 创建一个新的腾讯云短信网关
func NewQcloudGateway(config map[string]any) *QcloudGateway {
	return &QcloudGateway{
		BaseGateway: NewBaseGateway("qcloud", config),
	}
}

// Send 发送短信
func (g *QcloudGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 获取消息数据
	data := msg.GetData()

	// 获取签名，优先使用消息中的签名，如果没有则使用配置中的签名
	signName, ok := data["sign_name"].(string)
	if !ok || signName == "" {
		signName = g.GetConfigString("sign_name", "")
	}

	// 删除 sign_name 字段，避免作为模板参数发送
	delete(data, "sign_name")

	// 处理电话号码
	phone := to.String()
	if to.GetIDDCode() != 0 {
		phone = to.GetUniversalNumber()
	}

	// 构建请求参数
	params := map[string]any{
		"PhoneNumberSet": []string{phone},
		"SmsSdkAppId":    g.GetConfigString("sdk_app_id"),
		"SignName":       signName,
		"TemplateId":     msg.GetTemplate(),
		"TemplateParamSet": func() []string {
			// 将模板参数转换为字符串数组
			var values []string
			for _, v := range data {
				values = append(values, fmt.Sprintf("%v", v))
			}
			return values
		}(),
	}

	// 获取当前时间戳
	timestamp := time.Now().Unix()

	// 获取 endpoint
	endpoint := g.GetConfigString("endpoint", QcloudEndpointURL)
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	host := parsedURL.Host

	// 构建请求头
	headers := map[string]string{
		"Authorization":  g.generateSign(params, timestamp),
		"Host":           host,
		"Content-Type":   "application/json; charset=utf-8",
		"X-TC-Action":    QcloudEndpointMethod,
		"X-TC-Region":    g.GetConfigString("region", QcloudEndpointRegion),
		"X-TC-Timestamp": strconv.FormatInt(timestamp, 10),
		"X-TC-Version":   QcloudEndpointVersion,
	}

	// 发送请求
	result, err := g.PostJSON(endpoint, params, headers)
	if err != nil {
		return nil, err
	}

	// 检查错误
	if result != nil {
		if response, ok := result["Response"].(map[string]any); ok {
			// 检查错误信息
			if errorInfo, ok := response["Error"].(map[string]any); ok {
				code, _ := errorInfo["Code"].(string)
				message, _ := errorInfo["Message"].(string)
				return result, fmt.Errorf("腾讯云短信发送失败: [%s] %s", code, message)
			}

			// 检查发送状态
			if statusSet, ok := response["SendStatusSet"].([]any); ok {
				for _, status := range statusSet {
					if statusMap, ok := status.(map[string]any); ok {
						code, _ := statusMap["Code"].(string)
						if code != "Ok" {
							message, _ := statusMap["Message"].(string)
							return result, fmt.Errorf("腾讯云短信发送失败: [%s] %s", code, message)
						}
					}
				}
			}
		}
	}

	return result, nil
}

// generateSign 生成签名
func (g *QcloudGateway) generateSign(params map[string]any, timestamp int64) string {
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	secretKey := g.GetConfigString("secret_key")
	secretId := g.GetConfigString("secret_id")
	endpoint := g.GetConfigString("endpoint", QcloudEndpointURL)
	parsedURL, _ := url.Parse(endpoint)
	host := parsedURL.Host

	// 将参数转换为 JSON
	jsonParams, _ := json.Marshal(params)

	// 构建规范请求串
	canonicalRequest := "POST\n" +
		"/\n" +
		"\n" +
		"content-type:application/json; charset=utf-8\n" +
		"host:" + host + "\n\n" +
		"content-type;host\n" +
		g.sha256Hex(string(jsonParams))

	// 构建待签名字符串
	stringToSign := "TC3-HMAC-SHA256\n" +
		strconv.FormatInt(timestamp, 10) + "\n" +
		date + "/" + QcloudEndpointService + "/tc3_request\n" +
		g.sha256Hex(canonicalRequest)

	// 计算签名
	secretDate := g.hmacSha256("TC3"+secretKey, date)
	secretService := g.hmacSha256(secretDate, QcloudEndpointService)
	secretSigning := g.hmacSha256(secretService, "tc3_request")
	signature := hex.EncodeToString([]byte(g.hmacSha256(secretSigning, stringToSign)))

	// 构建授权字符串
	return "TC3-HMAC-SHA256" +
		" Credential=" + secretId + "/" + date + "/" + QcloudEndpointService + "/tc3_request" +
		", SignedHeaders=content-type;host, Signature=" + signature
}

// sha256Hex 计算字符串的 SHA256 哈希值并返回十六进制字符串
func (g *QcloudGateway) sha256Hex(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// hmacSha256 计算 HMAC-SHA256
func (g *QcloudGateway) hmacSha256(key, data string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
