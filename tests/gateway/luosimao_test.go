package gateway

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestLuosimaoGateway 测试螺丝帽短信网关
func TestLuosimaoGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 构建请求地址
	endpoint := "https://sms-api.luosimao.com/v1/send.json"

	// 注册成功响应
	httpmock.RegisterResponder("POST", endpoint,
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("mobile") != "18888888888" {
				t.Errorf("Expected mobile to be '18888888888', got: %s", req.Form.Get("mobile"))
			}

			if req.Form.Get("message") != "This is a test message." {
				t.Errorf("Expected message to be 'This is a test message.', got: %s", req.Form.Get("message"))
			}

			// 验证请求头
			expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("api:key-mock-api-key"))
			if req.Header.Get("Authorization") != expectedAuth {
				t.Errorf("Expected Authorization to be '%s', got: %s", expectedAuth, req.Header.Get("Authorization"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"error": 0,
				"msg":   "success",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"api_key": "mock-api-key",
	}

	// 创建网关
	g := gateway.NewLuosimaoGateway(config)

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

	errorCode, ok := respMap["error"].(float64)
	if !ok || int(errorCode) != 0 {
		t.Errorf("Expected error to be 0, got: %v", errorCode)
	}

	respMsg, ok := respMap["msg"].(string)
	if !ok || respMsg != "success" {
		t.Errorf("Expected msg to be 'success', got: %v", respMsg)
	}
}

// TestLuosimaoGatewayError 测试螺丝帽短信网关错误响应
func TestLuosimaoGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 构建请求地址
	endpoint := "https://sms-api.luosimao.com/v1/send.json"

	// 注册错误响应
	httpmock.RegisterResponder("POST", endpoint,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"error": 10000,
			"msg":   "mock-err-msg",
		}))

	// 创建网关配置
	config := map[string]any{
		"api_key": "mock-api-key",
	}

	// 创建网关
	g := gateway.NewLuosimaoGateway(config)

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
	expectedError := "螺丝帽短信发送失败: [10000] mock-err-msg"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
