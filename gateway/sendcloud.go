package gateway

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/anhao/go-easy-sms/message"
)

// SendCloud 短信网关常量
const (
	// SendcloudEndpointTemplate SendCloud 短信 API 地址模板
	SendcloudEndpointTemplate = "http://www.sendcloud.net/smsapi/%s"
)

// SendcloudGateway SendCloud 短信网关
type SendcloudGateway struct {
	*BaseGateway
}

// NewSendcloudGateway 创建一个新的 SendCloud 短信网关
func NewSendcloudGateway(config map[string]any) *SendcloudGateway {
	return &SendcloudGateway{
		BaseGateway: NewBaseGateway("sendcloud", config),
	}
}

// Send 发送短信
func (g *SendcloudGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 构建请求参数
	params := map[string]string{
		"smsUser":    g.GetConfigString("sms_user"),
		"templateId": msg.GetTemplate(),
		"msgType":    "0",
		"phone":      to.GetZeroPrefixedNumber(),
		"vars":       g.formatTemplateVars(msg.GetData()),
	}

	// 设置国际短信类型
	if to.GetIDDCode() != 0 {
		params["msgType"] = "2"
	}

	// 添加时间戳
	if g.GetConfigBool("timestamp", false) {
		params["timestamp"] = fmt.Sprintf("%d", time.Now().UnixMilli())
	}

	// 生成签名
	params["signature"] = g.sign(params)

	// 发送请求
	result, err := g.post(fmt.Sprintf(SendcloudEndpointTemplate, "send"), params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if resultValue, ok := result["result"].(bool); !ok || !resultValue {
		errorMsg := ""
		if msg, ok := result["message"].(string); ok {
			errorMsg = msg
		}

		statusCode := 0
		if code, ok := result["statusCode"].(float64); ok {
			statusCode = int(code)
		}

		return result, fmt.Errorf("SendCloud 短信发送失败: [%d] %s", statusCode, errorMsg)
	}

	return result, nil
}

// formatTemplateVars 格式化模板变量
func (g *SendcloudGateway) formatTemplateVars(vars map[string]any) string {
	formatted := make(map[string]any)

	for key, value := range vars {
		// 确保键名格式为 %key%
		formattedKey := fmt.Sprintf("%%%s%%", strings.Trim(key, "%"))
		formatted[formattedKey] = value
	}

	// 将格式化后的变量转换为 JSON 字符串
	jsonData, err := json.Marshal(formatted)
	if err != nil {
		return "{}"
	}

	return string(jsonData)
}

// sign 生成签名
func (g *SendcloudGateway) sign(params map[string]string) string {
	// 按照键名排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建查询字符串
	var queryString strings.Builder
	for i, k := range keys {
		if i > 0 {
			queryString.WriteString("&")
		}
		queryString.WriteString(fmt.Sprintf("%s=%s", k, params[k]))
	}

	// 构建签名字符串
	smsKey := g.GetConfigString("sms_key")
	signStr := fmt.Sprintf("%s&%s&%s", smsKey, queryString.String(), smsKey)

	// 计算 MD5 哈希
	h := md5.New()
	h.Write([]byte(signStr))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// post 发送 POST 请求
func (g *SendcloudGateway) post(endpoint string, params map[string]string) (map[string]any, error) {
	// 构建表单数据
	form := url.Values{}
	for k, v := range params {
		form.Add(k, v)
	}

	// 创建请求
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// 记录关闭响应体时的错误，但不影响主流程
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
