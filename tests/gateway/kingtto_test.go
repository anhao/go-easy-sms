package gateway

import (
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestKingttoGateway 测试金坷垃短信网关
func TestKingttoGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", gateway.KingttoEndpointURL,
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("action") != gateway.KingttoEndpointMethod {
				t.Errorf("Expected action to be '%s', got: %s", gateway.KingttoEndpointMethod, req.Form.Get("action"))
			}

			if req.Form.Get("userid") != "mock-userid" {
				t.Errorf("Expected userid to be 'mock-userid', got: %s", req.Form.Get("userid"))
			}

			if req.Form.Get("account") != "mock-account" {
				t.Errorf("Expected account to be 'mock-account', got: %s", req.Form.Get("account"))
			}

			if req.Form.Get("password") != "mock-password" {
				t.Errorf("Expected password to be 'mock-password', got: %s", req.Form.Get("password"))
			}

			if req.Form.Get("mobile") != "18888888888" {
				t.Errorf("Expected mobile to be '18888888888', got: %s", req.Form.Get("mobile"))
			}

			if req.Form.Get("content") != "This is a test message." {
				t.Errorf("Expected content to be 'This is a test message.', got: %s", req.Form.Get("content"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"returnstatus":  gateway.KingttoSuccessStatus,
				"message":       "ok",
				"remainpoint":   "56832",
				"taskID":        "106470408",
				"successCounts": "1",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"userid":   "mock-userid",
		"account":  "mock-account",
		"password": "mock-password",
	}

	// 创建网关
	g := gateway.NewKingttoGateway(config)

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

	status, ok := respMap["returnstatus"].(string)
	if !ok || status != gateway.KingttoSuccessStatus {
		t.Errorf("Expected returnstatus to be '%s', got: %v", gateway.KingttoSuccessStatus, status)
	}

	respMsg, ok := respMap["message"].(string)
	if !ok || respMsg != "ok" {
		t.Errorf("Expected message to be 'ok', got: %v", respMsg)
	}

	remainpoint, ok := respMap["remainpoint"].(string)
	if !ok || remainpoint != "56832" {
		t.Errorf("Expected remainpoint to be '56832', got: %v", remainpoint)
	}
}

// TestKingttoGatewayError 测试金坷垃短信网关错误响应
func TestKingttoGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", gateway.KingttoEndpointURL,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"returnstatus":  "Faild",
			"message":       "mock-error-message",
			"remainpoint":   "0",
			"taskID":        "0",
			"successCounts": "0",
		}))

	// 创建网关配置
	config := map[string]any{
		"userid":   "mock-userid",
		"account":  "mock-account",
		"password": "mock-password",
	}

	// 创建网关
	g := gateway.NewKingttoGateway(config)

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
	expectedError := "金坷垃短信发送失败: mock-error-message"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
