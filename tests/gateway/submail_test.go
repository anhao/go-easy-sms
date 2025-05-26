package gateway

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestSubmailGatewayWithTemplate 测试赛邮云短信网关模板发送
func TestSubmailGatewayWithTemplate(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", "https://api.mysubmail.com/message/xsend.json",
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("appid") != "mock-app-id" {
				t.Errorf("Expected appid to be 'mock-app-id', got: %s", req.Form.Get("appid"))
			}

			if req.Form.Get("signature") != "mock-app-key" {
				t.Errorf("Expected signature to be 'mock-app-key', got: %s", req.Form.Get("signature"))
			}

			if req.Form.Get("project") != "mock-project" {
				t.Errorf("Expected project to be 'mock-project', got: %s", req.Form.Get("project"))
			}

			if req.Form.Get("to") != "18888888888" {
				t.Errorf("Expected to to be '18888888888', got: %s", req.Form.Get("to"))
			}

			// 验证 vars 参数
			vars := req.Form.Get("vars")
			var varsData map[string]any
			if err := json.Unmarshal([]byte(vars), &varsData); err != nil {
				t.Errorf("Failed to parse vars: %v", err)
			}

			if code, ok := varsData["code"].(string); !ok || code != "123456" {
				t.Errorf("Expected vars.code to be '123456', got: %v", varsData["code"])
			}

			if time, ok := varsData["time"].(string); !ok || time != "15" {
				t.Errorf("Expected vars.time to be '15', got: %v", varsData["time"])
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"status":      gateway.SubmailSuccessStatus,
				"send_id":     "093c0a7df143c087d6cba9cdf0cf3738",
				"fee":         1,
				"sms_credits": 14197,
			})
		})

	// 创建网关配置
	config := map[string]any{
		"app_id":  "mock-app-id",
		"app_key": "mock-app-key",
		"project": "mock-project",
	}

	// 创建网关
	g := gateway.NewSubmailGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetData(map[string]any{
			"code": "123456",
			"time": "15",
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

	status, ok := respMap["status"].(string)
	if !ok || status != gateway.SubmailSuccessStatus {
		t.Errorf("Expected status to be '%s', got: %v", gateway.SubmailSuccessStatus, status)
	}

	sendID, ok := respMap["send_id"].(string)
	if !ok || sendID != "093c0a7df143c087d6cba9cdf0cf3738" {
		t.Errorf("Expected send_id to be '093c0a7df143c087d6cba9cdf0cf3738', got: %v", sendID)
	}
}

// TestSubmailGatewayWithContent 测试赛邮云短信网关内容发送
func TestSubmailGatewayWithContent(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", "https://api.mysubmail.com/sms/send.json",
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("appid") != "mock-app-id" {
				t.Errorf("Expected appid to be 'mock-app-id', got: %s", req.Form.Get("appid"))
			}

			if req.Form.Get("signature") != "mock-app-key" {
				t.Errorf("Expected signature to be 'mock-app-key', got: %s", req.Form.Get("signature"))
			}

			if req.Form.Get("content") != "【easysms】 mock-app-content" {
				t.Errorf("Expected content to be '【easysms】 mock-app-content', got: %s", req.Form.Get("content"))
			}

			if req.Form.Get("to") != "18888888888" {
				t.Errorf("Expected to to be '18888888888', got: %s", req.Form.Get("to"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"status":      gateway.SubmailSuccessStatus,
				"send_id":     "093c0a7df143c087d6cba9cdf0cf3738",
				"fee":         1,
				"sms_credits": 14197,
			})
		})

	// 创建网关配置
	config := map[string]any{
		"app_id":  "mock-app-id",
		"app_key": "mock-app-key",
	}

	// 创建网关
	g := gateway.NewSubmailGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("【easysms】 mock-app-content")

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
	if !ok || status != gateway.SubmailSuccessStatus {
		t.Errorf("Expected status to be '%s', got: %v", gateway.SubmailSuccessStatus, status)
	}
}

// TestSubmailGatewayError 测试赛邮云短信网关错误响应
func TestSubmailGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", "https://api.mysubmail.com/message/xsend.json",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"status": "error",
			"code":   100,
			"msg":    "mock-err-msg",
		}))

	// 创建网关配置
	config := map[string]any{
		"app_id":  "mock-app-id",
		"app_key": "mock-app-key",
		"project": "mock-project",
	}

	// 创建网关
	g := gateway.NewSubmailGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetData(map[string]any{
			"code": "123456",
			"time": "15",
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误信息
	expectedError := "赛邮云短信发送失败: [100] mock-err-msg"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestSubmailGatewayInternational 测试赛邮云短信网关国际短信
func TestSubmailGatewayInternational(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", "https://api.mysubmail.com/internationalsms/xsend.json",
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("to") != "+118888888888" {
				t.Errorf("Expected to to be '+118888888888', got: %s", req.Form.Get("to"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"status":      gateway.SubmailSuccessStatus,
				"send_id":     "093c0a7df143c087d6cba9cdf0cf3738",
				"fee":         1,
				"sms_credits": 14197,
			})
		})

	// 创建网关配置
	config := map[string]any{
		"app_id":  "mock-app-id",
		"app_key": "mock-app-key",
		"project": "mock-project",
	}

	// 创建网关
	g := gateway.NewSubmailGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetData(map[string]any{
			"code": "123456",
			"time": "15",
		})

	// 创建国际电话号码
	phone := message.NewPhoneNumber("18888888888", 1)

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
	if !ok || status != gateway.SubmailSuccessStatus {
		t.Errorf("Expected status to be '%s', got: %v", gateway.SubmailSuccessStatus, status)
	}
}
