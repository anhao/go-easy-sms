package gateway

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestHuyiGateway 测试互亿无线短信网关
func TestHuyiGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", gateway.HuyiEndpointURL,
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("account") != "mock-api-id" {
				t.Errorf("Expected account to be 'mock-api-id', got: %s", req.Form.Get("account"))
			}

			if req.Form.Get("mobile") != "18888888888" {
				t.Errorf("Expected mobile to be '18888888888', got: %s", req.Form.Get("mobile"))
			}

			if req.Form.Get("content") != "This is a test message." {
				t.Errorf("Expected content to be 'This is a test message.', got: %s", req.Form.Get("content"))
			}

			if req.Form.Get("format") != gateway.HuyiEndpointFormat {
				t.Errorf("Expected format to be '%s', got: %s", gateway.HuyiEndpointFormat, req.Form.Get("format"))
			}

			// 验证签名
			timestamp := req.Form.Get("time")
			signStr := "mock-api-id" + "mock-api-key" + "18888888888" + "This is a test message." + timestamp
			h := md5.New()
			h.Write([]byte(signStr))
			expectedPassword := fmt.Sprintf("%x", h.Sum(nil))

			if req.Form.Get("password") != expectedPassword {
				t.Errorf("Expected password to be '%s', got: %s", expectedPassword, req.Form.Get("password"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code": gateway.HuyiSuccessCode,
				"msg":  "mock-result",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"api_id":  "mock-api-id",
		"api_key": "mock-api-key",
	}

	// 创建网关
	g := gateway.NewHuyiGateway(config)

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
	if !ok || int(code) != gateway.HuyiSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.HuyiSuccessCode, code)
	}

	respMsg, ok := respMap["msg"].(string)
	if !ok || respMsg != "mock-result" {
		t.Errorf("Expected msg to be 'mock-result', got: %v", respMsg)
	}
}

// TestHuyiGatewayError 测试互亿无线短信网关错误响应
func TestHuyiGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", gateway.HuyiEndpointURL,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"code": 1234,
			"msg":  "mock-err-msg",
		}))

	// 创建网关配置
	config := map[string]any{
		"api_id":  "mock-api-id",
		"api_key": "mock-api-key",
	}

	// 创建网关
	g := gateway.NewHuyiGateway(config)

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
	expectedError := "互亿无线短信发送失败: [1234] mock-err-msg"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestHuyiGatewayWithInternationalNumber 测试互亿无线短信网关国际号码
func TestHuyiGatewayWithInternationalNumber(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", gateway.HuyiEndpointURL,
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("mobile") != "1 8888888888" {
				t.Errorf("Expected mobile to be '1 8888888888', got: %s", req.Form.Get("mobile"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code": gateway.HuyiSuccessCode,
				"msg":  "mock-result",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"api_id":  "mock-api-id",
		"api_key": "mock-api-key",
	}

	// 创建网关
	g := gateway.NewHuyiGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("This is a test message.")

	// 创建国际电话号码
	phone := message.NewPhoneNumber("8888888888", 1)

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
	if !ok || int(code) != gateway.HuyiSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.HuyiSuccessCode, code)
	}
}
