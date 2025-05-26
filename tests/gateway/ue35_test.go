package gateway

import (
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestUe35Gateway 测试联通短信网关
func TestUe35Gateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 构建请求 URL 模式
	endpointURI := "https://" + gateway.Ue35EndpointHost + gateway.Ue35EndpointURI

	// 注册成功响应
	httpmock.RegisterResponder("GET", `=~^`+endpointURI,
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			query := req.URL.Query()

			if query.Get("username") != "mock-username" {
				t.Errorf("Expected username to be 'mock-username', got: %s", query.Get("username"))
			}

			if query.Get("userpwd") != "mock-userpwd" {
				t.Errorf("Expected userpwd to be 'mock-userpwd', got: %s", query.Get("userpwd"))
			}

			if query.Get("mobiles") != "18888888888" {
				t.Errorf("Expected mobiles to be '18888888888', got: %s", query.Get("mobiles"))
			}

			if query.Get("content") != "This is a test message." {
				t.Errorf("Expected content to be 'This is a test message.', got: %s", query.Get("content"))
			}

			// 验证请求头
			if req.Header.Get("Host") != gateway.Ue35EndpointHost {
				t.Errorf("Expected Host to be '%s', got: %s", gateway.Ue35EndpointHost, req.Header.Get("Host"))
			}

			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type to be 'application/json', got: %s", req.Header.Get("Content-Type"))
			}

			if req.Header.Get("User-Agent") != "Go EasySms Client" {
				t.Errorf("Expected User-Agent to be 'Go EasySms Client', got: %s", req.Header.Get("User-Agent"))
			}

			// 返回 XML 成功响应
			return httpmock.NewStringResponse(200, `<?xml version="1.0" encoding="UTF-8"?>
<returnsms>
    <errorcode>1</errorcode>
    <message>发送成功</message>
</returnsms>`), nil
		})

	// 创建网关配置
	config := map[string]any{
		"username": "mock-username",
		"userpwd":  "mock-userpwd",
	}

	// 创建网关
	g := gateway.NewUe35Gateway(config)

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

	if respMap["errorcode"].(float64) != gateway.Ue35SuccessCode {
		t.Errorf("Expected errorcode to be %d, got: %v", gateway.Ue35SuccessCode, respMap["errorcode"])
	}

	if respMap["message"].(string) != "发送成功" {
		t.Errorf("Expected message to be '发送成功', got: %v", respMap["message"])
	}
}

// TestUe35GatewayError 测试联通短信网关错误响应
func TestUe35GatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 构建请求 URL 模式
	endpointURI := "https://" + gateway.Ue35EndpointHost + gateway.Ue35EndpointURI

	// 注册错误响应
	httpmock.RegisterResponder("GET", `=~^`+endpointURI,
		httpmock.NewStringResponder(200, `<?xml version="1.0" encoding="UTF-8"?>
<returnsms>
    <errorcode>0</errorcode>
    <message>用户名或密码错误</message>
</returnsms>`))

	// 创建网关配置
	config := map[string]any{
		"username": "mock-username",
		"userpwd":  "mock-userpwd",
	}

	// 创建网关
	g := gateway.NewUe35Gateway(config)

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

	// 验证错误消息
	expectedError := "ue35 gateway error: 用户名或密码错误 (code: 0)"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestUe35GatewayWithJSONResponse 测试联通短信网关 JSON 响应
func TestUe35GatewayWithJSONResponse(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 构建请求 URL 模式
	endpointURI := "https://" + gateway.Ue35EndpointHost + gateway.Ue35EndpointURI

	// 注册 JSON 成功响应
	httpmock.RegisterResponder("GET", `=~^`+endpointURI,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"errorcode": gateway.Ue35SuccessCode,
			"message":   "发送成功",
		}))

	// 创建网关配置
	config := map[string]any{
		"username": "mock-username",
		"userpwd":  "mock-userpwd",
	}

	// 创建网关
	g := gateway.NewUe35Gateway(config)

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

	if respMap["errorcode"].(float64) != gateway.Ue35SuccessCode {
		t.Errorf("Expected errorcode to be %d, got: %v", gateway.Ue35SuccessCode, respMap["errorcode"])
	}

	if respMap["message"].(string) != "发送成功" {
		t.Errorf("Expected message to be '发送成功', got: %v", respMap["message"])
	}
}
