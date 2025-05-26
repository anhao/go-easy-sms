package gateway

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestSmsbaoGatewayWithSMS 测试短信宝网关国内短信
func TestSmsbaoGatewayWithSMS(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 计算密码 MD5
	h := md5.New()
	h.Write([]byte("mock-password"))
	passwordMD5 := hex.EncodeToString(h.Sum(nil))

	// 注册成功响应
	httpmock.RegisterResponder("GET", "http://api.smsbao.com/sms",
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			query := req.URL.Query()

			if query.Get("u") != "mock-user" {
				t.Errorf("Expected u to be 'mock-user', got: %s", query.Get("u"))
			}

			if query.Get("p") != passwordMD5 {
				t.Errorf("Expected p to be '%s', got: %s", passwordMD5, query.Get("p"))
			}

			if query.Get("m") != "18188888888" {
				t.Errorf("Expected m to be '18188888888', got: %s", query.Get("m"))
			}

			if query.Get("c") != "This is a test message." {
				t.Errorf("Expected c to be 'This is a test message.', got: %s", query.Get("c"))
			}

			// 返回成功响应
			return httpmock.NewStringResponse(200, gateway.SmsbaoSuccessCode), nil
		})

	// 创建网关配置
	config := map[string]any{
		"user":     "mock-user",
		"password": "mock-password",
	}

	// 创建网关
	g := gateway.NewSmsbaoGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("This is a test message.")

	// 创建电话号码
	phone := message.NewPhoneNumber("18188888888")

	// 测试发送
	resp, err := g.Send(phone, msg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 验证响应
	if resp != gateway.SmsbaoSuccessCode {
		t.Errorf("Expected response to be '%s', got: %v", gateway.SmsbaoSuccessCode, resp)
	}
}

// TestSmsbaoGatewayWithWSMS 测试短信宝网关国际短信
func TestSmsbaoGatewayWithWSMS(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 计算密码 MD5
	h := md5.New()
	h.Write([]byte("mock-password"))
	passwordMD5 := hex.EncodeToString(h.Sum(nil))

	// 注册成功响应
	httpmock.RegisterResponder("GET", "http://api.smsbao.com/wsms",
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			query := req.URL.Query()

			if query.Get("u") != "mock-user" {
				t.Errorf("Expected u to be 'mock-user', got: %s", query.Get("u"))
			}

			if query.Get("p") != passwordMD5 {
				t.Errorf("Expected p to be '%s', got: %s", passwordMD5, query.Get("p"))
			}

			if query.Get("m") != "+8518188888888" {
				t.Errorf("Expected m to be '+8518188888888', got: %s", query.Get("m"))
			}

			if query.Get("c") != "This is a test message." {
				t.Errorf("Expected c to be 'This is a test message.', got: %s", query.Get("c"))
			}

			// 返回成功响应
			return httpmock.NewStringResponse(200, gateway.SmsbaoSuccessCode), nil
		})

	// 创建网关配置
	config := map[string]any{
		"user":     "mock-user",
		"password": "mock-password",
	}

	// 创建网关
	g := gateway.NewSmsbaoGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("This is a test message.")

	// 创建电话号码
	phone := message.NewPhoneNumber("18188888888", 85)

	// 测试发送
	resp, err := g.Send(phone, msg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 验证响应
	if resp != gateway.SmsbaoSuccessCode {
		t.Errorf("Expected response to be '%s', got: %v", gateway.SmsbaoSuccessCode, resp)
	}
}

// TestSmsbaoGatewayError 测试短信宝网关错误响应
func TestSmsbaoGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("GET", "http://api.smsbao.com/sms",
		httpmock.NewStringResponder(200, "30"))

	// 创建网关配置
	config := map[string]any{
		"user":     "mock-user",
		"password": "mock-password",
	}

	// 创建网关
	g := gateway.NewSmsbaoGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("This is a test message.")

	// 创建电话号码
	phone := message.NewPhoneNumber("18188888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误信息
	expectedError := "短信宝短信发送失败: [30] 密码错误"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
