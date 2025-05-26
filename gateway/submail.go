package gateway

import (
	"encoding/json"
	"fmt"

	"github.com/anhao/go-easy-sms/message"
)

// 赛邮云短信网关常量
const (
	// SubmailEndpointTemplate 赛邮云短信 API 地址模板
	SubmailEndpointTemplate = "https://api.mysubmail.com/%s.%s"
	// SubmailEndpointFormat 赛邮云短信 API 响应格式
	SubmailEndpointFormat = "json"
	// SubmailSuccessStatus 赛邮云短信 API 成功状态
	SubmailSuccessStatus = "success"
)

// SubmailGateway 赛邮云短信网关
type SubmailGateway struct {
	*BaseGateway
}

// NewSubmailGateway 创建一个新的赛邮云短信网关
func NewSubmailGateway(config map[string]any) *SubmailGateway {
	return &SubmailGateway{
		BaseGateway: NewBaseGateway("submail", config),
	}
}

// Send 发送短信
func (g *SubmailGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 判断是否使用内容发送
	isContent := msg.GetContent() != ""
	var endpoint string
	var params map[string]string

	if isContent {
		// 使用内容发送
		if g.inChineseMainland(to) {
			endpoint = g.buildEndpoint("sms/send")
		} else {
			endpoint = g.buildEndpoint("internationalsms/send")
		}

		params = map[string]string{
			"appid":     g.GetConfigString("app_id"),
			"signature": g.GetConfigString("app_key"),
			"content":   msg.GetContent(),
			"to":        to.GetUniversalNumber(),
		}
	} else {
		// 使用模板发送
		if g.inChineseMainland(to) {
			endpoint = g.buildEndpoint("message/xsend")
		} else {
			endpoint = g.buildEndpoint("internationalsms/xsend")
		}

		// 获取模板 ID
		data := msg.GetData()
		templateCode := msg.GetTemplate()
		project := ""

		// 优先使用模板 ID，其次使用 data 中的 project，最后使用配置中的 project
		if templateCode != "" {
			project = templateCode
		} else if projectValue, ok := data["project"]; ok {
			if projectStr, ok := projectValue.(string); ok {
				project = projectStr
			}
		} else {
			project = g.GetConfigString("project")
		}

		// 将数据转换为 JSON 字符串
		dataJSON, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		params = map[string]string{
			"appid":     g.GetConfigString("app_id"),
			"signature": g.GetConfigString("app_key"),
			"project":   project,
			"to":        to.GetUniversalNumber(),
			"vars":      string(dataJSON),
		}
	}

	// 发送请求
	result, err := g.request(endpoint, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if status, ok := result["status"].(string); !ok || status != SubmailSuccessStatus {
		errorMsg := ""
		if msg, ok := result["msg"].(string); ok {
			errorMsg = msg
		}

		errorCode := 0
		if code, ok := result["code"].(float64); ok {
			errorCode = int(code)
		}

		return result, fmt.Errorf("赛邮云短信发送失败: [%d] %s", errorCode, errorMsg)
	}

	return result, nil
}

// buildEndpoint 构建请求地址
func (g *SubmailGateway) buildEndpoint(function string) string {
	return fmt.Sprintf(SubmailEndpointTemplate, function, SubmailEndpointFormat)
}

// inChineseMainland 判断是否是中国大陆号码
func (g *SubmailGateway) inChineseMainland(to *message.PhoneNumber) bool {
	code := to.GetIDDCode()
	return code == 0 || code == 86
}

// request 发送 HTTP 请求
func (g *SubmailGateway) request(endpoint string, params map[string]string) (map[string]any, error) {
	// 使用 BaseGateway 的 Post 方法发送请求
	return g.Post(endpoint, params, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
}
