package gateway

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestRongcloudGatewaySendCode 测试融云短信网关发送验证码
func TestRongcloudGatewaySendCode(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", fmt.Sprintf(gateway.RongcloudEndpointTemplate, gateway.RongcloudEndpointAction, gateway.RongcloudEndpointFormat),
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("mobile") != "18888888888" {
				t.Errorf("Expected mobile to be '18888888888', got: %s", req.Form.Get("mobile"))
			}

			if req.Form.Get("region") != gateway.RongcloudEndpointRegion {
				t.Errorf("Expected region to be '%s', got: %s", gateway.RongcloudEndpointRegion, req.Form.Get("region"))
			}

			if req.Form.Get("templateId") != "mock-tpl-id" {
				t.Errorf("Expected templateId to be 'mock-tpl-id', got: %s", req.Form.Get("templateId"))
			}

			// 验证请求头
			if req.Header.Get("App-Key") != "mock-app-key" {
				t.Errorf("Expected App-Key to be 'mock-app-key', got: %s", req.Header.Get("App-Key"))
			}

			nonce := req.Header.Get("Nonce")
			if nonce == "" {
				t.Errorf("Expected Nonce to be set")
			}

			timestamp := req.Header.Get("Timestamp")
			if timestamp == "" {
				t.Errorf("Expected Timestamp to be set")
			}

			signature := req.Header.Get("Signature")
			if signature == "" {
				t.Errorf("Expected Signature to be set")
			}

			// 验证签名
			signStr := fmt.Sprintf("%s%s%s", "mock-app-secret", nonce, timestamp)
			h := sha1.New()
			h.Write([]byte(signStr))
			expectedSignature := fmt.Sprintf("%x", h.Sum(nil))

			if signature != expectedSignature {
				t.Errorf("Expected Signature to be '%s', got: %s", expectedSignature, signature)
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code":      gateway.RongcloudSuccessCode,
				"sessionId": "mock-session-id",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"app_key":    "mock-app-key",
		"app_secret": "mock-app-secret",
	}

	// 创建网关
	g := gateway.NewRongcloudGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-tpl-id")

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
	if !ok || int(code) != gateway.RongcloudSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.RongcloudSuccessCode, code)
	}

	sessionID, ok := respMap["sessionId"].(string)
	if !ok || sessionID != "mock-session-id" {
		t.Errorf("Expected sessionId to be 'mock-session-id', got: %v", sessionID)
	}
}

// TestRongcloudGatewayVerifyCode 测试融云短信网关验证验证码
func TestRongcloudGatewayVerifyCode(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", fmt.Sprintf(gateway.RongcloudEndpointTemplate, "verifyCode", gateway.RongcloudEndpointFormat),
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("code") != "1234" {
				t.Errorf("Expected code to be '1234', got: %s", req.Form.Get("code"))
			}

			if req.Form.Get("sessionId") != "mock-session-id" {
				t.Errorf("Expected sessionId to be 'mock-session-id', got: %s", req.Form.Get("sessionId"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code":    gateway.RongcloudSuccessCode,
				"success": true,
			})
		})

	// 创建网关配置
	config := map[string]any{
		"app_key":    "mock-app-key",
		"app_secret": "mock-app-secret",
	}

	// 创建网关
	g := gateway.NewRongcloudGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetData(map[string]any{
			"action":    "verifyCode",
			"code":      "1234",
			"sessionId": "mock-session-id",
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

	code, ok := respMap["code"].(float64)
	if !ok || int(code) != gateway.RongcloudSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.RongcloudSuccessCode, code)
	}

	success, ok := respMap["success"].(bool)
	if !ok || !success {
		t.Errorf("Expected success to be true, got: %v", success)
	}
}

// TestRongcloudGatewayError 测试融云短信网关错误响应
func TestRongcloudGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", fmt.Sprintf(gateway.RongcloudEndpointTemplate, gateway.RongcloudEndpointAction, gateway.RongcloudEndpointFormat),
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"code":         1000,
			"errorMessage": "mock-error-message",
		}))

	// 创建网关配置
	config := map[string]any{
		"app_key":    "mock-app-key",
		"app_secret": "mock-app-secret",
	}

	// 创建网关
	g := gateway.NewRongcloudGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-tpl-id")

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误信息
	expectedError := "融云短信发送失败: [1000] mock-error-message"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', got: '%s'", expectedError, err.Error())
	}
}
