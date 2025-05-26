package gateway

import (
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestChuanglanGatewayValidateChannel 测试创蓝短信网关验证码通道
func TestChuanglanGatewayValidateChannel(t *testing.T) {
	// 设置模拟服务器
	baseURL := "https://smsbj1.253.com/msg/send/json"

	// 激活 httpmock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 重置 httpmock 并设置成功响应
	httpmock.Reset()
	httpmock.RegisterResponder("POST", baseURL,
		httpmock.NewJsonResponderOrPanic(http.StatusOK, map[string]any{
			"code":     "0",
			"msgId":    "17041010383624511",
			"time":     "17041010383624511",
			"errorMsg": "",
		}))

	// 创建网关配置
	config := map[string]any{
		"account":  "mock-account",
		"password": "mock-password",
		"channel":  gateway.ChuanglanChannelValidateCode,
	}

	// 创建网关
	g := gateway.NewChuanglanGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("This is a test message.")

	// 创建电话号码
	phone := message.NewPhoneNumber("18188888888")

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

// TestChuanglanGatewayPromotionChannel 测试创蓝短信网关营销通道
func TestChuanglanGatewayPromotionChannel(t *testing.T) {
	// 设置模拟服务器
	baseURL := "https://smssh1.253.com/msg/send/json"

	// 激活 httpmock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 重置 httpmock 并设置成功响应
	httpmock.Reset()
	httpmock.RegisterResponder("POST", baseURL,
		httpmock.NewJsonResponderOrPanic(http.StatusOK, map[string]any{
			"code":     "0",
			"msgId":    "17041010383624514",
			"time":     "17041010383624514",
			"errorMsg": "",
		}))

	// 创建网关配置
	config := map[string]any{
		"account":     "mock-account",
		"password":    "mock-password",
		"channel":     gateway.ChuanglanChannelPromotionCode,
		"sign":        "【通讯云】",
		"unsubscribe": "回TD退订",
	}

	// 创建网关
	g := gateway.NewChuanglanGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("This is a test message.")

	// 创建电话号码
	phone := message.NewPhoneNumber("18188888888")

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
	if !ok || msgId != "17041010383624514" {
		t.Errorf("Expected msgId to be '17041010383624514', got: %v", msgId)
	}
}

// TestChuanglanGatewayError 测试创蓝短信网关错误响应
func TestChuanglanGatewayError(t *testing.T) {
	// 设置模拟服务器
	baseURL := "https://smsbj1.253.com/msg/send/json"

	// 激活 httpmock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 重置 httpmock 并设置错误响应
	httpmock.Reset()
	httpmock.RegisterResponder("POST", baseURL,
		httpmock.NewJsonResponderOrPanic(http.StatusOK, map[string]any{
			"code":     "110",
			"msgId":    "",
			"time":     "17041010383624512",
			"errorMsg": "Error Message",
		}))

	// 创建网关配置
	config := map[string]any{
		"account":  "mock-account",
		"password": "mock-password",
		"channel":  gateway.ChuanglanChannelValidateCode,
	}

	// 创建网关
	g := gateway.NewChuanglanGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("This is a test message.")

	// 创建电话号码
	phone := message.NewPhoneNumber("18188888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误信息
	expectedError := "创蓝短信发送失败: [110] Error Message"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestChuanglanGatewayInternational 测试创蓝短信网关国际短信
func TestChuanglanGatewayInternational(t *testing.T) {
	// 设置模拟服务器
	baseURL := "http://intapi.253.com/send/json"

	// 激活 httpmock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 重置 httpmock 并设置成功响应
	httpmock.Reset()
	httpmock.RegisterResponder("POST", baseURL,
		httpmock.NewJsonResponderOrPanic(http.StatusOK, map[string]any{
			"code":     "0",
			"msgId":    "17041010383624516",
			"time":     "17041010383624516",
			"errorMsg": "",
		}))

	// 创建网关配置
	config := map[string]any{
		"account":        "mock-account",
		"password":       "mock-password",
		"intel_account":  "mock-intel-account",
		"intel_password": "mock-intel-password",
	}

	// 创建网关
	g := gateway.NewChuanglanGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("This is a test message.")

	// 创建国际电话号码
	phone := message.NewPhoneNumber("18188888888", 1)

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
	if !ok || msgId != "17041010383624516" {
		t.Errorf("Expected msgId to be '17041010383624516', got: %v", msgId)
	}
}
