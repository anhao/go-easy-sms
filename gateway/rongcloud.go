package gateway

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/anhao/go-easy-sms/message"
	"github.com/google/uuid"
)

// 融云短信网关常量
const (
	// RongcloudEndpointTemplate 融云短信 API 地址模板
	RongcloudEndpointTemplate = "http://api.sms.ronghub.com/%s.%s"
	// RongcloudEndpointAction 融云短信 API 动作
	RongcloudEndpointAction = "sendCode"
	// RongcloudEndpointFormat 融云短信 API 格式
	RongcloudEndpointFormat = "json"
	// RongcloudEndpointRegion 融云短信 API 区域
	RongcloudEndpointRegion = "86" // 中国区，目前只支持此国别
	// RongcloudSuccessCode 融云短信 API 成功状态码
	RongcloudSuccessCode = 200
)

// RongcloudGateway 融云短信网关
type RongcloudGateway struct {
	*BaseGateway
}

// NewRongcloudGateway 创建一个新的融云短信网关
func NewRongcloudGateway(config map[string]any) *RongcloudGateway {
	return &RongcloudGateway{
		BaseGateway: NewBaseGateway("rongcloud", config),
	}
}

// Send 发送短信
func (g *RongcloudGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 获取消息数据
	data := msg.GetData()

	// 获取动作
	action := RongcloudEndpointAction
	if actionValue, ok := data["action"].(string); ok {
		action = actionValue
		delete(data, "action")
	}

	// 构建请求地址
	endpoint := g.buildEndpoint(action)

	// 生成随机数
	nonce := uuid.New().String()

	// 获取当前时间戳
	timestamp := time.Now().Unix()

	// 构建请求头
	headers := map[string]string{
		"Nonce":     nonce,
		"App-Key":   g.GetConfigString("app_key"),
		"Timestamp": fmt.Sprintf("%d", timestamp),
		"Signature": g.generateSign(nonce, timestamp),
	}

	// 构建请求参数
	var params map[string]string
	switch action {
	case "sendCode":
		params = map[string]string{
			"mobile":     to.GetNumber(),
			"region":     RongcloudEndpointRegion,
			"templateId": msg.GetTemplate(),
		}
	case "verifyCode":
		if codeValue, ok := data["code"].(string); !ok {
			return nil, fmt.Errorf("code is not set")
		} else if sessionIDValue, ok := data["sessionId"].(string); !ok {
			return nil, fmt.Errorf("sessionId is not set")
		} else {
			params = map[string]string{
				"code":      codeValue,
				"sessionId": sessionIDValue,
			}
		}
	case "sendNotify":
		params = map[string]string{
			"mobile":     to.GetNumber(),
			"region":     RongcloudEndpointRegion,
			"templateId": msg.GetTemplate(),
		}
		// 添加其他参数
		for k, v := range data {
			params[k] = fmt.Sprintf("%v", v)
		}
	default:
		return nil, fmt.Errorf("action: %s not supported", action)
	}

	// 发送请求
	result, err := g.post(endpoint, params, headers)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if resultCode, ok := result["code"].(float64); ok && int(resultCode) != RongcloudSuccessCode {
		errorMsg := ""
		if msg, ok := result["errorMessage"].(string); ok {
			errorMsg = msg
		}

		return result, fmt.Errorf("融云短信发送失败: [%d] %s", int(resultCode), errorMsg)
	}

	return result, nil
}

// generateSign 生成签名
func (g *RongcloudGateway) generateSign(nonce string, timestamp int64) string {
	// 构建签名字符串
	signStr := fmt.Sprintf("%s%s%d", g.GetConfigString("app_secret"), nonce, timestamp)

	// 计算 SHA1 哈希
	h := sha1.New()
	h.Write([]byte(signStr))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// buildEndpoint 构建请求地址
func (g *RongcloudGateway) buildEndpoint(action string) string {
	return fmt.Sprintf(RongcloudEndpointTemplate, action, RongcloudEndpointFormat)
}

// post 发送 POST 请求
func (g *RongcloudGateway) post(endpoint string, params map[string]string, headers map[string]string) (map[string]any, error) {
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
	for k, v := range headers {
		req.Header.Set(k, v)
	}

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
