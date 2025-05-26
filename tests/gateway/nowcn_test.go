package gateway

import (
	"net/http"
	"strings"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestNowcnGateway 测试现在云短信网关
func TestNowcnGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("GET", `=~^http://ad1200\.now\.net\.cn:2003/sms/sendSMS`,
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			query := req.URL.Query()

			if query.Get("mobile") != "18888888888" {
				t.Errorf("Expected mobile to be '18888888888', got: %s", query.Get("mobile"))
			}

			if query.Get("content") != "This is a test message." {
				t.Errorf("Expected content to be 'This is a test message.', got: %s", query.Get("content"))
			}

			if query.Get("userId") != "mock-key" {
				t.Errorf("Expected userId to be 'mock-key', got: %s", query.Get("userId"))
			}

			if query.Get("password") != "mock-secret" {
				t.Errorf("Expected password to be 'mock-secret', got: %s", query.Get("password"))
			}

			if query.Get("apiType") != "mock-api-type" {
				t.Errorf("Expected apiType to be 'mock-api-type', got: %s", query.Get("apiType"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code": gateway.NowcnSuccessCode,
			})
		})

	// 创建网关配置
	config := map[string]any{
		"key":      "mock-key",
		"secret":   "mock-secret",
		"api_type": "mock-api-type",
	}

	// 创建网关
	g := gateway.NewNowcnGateway(config)

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

	code, ok := respMap["code"].(float64)
	if !ok || int(code) != gateway.NowcnSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.NowcnSuccessCode, code)
	}
}

// TestNowcnGatewayError 测试现在云短信网关错误响应
func TestNowcnGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("GET", `=~^http://ad1200\.now\.net\.cn:2003/sms/sendSMS`,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"code": -4,
			"msg":  "authorize failed",
		}))

	// 创建网关配置
	config := map[string]any{
		"key":      "mock-key",
		"secret":   "mock-secret",
		"api_type": "mock-api-type",
	}

	// 创建网关
	g := gateway.NewNowcnGateway(config)

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
	expectedError := "现在云短信发送失败: [-4] authorize failed"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestNowcnGatewayMissingKey 测试现在云短信网关缺少 key 配置
func TestNowcnGatewayMissingKey(t *testing.T) {
	// 创建网关配置
	config := map[string]any{
		"secret":   "mock-secret",
		"api_type": "mock-api-type",
	}

	// 创建网关
	g := gateway.NewNowcnGateway(config)

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
	expectedError := "key not found"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', got: '%s'", expectedError, err.Error())
	}
}
