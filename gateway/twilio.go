package gateway

import (
	"encoding/base64"
	"fmt"

	"github.com/anhao/go-easy-sms/message"
)

// Twilio 短信网关常量
const (
	// TwilioEndpointURL Twilio 短信 API 地址模板
	TwilioEndpointURL = "https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json"
)

// TwilioGateway Twilio 短信网关
type TwilioGateway struct {
	*BaseGateway
	errorStatuses []string
}

// NewTwilioGateway 创建一个新的 Twilio 短信网关
func NewTwilioGateway(config map[string]any) *TwilioGateway {
	return &TwilioGateway{
		BaseGateway: NewBaseGateway("twilio", config),
		errorStatuses: []string{
			"failed",
			"undelivered",
		},
	}
}

// Send 发送短信
func (g *TwilioGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 获取账号 SID
	accountSid := g.GetConfigString("account_sid")

	// 构建请求地址
	endpoint := g.buildEndpoint(accountSid)

	// 构建请求参数
	var formattedNumber string
	if to.GetIDDCode() != 0 {
		formattedNumber = fmt.Sprintf("+%d%s", to.GetIDDCode(), to.GetNumber())
	} else {
		formattedNumber = fmt.Sprintf("+86%s", to.GetNumber())
	}

	params := map[string]string{
		"To":   formattedNumber,
		"From": g.GetConfigString("from"),
		"Body": msg.GetContent(),
	}

	// 发送请求
	result, err := g.post(endpoint, params, accountSid, g.GetConfigString("token"))
	if err != nil {
		return nil, err
	}

	// 检查响应
	if g.isErrorStatus(result) {
		errorMsg := ""
		if msg, ok := result["message"].(string); ok {
			errorMsg = msg
		}

		errorCode := 0
		if code, ok := result["error_code"].(float64); ok {
			errorCode = int(code)
		}

		return result, fmt.Errorf("twilio 短信发送失败: [%d] %s", errorCode, errorMsg)
	}

	return result, nil
}

// buildEndpoint 构建请求地址
func (g *TwilioGateway) buildEndpoint(accountSid string) string {
	return fmt.Sprintf(TwilioEndpointURL, accountSid)
}

// isErrorStatus 检查是否为错误状态
func (g *TwilioGateway) isErrorStatus(result map[string]any) bool {
	// 检查状态
	if status, ok := result["status"].(string); ok {
		for _, errorStatus := range g.errorStatuses {
			if status == errorStatus {
				return true
			}
		}
	}

	// 检查错误码
	if errorCode, ok := result["error_code"].(float64); ok && errorCode != 0 {
		return true
	}

	return false
}

// post 发送 POST 请求
func (g *TwilioGateway) post(endpoint string, params map[string]string, username, password string) (map[string]any, error) {
	// 使用 BaseGateway 的 Post 方法发送请求
	headers := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password)),
	}
	return g.Post(endpoint, params, headers)
}
