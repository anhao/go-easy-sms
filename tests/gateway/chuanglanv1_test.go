package gateway

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestChuanglanv1GatewayNormalChannel 测试创蓝 v1 版本 API 普通通道
func TestChuanglanv1GatewayNormalChannel(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", "https://smssh1.253.com/msg/v1/send/json",
		func(req *http.Request) (*http.Response, error) {
			// 解析请求体
			var params map[string]any
			if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// 验证请求参数
			if account, ok := params["account"].(string); !ok || account != "mock-account" {
				t.Errorf("Expected account to be 'mock-account', got: %v", account)
			}

			if password, ok := params["password"].(string); !ok || password != "mock-password" {
				t.Errorf("Expected password to be 'mock-password', got: %v", password)
			}

			if phone, ok := params["phone"].(string); !ok || phone != "18888888888" {
				t.Errorf("Expected phone to be '18888888888', got: %v", phone)
			}

			if msg, ok := params["msg"].(string); !ok || msg != "This is a test message." {
				t.Errorf("Expected msg to be 'This is a test message.', got: %v", msg)
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code":     "0",
				"msgId":    "17041010383624511",
				"time":     "17041010383624511",
				"errorMsg": "",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"account":  "mock-account",
		"password": "mock-password",
		"channel":  gateway.Chuanglanv1ChannelNormalCode,
	}

	// 创建网关
	g := gateway.NewChuanglanv1Gateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("This is a test message.")

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

	code, ok := respMap["code"].(string)
	if !ok || code != "0" {
		t.Errorf("Expected code to be '0', got: %v", code)
	}

	msgId, ok := respMap["msgId"].(string)
	if !ok || msgId != "17041010383624511" {
		t.Errorf("Expected msgId to be '17041010383624511', got: %v", msgId)
	}
}

// TestChuanglanv1GatewayVariableChannel 测试创蓝 v1 版本 API 变量通道
func TestChuanglanv1GatewayVariableChannel(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", "https://smssh1.253.com/msg/variable/json",
		func(req *http.Request) (*http.Response, error) {
			// 解析请求体
			var params map[string]any
			if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// 验证请求参数
			if account, ok := params["account"].(string); !ok || account != "mock-account" {
				t.Errorf("Expected account to be 'mock-account', got: %v", account)
			}

			if password, ok := params["password"].(string); !ok || password != "mock-password" {
				t.Errorf("Expected password to be 'mock-password', got: %v", password)
			}

			if msg, ok := params["msg"].(string); !ok || msg != "mock-template" {
				t.Errorf("Expected msg to be 'mock-template', got: %v", msg)
			}

			// 验证模板参数
			paramsData, ok := params["params"].(map[string]any)
			if !ok {
				t.Errorf("Expected params to be map[string]any, got: %T", params["params"])
			}

			if code, ok := paramsData["code"].(string); !ok || code != "123456" {
				t.Errorf("Expected params.code to be '123456', got: %v", code)
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code":     "0",
				"msgId":    "17041010383624512",
				"time":     "17041010383624512",
				"errorMsg": "",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"account":  "mock-account",
		"password": "mock-password",
		"channel":  gateway.Chuanglanv1ChannelVariableCode,
	}

	// 创建网关
	g := gateway.NewChuanglanv1Gateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template").
		SetData(map[string]any{
			"code": "123456",
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

	code, ok := respMap["code"].(string)
	if !ok || code != "0" {
		t.Errorf("Expected code to be '0', got: %v", code)
	}

	msgId, ok := respMap["msgId"].(string)
	if !ok || msgId != "17041010383624512" {
		t.Errorf("Expected msgId to be '17041010383624512', got: %v", msgId)
	}
}

// TestChuanglanv1GatewayError 测试创蓝 v1 版本 API 错误响应
func TestChuanglanv1GatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", "https://smssh1.253.com/msg/v1/send/json",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"code":     "110",
			"msgId":    "",
			"time":     "17041010383624512",
			"errorMsg": "Error Message",
		}))

	// 创建网关配置
	config := map[string]any{
		"account":  "mock-account",
		"password": "mock-password",
		"channel":  gateway.Chuanglanv1ChannelNormalCode,
	}

	// 创建网关
	g := gateway.NewChuanglanv1Gateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("This is a test message.")

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误信息
	expectedError := "创蓝 v1 版本 API 短信发送失败: Error Message"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
