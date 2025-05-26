package gateway

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/anhao/go-easy-sms/message"
)

// 移动MAS模式短信网关常量
const (
	// YidongmasblackEndpointURL 移动MAS模式短信 API 地址
	YidongmasblackEndpointURL = "http://112.35.1.155:1992/sms/norsubmit"
	// YidongmasblackEndpointMethod 移动MAS模式短信 API 方法
	YidongmasblackEndpointMethod = "send"
	// YidongmasblackSuccessStatus 移动MAS模式短信 API 成功状态
	YidongmasblackSuccessStatus = "true"
)

// YidongmasblackGateway 移动MAS模式短信网关
type YidongmasblackGateway struct {
	*BaseGateway
}

// NewYidongmasblackGateway 创建一个新的移动MAS模式短信网关
func NewYidongmasblackGateway(config map[string]any) *YidongmasblackGateway {
	return &YidongmasblackGateway{
		BaseGateway: NewBaseGateway("yidongmasblack", config),
	}
}

// Send 发送短信
func (g *YidongmasblackGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 构建请求参数
	params := map[string]any{
		"ecName":    g.GetConfigString("ecName"),
		"apId":      g.GetConfigString("apId"),
		"sign":      g.GetConfigString("sign"),
		"addSerial": g.GetConfigString("addSerial"),
		"mobiles":   to.GetNumber(),
		"content":   msg.GetContent(),
	}

	// 生成内容
	content := g.GenerateContent(params)

	// 发送请求
	result, err := g.postJSON(YidongmasblackEndpointURL, content)
	if err != nil {
		return nil, err
	}

	// 检查响应
	success, ok := result["success"].(string)
	if !ok || success != YidongmasblackSuccessStatus {
		errorCode := ""
		if rspcod, ok := result["rspcod"].(string); ok {
			errorCode = rspcod
		}

		return nil, fmt.Errorf("yidongmasblack gateway error: %s (code: %s)", success, errorCode)
	}

	return result, nil
}

// GenerateContent 生成内容
func (g *YidongmasblackGateway) GenerateContent(params map[string]any) string {
	// 获取密钥
	secretKey := g.GetConfigString("secretKey")

	// 生成 MAC
	h := md5.New()
	_, _ = fmt.Fprintf(h, "%s%s%s%s%s%s%s",
		params["ecName"],
		params["apId"],
		secretKey,
		params["mobiles"],
		params["content"],
		params["sign"],
		params["addSerial"],
	)
	params["mac"] = fmt.Sprintf("%x", h.Sum(nil))

	// 将参数转换为 JSON
	jsonData, _ := json.Marshal(params)

	// 返回 Base64 编码的 JSON
	return base64.StdEncoding.EncodeToString(jsonData)
}

// postJSON 发送 JSON 请求
func (g *YidongmasblackGateway) postJSON(endpoint string, content string) (map[string]any, error) {
	// 创建请求
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	req.Header.Set("Accept", "application/json")

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
