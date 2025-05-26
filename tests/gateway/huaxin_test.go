package gateway

import (
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestHuaxinGateway 测试华信短信网关
func TestHuaxinGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", "http://127.0.0.1/smsJson.aspx",
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("userid") != "mock-user-id" {
				t.Errorf("Expected userid to be 'mock-user-id', got: %s", req.Form.Get("userid"))
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

			if req.Form.Get("action") != "send" {
				t.Errorf("Expected action to be 'send', got: %s", req.Form.Get("action"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"returnstatus":  gateway.HuaxinSuccessStatus,
				"message":       "操作成功",
				"remainpoint":   "100",
				"taskID":        "1504080852350206",
				"successCounts": "1",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"user_id":  "mock-user-id",
		"account":  "mock-account",
		"password": "mock-password",
		"ip":       "127.0.0.1",
		"ext_no":   "",
	}

	// 创建网关
	g := gateway.NewHuaxinGateway(config)

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
	if !ok || status != gateway.HuaxinSuccessStatus {
		t.Errorf("Expected returnstatus to be '%s', got: %v", gateway.HuaxinSuccessStatus, status)
	}

	message, ok := respMap["message"].(string)
	if !ok || message != "操作成功" {
		t.Errorf("Expected message to be '操作成功', got: %v", message)
	}
}

// TestHuaxinGatewayError 测试华信短信网关错误响应
func TestHuaxinGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", "http://127.0.0.1/smsJson.aspx",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"returnstatus":  "Faild",
			"message":       "操作失败",
			"remainpoint":   "0",
			"taskID":        "0",
			"successCounts": "0",
		}))

	// 创建网关配置
	config := map[string]any{
		"user_id":  "mock-user-id",
		"account":  "mock-account",
		"password": "mock-password",
		"ip":       "127.0.0.1",
		"ext_no":   "",
	}

	// 创建网关
	g := gateway.NewHuaxinGateway(config)

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
	expectedError := "华信短信发送失败: 操作失败"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
