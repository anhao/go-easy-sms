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
)

// 网易云信短信网关常量
const (
	// YunxinEndpointTemplate 网易云信短信 API 地址模板
	YunxinEndpointTemplate = "https://api.netease.im/%s/%s.action"
	// YunxinEndpointAction 网易云信短信 API 默认动作
	YunxinEndpointAction = "sendCode"
	// YunxinSuccessCode 网易云信短信 API 成功状态码
	YunxinSuccessCode = 200
)

// YunxinGateway 网易云信短信网关
type YunxinGateway struct {
	*BaseGateway
}

// NewYunxinGateway 创建一个新的网易云信短信网关
func NewYunxinGateway(config map[string]any) *YunxinGateway {
	return &YunxinGateway{
		BaseGateway: NewBaseGateway("yunxin", config),
	}
}

// Send 发送短信
func (g *YunxinGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 获取消息数据
	data := msg.GetData()

	// 获取动作
	action := YunxinEndpointAction
	if actionVal, ok := data["action"]; ok {
		action = actionVal.(string)
	}

	// 构建请求地址
	endpoint := g.buildEndpoint("sms", action)

	// 根据不同的动作构建不同的参数
	var params map[string]string
	var err error

	switch action {
	case "sendCode":
		params = g.buildSendCodeParams(to, msg)
	case "verifyCode":
		params, err = g.buildVerifyCodeParams(to, msg)
		if err != nil {
			return nil, err
		}
	case "sendTemplate":
		params = g.buildTemplateParams(to, msg)
	default:
		return nil, fmt.Errorf("yunxin gateway error: action %s not supported", action)
	}

	// 构建请求头
	headers := g.buildHeaders()

	// 发送请求
	result, err := g.post(endpoint, params, headers)
	if err != nil {
		return nil, err
	}

	// 检查响应
	code, ok := result["code"].(float64)
	if !ok || int(code) != YunxinSuccessCode {
		errMsg := "未知错误"
		if msg, ok := result["msg"].(string); ok {
			errMsg = msg
		}
		return nil, fmt.Errorf("yunxin gateway error: %s (code: %v)", errMsg, code)
	}

	return result, nil
}

// buildEndpoint 构建请求地址
func (g *YunxinGateway) buildEndpoint(resource, function string) string {
	return fmt.Sprintf(YunxinEndpointTemplate, resource, strings.ToLower(function))
}

// buildHeaders 构建请求头
func (g *YunxinGateway) buildHeaders() map[string]string {
	// 获取配置
	appKey := g.GetConfigString("app_key")
	appSecret := g.GetConfigString("app_secret")

	// 生成随机数
	nonce := fmt.Sprintf("%x", time.Now().UnixNano())

	// 获取当前时间戳
	curTime := fmt.Sprintf("%d", time.Now().Unix())

	// 计算校验和
	h := sha1.New()
	h.Write([]byte(appSecret + nonce + curTime))
	checkSum := fmt.Sprintf("%x", h.Sum(nil))

	// 构建请求头
	return map[string]string{
		"AppKey":       appKey,
		"Nonce":        nonce,
		"CurTime":      curTime,
		"CheckSum":     checkSum,
		"Content-Type": "application/x-www-form-urlencoded;charset=utf-8",
	}
}

// buildSendCodeParams 构建发送验证码参数
func (g *YunxinGateway) buildSendCodeParams(to *message.PhoneNumber, msg *message.Message) map[string]string {
	// 获取消息数据
	data := msg.GetData()
	template := msg.GetTemplate()

	// 构建参数
	params := map[string]string{
		"mobile":     to.GetUniversalNumber(),
		"templateid": template,
		"codeLen":    g.GetConfigString("code_length", "4"),
		"needUp":     g.GetConfigString("need_up", "false"),
	}

	// 添加可选参数
	if code, ok := data["code"]; ok {
		params["authCode"] = fmt.Sprintf("%v", code)
	}

	if deviceID, ok := data["device_id"]; ok {
		params["deviceId"] = fmt.Sprintf("%v", deviceID)
	}

	return params
}

// buildVerifyCodeParams 构建验证验证码参数
func (g *YunxinGateway) buildVerifyCodeParams(to *message.PhoneNumber, msg *message.Message) (map[string]string, error) {
	// 获取消息数据
	data := msg.GetData()

	// 检查必要参数
	code, ok := data["code"]
	if !ok {
		return nil, fmt.Errorf("yunxin gateway error: code cannot be empty")
	}

	// 构建参数
	return map[string]string{
		"mobile": to.GetUniversalNumber(),
		"code":   fmt.Sprintf("%v", code),
	}, nil
}

// buildTemplateParams 构建模板参数
func (g *YunxinGateway) buildTemplateParams(to *message.PhoneNumber, msg *message.Message) map[string]string {
	// 获取消息数据
	data := msg.GetData()
	template := msg.GetTemplate()

	// 构建参数
	params := map[string]string{
		"templateid": template,
		"mobiles":    fmt.Sprintf("[\"%s\"]", to.GetUniversalNumber()),
		"needUp":     g.GetConfigString("need_up", "false"),
	}

	// 添加可选参数
	if templateParams, ok := data["params"]; ok {
		// 将参数转换为 JSON 字符串
		paramsJSON, err := json.Marshal(templateParams)
		if err == nil {
			params["params"] = string(paramsJSON)
		}
	}

	return params
}

// post 发送 POST 请求
func (g *YunxinGateway) post(endpoint string, params map[string]string, headers map[string]string) (map[string]any, error) {
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
