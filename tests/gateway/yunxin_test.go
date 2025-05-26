package gateway

import (
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestYunxinGateway 测试网易云信短信网关
func TestYunxinGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应 - sendCode
	httpmock.RegisterResponder("POST", "https://api.netease.im/sms/sendcode.action",
		func(req *http.Request) (*http.Response, error) {
			// 验证请求头
			if req.Header.Get("AppKey") != "mock-app-key" {
				t.Errorf("Expected AppKey to be 'mock-app-key', got: %s", req.Header.Get("AppKey"))
			}

			if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded;charset=utf-8" {
				t.Errorf("Expected Content-Type to be 'application/x-www-form-urlencoded;charset=utf-8', got: %s", req.Header.Get("Content-Type"))
			}

			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("mobile") != "18888888888" {
				t.Errorf("Expected mobile to be '18888888888', got: %s", req.Form.Get("mobile"))
			}

			if req.Form.Get("templateid") != "mock-template-id" {
				t.Errorf("Expected templateid to be 'mock-template-id', got: %s", req.Form.Get("templateid"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code": gateway.YunxinSuccessCode,
				"msg":  "发送成功",
				"obj":  "123456",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"app_key":     "mock-app-key",
		"app_secret":  "mock-app-secret",
		"code_length": "6",
		"need_up":     "true",
	}

	// 创建网关
	g := gateway.NewYunxinGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-id").
		SetData(map[string]any{
			"code":      "123456",
			"device_id": "mock-device-id",
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	resp, err := g.Send(phone, msg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 验证响应
	respMap, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("Expected response to be map[string]any, got: %T", resp)
	}

	if respMap["code"].(float64) != gateway.YunxinSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.YunxinSuccessCode, respMap["code"])
	}

	if respMap["msg"].(string) != "发送成功" {
		t.Errorf("Expected msg to be '发送成功', got: %v", respMap["msg"])
	}

	if respMap["obj"].(string) != "123456" {
		t.Errorf("Expected obj to be '123456', got: %v", respMap["obj"])
	}
}

// TestYunxinGatewayVerifyCode 测试网易云信短信网关验证验证码
func TestYunxinGatewayVerifyCode(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应 - verifyCode
	httpmock.RegisterResponder("POST", "https://api.netease.im/sms/verifycode.action",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"code": gateway.YunxinSuccessCode,
			"msg":  "验证成功",
		}))

	// 创建网关配置
	config := map[string]any{
		"app_key":    "mock-app-key",
		"app_secret": "mock-app-secret",
	}

	// 创建网关
	g := gateway.NewYunxinGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetData(map[string]any{
			"action": "verifyCode",
			"code":   "123456",
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	resp, err := g.Send(phone, msg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 验证响应
	respMap, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("Expected response to be map[string]any, got: %T", resp)
	}

	if respMap["code"].(float64) != gateway.YunxinSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.YunxinSuccessCode, respMap["code"])
	}

	if respMap["msg"].(string) != "验证成功" {
		t.Errorf("Expected msg to be '验证成功', got: %v", respMap["msg"])
	}
}

// TestYunxinGatewayVerifyCodeError 测试网易云信短信网关验证验证码错误
func TestYunxinGatewayVerifyCodeError(t *testing.T) {
	// 创建网关配置
	config := map[string]any{
		"app_key":    "mock-app-key",
		"app_secret": "mock-app-secret",
	}

	// 创建网关
	g := gateway.NewYunxinGateway(config)

	// 创建消息 - 缺少 code
	msg := message.NewMessage().
		SetData(map[string]any{
			"action": "verifyCode",
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误消息
	expectedError := "yunxin gateway error: code cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestYunxinGatewaySendTemplate 测试网易云信短信网关发送模板
func TestYunxinGatewaySendTemplate(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应 - sendTemplate
	httpmock.RegisterResponder("POST", "https://api.netease.im/sms/sendtemplate.action",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"code": gateway.YunxinSuccessCode,
			"msg":  "发送成功",
			"data": map[string]any{
				"msgid": "mock-msgid",
			},
		}))

	// 创建网关配置
	config := map[string]any{
		"app_key":    "mock-app-key",
		"app_secret": "mock-app-secret",
	}

	// 创建网关
	g := gateway.NewYunxinGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-id").
		SetData(map[string]any{
			"action": "sendTemplate",
			"params": map[string]any{
				"code": "123456",
				"time": "5分钟",
			},
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	resp, err := g.Send(phone, msg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 验证响应
	respMap, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("Expected response to be map[string]any, got: %T", resp)
	}

	if respMap["code"].(float64) != gateway.YunxinSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.YunxinSuccessCode, respMap["code"])
	}

	if respMap["msg"].(string) != "发送成功" {
		t.Errorf("Expected msg to be '发送成功', got: %v", respMap["msg"])
	}
}

// TestYunxinGatewayError 测试网易云信短信网关错误响应
func TestYunxinGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", "https://api.netease.im/sms/sendcode.action",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"code": 414,
			"msg":  "参数错误",
		}))

	// 创建网关配置
	config := map[string]any{
		"app_key":    "mock-app-key",
		"app_secret": "mock-app-secret",
	}

	// 创建网关
	g := gateway.NewYunxinGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-id")

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误消息
	expectedError := "yunxin gateway error: 参数错误 (code: 414)"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestYunxinGatewayUnsupportedAction 测试网易云信短信网关不支持的动作
func TestYunxinGatewayUnsupportedAction(t *testing.T) {
	// 创建网关配置
	config := map[string]any{
		"app_key":    "mock-app-key",
		"app_secret": "mock-app-secret",
	}

	// 创建网关
	g := gateway.NewYunxinGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetData(map[string]any{
			"action": "unsupportedAction",
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误消息
	expectedError := "yunxin gateway error: action unsupportedAction not supported"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
