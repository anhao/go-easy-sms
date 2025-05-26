package gateway

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestTwilioGateway 测试 Twilio 短信网关
func TestTwilioGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", fmt.Sprintf(gateway.TwilioEndpointURL, "mock-account-sid"),
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("To") != "+8618888888888" {
				t.Errorf("Expected To to be '+8618888888888', got: %s", req.Form.Get("To"))
			}

			if req.Form.Get("From") != "mock-from" {
				t.Errorf("Expected From to be 'mock-from', got: %s", req.Form.Get("From"))
			}

			if req.Form.Get("Body") != "This is a test message." {
				t.Errorf("Expected Body to be 'This is a test message.', got: %s", req.Form.Get("Body"))
			}

			// 验证 Basic Auth
			username, password, ok := req.BasicAuth()
			if !ok {
				t.Errorf("Expected Basic Auth to be set")
			}

			if username != "mock-account-sid" {
				t.Errorf("Expected username to be 'mock-account-sid', got: %s", username)
			}

			if password != "mock-token" {
				t.Errorf("Expected password to be 'mock-token', got: %s", password)
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"status":     "queued",
				"from":       "mock-from",
				"to":         "+8618888888888",
				"body":       "This is a test message.",
				"sid":        "mock-sid",
				"error_code": nil,
			})
		})

	// 创建网关配置
	config := map[string]any{
		"account_sid": "mock-account-sid",
		"token":       "mock-token",
		"from":        "mock-from",
	}

	// 创建网关
	g := gateway.NewTwilioGateway(config)

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

	status, ok := respMap["status"].(string)
	if !ok || status != "queued" {
		t.Errorf("Expected status to be 'queued', got: %v", status)
	}

	from, ok := respMap["from"].(string)
	if !ok || from != "mock-from" {
		t.Errorf("Expected from to be 'mock-from', got: %v", from)
	}

	to, ok := respMap["to"].(string)
	if !ok || to != "+8618888888888" {
		t.Errorf("Expected to to be '+8618888888888', got: %v", to)
	}

	body, ok := respMap["body"].(string)
	if !ok || body != "This is a test message." {
		t.Errorf("Expected body to be 'This is a test message.', got: %v", body)
	}

	sid, ok := respMap["sid"].(string)
	if !ok || sid != "mock-sid" {
		t.Errorf("Expected sid to be 'mock-sid', got: %v", sid)
	}
}

// TestTwilioGatewayError 测试 Twilio 短信网关错误响应
func TestTwilioGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", fmt.Sprintf(gateway.TwilioEndpointURL, "mock-account-sid"),
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"status":     "failed",
			"from":       "mock-from",
			"to":         "+8618888888888",
			"body":       "This is a test message.",
			"sid":        "mock-sid",
			"error_code": 30001,
			"message":    "Queue overflow",
		}))

	// 创建网关配置
	config := map[string]any{
		"account_sid": "mock-account-sid",
		"token":       "mock-token",
		"from":        "mock-from",
	}

	// 创建网关
	g := gateway.NewTwilioGateway(config)

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
	expectedError := "twilio 短信发送失败: [30001] Queue overflow"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
