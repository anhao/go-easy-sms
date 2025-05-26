package gateway

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestCtyunGateway 测试天翼云短信网关
func TestCtyunGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", gateway.CtyunEndpointHost+"/sms/api/v1",
		func(req *http.Request) (*http.Response, error) {
			// 验证请求头
			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type to be 'application/json', got: %s", req.Header.Get("Content-Type"))
			}

			if !strings.HasPrefix(req.Header.Get("Eop-Authorization"), "mock-access-key") {
				t.Errorf("Expected Eop-Authorization to start with 'mock-access-key', got: %s", req.Header.Get("Eop-Authorization"))
			}

			if req.Header.Get("ctyun-eop-request-id") == "" {
				t.Errorf("Expected ctyun-eop-request-id to be set")
			}

			if req.Header.Get("eop-date") == "" {
				t.Errorf("Expected eop-date to be set")
			}

			// 解析请求体
			var params map[string]any
			if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// 验证请求参数
			if phoneNumber, ok := params["phoneNumber"].(string); !ok || phoneNumber != "18888888888" {
				t.Errorf("Expected phoneNumber to be '18888888888', got: %v", phoneNumber)
			}

			if templateCode, ok := params["templateCode"].(string); !ok || templateCode != "mock-template-code" {
				t.Errorf("Expected templateCode to be 'mock-template-code', got: %v", templateCode)
			}

			if templateParam, ok := params["templateParam"].(string); !ok || templateParam != `{"code":"123456"}` {
				t.Errorf("Expected templateParam to be '{\"code\":\"123456\"}', got: %v", templateParam)
			}

			if signName, ok := params["signName"].(string); !ok || signName != "mock-sign-name" {
				t.Errorf("Expected signName to be 'mock-sign-name', got: %v", signName)
			}

			if action, ok := params["action"].(string); !ok || action != "SendSms" {
				t.Errorf("Expected action to be 'SendSms', got: %v", action)
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code":      gateway.CtyunSuccessCode,
				"message":   "Success",
				"requestId": "mock-request-id",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"secret_key":    "mock-secret-key",
		"access_key":    "mock-access-key",
		"template_code": "mock-template-code",
		"sign_name":     "mock-sign-name",
	}

	// 创建网关
	g := gateway.NewCtyunGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-tpl-id").
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
	if !ok || code != gateway.CtyunSuccessCode {
		t.Errorf("Expected code to be '%s', got: %v", gateway.CtyunSuccessCode, code)
	}

	message, ok := respMap["message"].(string)
	if !ok || message != "Success" {
		t.Errorf("Expected message to be 'Success', got: %v", message)
	}
}

// TestCtyunGatewayError 测试天翼云短信网关错误响应
func TestCtyunGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", gateway.CtyunEndpointHost+"/sms/api/v1",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"code":      "FAIL",
			"message":   "mock-error-message",
			"requestId": "mock-request-id",
		}))

	// 创建网关配置
	config := map[string]any{
		"secret_key":    "mock-secret-key",
		"access_key":    "mock-access-key",
		"template_code": "mock-template-code",
		"sign_name":     "mock-sign-name",
	}

	// 创建网关
	g := gateway.NewCtyunGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-tpl-id").
		SetData(map[string]any{
			"code": "123456",
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误信息
	expectedError := "天翼云短信发送失败: [FAIL] mock-error-message"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
