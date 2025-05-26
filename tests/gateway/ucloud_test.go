package gateway

import (
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestUcloudGateway 测试 UCloud 短信网关
func TestUcloudGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("GET", "https://api.ucloud.cn",
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			query := req.URL.Query()
			if query.Get("Action") != gateway.UcloudEndpointAction {
				t.Errorf("Expected Action to be %s, got: %s", gateway.UcloudEndpointAction, query.Get("Action"))
			}
			if query.Get("PublicKey") != "mock-public-key" {
				t.Errorf("Expected PublicKey to be mock-public-key, got: %s", query.Get("PublicKey"))
			}
			if query.Get("TemplateId") != "mock-template-id" {
				t.Errorf("Expected TemplateId to be mock-template-id, got: %s", query.Get("TemplateId"))
			}
			if query.Get("SigContent") != "mock-sig-content" {
				t.Errorf("Expected SigContent to be mock-sig-content, got: %s", query.Get("SigContent"))
			}
			if query.Get("PhoneNumbers.0") != "18888888888" {
				t.Errorf("Expected PhoneNumbers.0 to be 18888888888, got: %s", query.Get("PhoneNumbers.0"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"RetCode": gateway.UcloudSuccessCode,
				"Message": "Success",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"private_key": "mock-private-key",
		"public_key":  "mock-public-key",
		"sig_content": "mock-sig-content",
		"project_id":  "mock-project-id",
	}

	// 创建网关
	g := gateway.NewUcloudGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-id").
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

	retCode, ok := respMap["RetCode"].(float64)
	if !ok || int(retCode) != gateway.UcloudSuccessCode {
		t.Errorf("Expected RetCode to be %d, got: %v", gateway.UcloudSuccessCode, retCode)
	}
}

// TestUcloudGatewayError 测试 UCloud 短信网关错误响应
func TestUcloudGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("GET", "https://api.ucloud.cn",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"RetCode": 170,
			"Message": "Missing signature",
		}))

	// 创建网关配置
	config := map[string]any{
		"private_key": "mock-private-key",
		"public_key":  "mock-public-key",
		"sig_content": "mock-sig-content",
		"project_id":  "mock-project-id",
	}

	// 创建网关
	g := gateway.NewUcloudGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-id").
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
	expectedError := "UCloud 短信发送失败: [170] Missing signature"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestUcloudGatewayWithCustomSigContent 测试 UCloud 短信网关自定义签名内容
func TestUcloudGatewayWithCustomSigContent(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("GET", "https://api.ucloud.cn",
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			query := req.URL.Query()
			if query.Get("SigContent") != "custom-sig-content" {
				t.Errorf("Expected SigContent to be custom-sig-content, got: %s", query.Get("SigContent"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"RetCode": gateway.UcloudSuccessCode,
				"Message": "Success",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"private_key": "mock-private-key",
		"public_key":  "mock-public-key",
		"sig_content": "mock-sig-content",
		"project_id":  "mock-project-id",
	}

	// 创建网关
	g := gateway.NewUcloudGateway(config)

	// 创建消息，包含自定义签名内容
	msg := message.NewMessage().
		SetTemplate("mock-template-id").
		SetData(map[string]any{
			"code":        "123456",
			"sig_content": "custom-sig-content",
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

	retCode, ok := respMap["RetCode"].(float64)
	if !ok || int(retCode) != gateway.UcloudSuccessCode {
		t.Errorf("Expected RetCode to be %d, got: %v", gateway.UcloudSuccessCode, retCode)
	}
}
