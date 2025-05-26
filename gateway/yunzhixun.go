package gateway

import (
	"fmt"

	"github.com/anhao/go-easy-sms/message"
)

// 云之讯短信网关常量
const (
	// YunzhixunSuccessCode 云之讯短信 API 成功状态码
	YunzhixunSuccessCode = "000000"
	// YunzhixunFunctionSendSMS 云之讯短信 API 发送短信函数
	YunzhixunFunctionSendSMS = "sendsms"
	// YunzhixunFunctionBatchSendSMS 云之讯短信 API 批量发送短信函数
	YunzhixunFunctionBatchSendSMS = "sendsms_batch"
	// YunzhixunEndpointTemplate 云之讯短信 API 地址模板
	YunzhixunEndpointTemplate = "https://open.ucpaas.com/ol/%s/%s"
)

// YunzhixunGateway 云之讯短信网关
type YunzhixunGateway struct {
	*BaseGateway
}

// NewYunzhixunGateway 创建一个新的云之讯短信网关
func NewYunzhixunGateway(config map[string]any) *YunzhixunGateway {
	return &YunzhixunGateway{
		BaseGateway: NewBaseGateway("yunzhixun", config),
	}
}

// Send 发送短信
func (g *YunzhixunGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 获取消息数据
	data := msg.GetData()

	// 确定发送函数
	function := YunzhixunFunctionSendSMS
	if _, ok := data["mobiles"]; ok {
		function = YunzhixunFunctionBatchSendSMS
	}

	// 构建请求地址
	endpoint := g.buildEndpoint("sms", function)

	// 构建请求参数
	params := g.buildParams(to, msg)

	// 发送请求
	return g.execute(endpoint, params)
}

// buildEndpoint 构建请求地址
func (g *YunzhixunGateway) buildEndpoint(resource, function string) string {
	return fmt.Sprintf(YunzhixunEndpointTemplate, resource, function)
}

// buildParams 构建请求参数
func (g *YunzhixunGateway) buildParams(to *message.PhoneNumber, msg *message.Message) map[string]any {
	// 获取消息数据
	data := msg.GetData()

	// 构建参数
	params := map[string]any{
		"sid":        g.GetConfigString("sid"),
		"token":      g.GetConfigString("token"),
		"appid":      g.GetConfigString("app_id"),
		"templateid": msg.GetTemplate(),
		"uid":        "",
		"param":      "",
	}

	// 添加可选参数
	if uid, ok := data["uid"]; ok {
		params["uid"] = uid
	}

	if param, ok := data["params"]; ok {
		params["param"] = param
	}

	// 添加手机号码
	if mobiles, ok := data["mobiles"]; ok {
		params["mobile"] = mobiles
	} else {
		params["mobile"] = to.GetNumber()
	}

	return params
}

// execute 执行请求
func (g *YunzhixunGateway) execute(endpoint string, params map[string]any) (any, error) {
	// 发送请求
	result, err := g.postJSON(endpoint, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	code, ok := result["code"].(string)
	if !ok || code != YunzhixunSuccessCode {
		errMsg := "未知错误"
		if msg, ok := result["msg"].(string); ok {
			errMsg = msg
		}
		return nil, fmt.Errorf("yunzhixun gateway error: %s (code: %s)", errMsg, code)
	}

	return result, nil
}

// postJSON 发送 JSON 请求
func (g *YunzhixunGateway) postJSON(endpoint string, params map[string]any) (map[string]any, error) {
	// 使用 BaseGateway 的 PostJSON 方法发送请求
	return g.PostJSON(endpoint, params, map[string]string{
		"Content-Type": "application/json;charset=utf-8",
		"Accept":       "application/json",
	})
}
