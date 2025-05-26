package gateway

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anhao/go-easy-sms/message"
)

// 天翼云短信网关常量
const (
	// CtyunSuccessCode 天翼云短信 API 成功状态码
	CtyunSuccessCode = "OK"
	// CtyunEndpointHost 天翼云短信 API 主机地址
	CtyunEndpointHost = "https://sms-global.ctapi.ctyun.cn"
)

// CtyunGateway 天翼云短信网关
type CtyunGateway struct {
	*BaseGateway
}

// NewCtyunGateway 创建一个新的天翼云短信网关
func NewCtyunGateway(config map[string]any) *CtyunGateway {
	return &CtyunGateway{
		BaseGateway: NewBaseGateway("ctyun", config),
	}
}

// Send 发送短信
func (g *CtyunGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	data := msg.GetData()
	endpoint := CtyunEndpointHost + "/sms/api/v1"

	// 构建请求参数
	params := map[string]any{
		"phoneNumber":   to.String(),
		"templateCode":  g.GetConfigString("template_code"),
		"templateParam": fmt.Sprintf(`{"code":"%v"}`, data["code"]),
		"signName":      g.GetConfigString("sign_name"),
		"action":        "SendSms",
	}

	// 执行请求
	result, err := g.execute(endpoint, params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// execute 执行请求
func (g *CtyunGateway) execute(url string, data map[string]any) (map[string]any, error) {
	// 生成请求 ID
	uuid := g.generateUUID()

	// 生成时间戳
	now := time.Now().UTC()
	timeStr := now.Format("20060102T150405Z")
	timeDate := timeStr[:8]

	// 将请求体转换为 JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// 计算请求体的 SHA256 哈希值
	h := sha256.New()
	h.Write(jsonData)
	body := hex.EncodeToString(h.Sum(nil))

	// 构建签名字符串
	query := ""
	strSignature := fmt.Sprintf("ctyun-eop-request-id:%s\neop-date:%s\n\n%s\n%s", uuid, timeStr, query, body)

	// 获取配置
	secretKey := g.GetConfigString("secret_key")
	accessKey := g.GetConfigString("access_key")

	// 计算签名
	kTime := g.sha256HMAC(timeStr, secretKey)
	kAk := g.sha256HMAC(accessKey, string(kTime))
	kDate := g.sha256HMAC(timeDate, string(kAk))
	signature := base64.StdEncoding.EncodeToString(g.sha256HMAC(strSignature, string(kDate)))

	// 构建请求头
	headers := map[string]string{
		"Content-Type":         "application/json",
		"ctyun-eop-request-id": uuid,
		"Eop-Authorization":    fmt.Sprintf("%s Headers=ctyun-eop-request-id;eop-date Signature=%s", accessKey, signature),
		"eop-date":             timeStr,
	}

	// 发送请求
	result, err := g.postJSON(url, data, headers)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if code, ok := result["code"].(string); !ok || code != CtyunSuccessCode {
		errorMsg := ""
		if msg, ok := result["message"].(string); ok {
			errorMsg = msg
		}

		return result, fmt.Errorf("天翼云短信发送失败: [%s] %s", code, errorMsg)
	}

	return result, nil
}

// sha256HMAC 计算 HMAC-SHA256
func (g *CtyunGateway) sha256HMAC(str, pass string) []byte {
	h := hmac.New(sha256.New, []byte(pass))
	h.Write([]byte(str))
	return h.Sum(nil)
}

// generateUUID 生成 UUID
func (g *CtyunGateway) generateUUID() string {
	// 生成格式：年月日时分秒 + 微秒 + 3位随机数
	now := time.Now()
	dateStr := now.Format("060102150405")
	microStr := fmt.Sprintf("%06d", now.Nanosecond()/1000)

	// 生成 3 位随机数
	var randomStr string
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		// 如果随机数生成失败，使用时间的纳秒部分作为备选
		randomNum := now.Nanosecond() % 1000
		randomStr = fmt.Sprintf("%03d", randomNum)
	} else {
		randomNum := int(b[0])<<8 + int(b[1])
		randomStr = fmt.Sprintf("%03d", randomNum%1000)
	}

	return dateStr + microStr + randomStr
}

// postJSON 发送 JSON 请求
func (g *CtyunGateway) postJSON(url string, data map[string]any, headers map[string]string) (map[string]any, error) {
	// 使用 BaseGateway 的 PostJSON 方法发送请求
	return g.PostJSON(url, data, headers)
}
