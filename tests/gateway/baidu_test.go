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

// TestBaiduGateway 测试百度云短信网关
func TestBaiduGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", "http://smsv3.bj.baidubce.com/api/v3/sendSms",
		func(req *http.Request) (*http.Response, error) {
			// 验证请求头
			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type to be 'application/json', got: %s", req.Header.Get("Content-Type"))
			}

			if req.Header.Get("host") != gateway.BaiduEndpointHost {
				t.Errorf("Expected host to be '%s', got: %s", gateway.BaiduEndpointHost, req.Header.Get("host"))
			}

			if !strings.Contains(req.Header.Get("Authorization"), gateway.BaiduAuthVersion) {
				t.Errorf("Expected Authorization header to contain '%s', got: %s", gateway.BaiduAuthVersion, req.Header.Get("Authorization"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code":    gateway.BaiduSuccessCode,
				"message": "success",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"ak":        "mock-ak",
		"sk":        "mock-sk",
		"invoke_id": "mock-invoke-id",
	}

	// 创建网关
	g := gateway.NewBaiduGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-tpl-id").
		SetData(map[string]any{
			"mock-data-1": "value1",
			"mock-data-2": "value2",
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
	if !ok || int(code) != gateway.BaiduSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.BaiduSuccessCode, code)
	}

	message, ok := respMap["message"].(string)
	if !ok || message != "success" {
		t.Errorf("Expected message to be 'success', got: %v", message)
	}
}

// TestBaiduGatewayError 测试百度云短信网关错误响应
func TestBaiduGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", "http://smsv3.bj.baidubce.com/api/v3/sendSms",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"code":    100,
			"message": "mock-error-message",
		}))

	// 创建网关配置
	config := map[string]any{
		"ak":        "mock-ak",
		"sk":        "mock-sk",
		"invoke_id": "mock-invoke-id",
	}

	// 创建网关
	g := gateway.NewBaiduGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-tpl-id").
		SetData(map[string]any{
			"mock-data-1": "value1",
			"mock-data-2": "value2",
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误信息
	expectedError := "百度云短信发送失败: [100] mock-error-message"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestBaiduGatewayWithCustomDomain 测试百度云短信网关自定义域名
func TestBaiduGatewayWithCustomDomain(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 自定义域名
	customDomain := "custom-sms.baidubce.com"

	// 注册成功响应
	httpmock.RegisterResponder("POST", "http://"+customDomain+"/api/v3/sendSms",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"code":    gateway.BaiduSuccessCode,
			"message": "success",
		}))

	// 创建网关配置
	config := map[string]any{
		"ak":        "mock-ak",
		"sk":        "mock-sk",
		"invoke_id": "mock-invoke-id",
		"domain":    customDomain,
	}

	// 创建网关
	g := gateway.NewBaiduGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-tpl-id").
		SetData(map[string]any{
			"mock-data-1": "value1",
			"mock-data-2": "value2",
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
	if !ok || int(code) != gateway.BaiduSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.BaiduSuccessCode, code)
	}
}

// TestBaiduGatewayWithCustomParams 测试百度云短信网关自定义参数
func TestBaiduGatewayWithCustomParams(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", "http://smsv3.bj.baidubce.com/api/v3/sendSms",
		func(req *http.Request) (*http.Response, error) {
			// 解析请求体
			var params map[string]any
			if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// 验证自定义参数
			if custom, ok := params["custom"]; !ok || custom != "custom-value" {
				t.Errorf("Expected custom to be 'custom-value', got: %v", custom)
			}

			if userExtId, ok := params["userExtId"]; !ok || userExtId != "123456" {
				t.Errorf("Expected userExtId to be '123456', got: %v", userExtId)
			}

			// 验证 contentVar 中不包含 custom 和 userExtId
			if contentVar, ok := params["contentVar"].(map[string]any); ok {
				if _, ok := contentVar["custom"]; ok {
					t.Errorf("Expected contentVar not to contain 'custom'")
				}
				if _, ok := contentVar["userExtId"]; ok {
					t.Errorf("Expected contentVar not to contain 'userExtId'")
				}
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code":    gateway.BaiduSuccessCode,
				"message": "success",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"ak":        "mock-ak",
		"sk":        "mock-sk",
		"invoke_id": "mock-invoke-id",
	}

	// 创建网关
	g := gateway.NewBaiduGateway(config)

	// 创建消息，包含自定义参数
	msg := message.NewMessage().
		SetTemplate("mock-tpl-id").
		SetData(map[string]any{
			"mock-data-1": "value1",
			"mock-data-2": "value2",
			"custom":      "custom-value",
			"userExtId":   "123456",
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
	if !ok || int(code) != gateway.BaiduSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.BaiduSuccessCode, code)
	}
}
