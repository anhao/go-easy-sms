package gateway

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/anhao/go-easy-sms/message"
)

// 聚合数据短信网关常量
const (
	// JuheEndpointURL 聚合数据短信 API 地址
	JuheEndpointURL = "http://v.juhe.cn/sms/send"
	// JuheEndpointFormat 聚合数据短信 API 格式
	JuheEndpointFormat = "json"
)

// JuheGateway 聚合数据短信网关
type JuheGateway struct {
	*BaseGateway
}

// NewJuheGateway 创建一个新的聚合数据短信网关
func NewJuheGateway(config map[string]any) *JuheGateway {
	return &JuheGateway{
		BaseGateway: NewBaseGateway("juhe", config),
	}
}

// Send 发送短信
func (g *JuheGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 构建请求参数
	params := map[string]string{
		"mobile":    to.GetNumber(),
		"tpl_id":    msg.GetTemplate(),
		"tpl_value": g.formatTemplateVars(msg.GetData()),
		"dtype":     JuheEndpointFormat,
		"key":       g.GetConfigString("app_key"),
	}

	// 发送请求
	result, err := g.get(JuheEndpointURL, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if errorCode, ok := result["error_code"].(float64); ok && errorCode != 0 {
		errorMsg := ""
		if reason, ok := result["reason"].(string); ok {
			errorMsg = reason
		}

		return result, fmt.Errorf("聚合数据短信发送失败: [%d] %s", int(errorCode), errorMsg)
	}

	return result, nil
}

// formatTemplateVars 格式化模板变量
func (g *JuheGateway) formatTemplateVars(vars map[string]any) string {
	formatted := make(map[string]string)

	for key, value := range vars {
		// 确保键名格式为 #key#
		formattedKey := fmt.Sprintf("#%s#", strings.Trim(key, "#"))
		formatted[formattedKey] = fmt.Sprintf("%v", value)
	}

	// 构建查询字符串
	query := url.Values{}
	for k, v := range formatted {
		query.Add(k, v)
	}

	return query.Encode()
}

// get 发送 GET 请求
func (g *JuheGateway) get(endpoint string, params map[string]string) (map[string]any, error) {
	// 使用 BaseGateway 的 Get 方法发送请求
	return g.Get(endpoint, params, nil)
}
