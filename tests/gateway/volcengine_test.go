package gateway

import (
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestVolcengineGateway 测试火山引擎短信网关
func TestVolcengineGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", `=~^https://sms.volcengineapi.com\?Action=SendSms&Version=2020-01-01`,
		func(req *http.Request) (*http.Response, error) {
			// 验证请求头
			if req.Header.Get("Content-Type") != gateway.VolcengineEndpointContentType {
				t.Errorf("Expected Content-Type to be %s, got: %s", gateway.VolcengineEndpointContentType, req.Header.Get("Content-Type"))
			}

			if req.Header.Get("Accept") != gateway.VolcengineEndpointAccept {
				t.Errorf("Expected Accept to be %s, got: %s", gateway.VolcengineEndpointAccept, req.Header.Get("Accept"))
			}

			if req.Header.Get("User-Agent") != gateway.VolcengineEndpointUserAgent {
				t.Errorf("Expected User-Agent to be %s, got: %s", gateway.VolcengineEndpointUserAgent, req.Header.Get("User-Agent"))
			}

			// 验证授权头
			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				t.Errorf("Expected Authorization header to be set")
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"ResponseMetadata": map[string]any{
					"RequestId": "mock-request-id",
					"Action":    "SendSms",
					"Version":   "2020-01-01",
					"Service":   "volcSMS",
				},
				"Result": map[string]any{
					"MessageId": "mock-message-id",
					"Code":      "OK",
					"Message":   "Success",
				},
			})
		})

	// 创建网关配置
	config := map[string]any{
		"access_key_id":     "mock-access-key-id",
		"access_key_secret": "mock-access-key-secret",
		"sign_name":         "mock-sign-name",
		"sms_account":       "mock-sms-account",
	}

	// 创建网关
	g := gateway.NewVolcengineGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-id").
		SetData(map[string]any{
			"code": "1234",
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

	// 验证响应内容
	if respMap["ResponseMetadata"] == nil {
		t.Fatalf("Expected ResponseMetadata in response, got: %v", respMap)
	}

	metadata := respMap["ResponseMetadata"].(map[string]any)
	if metadata["RequestId"] != "mock-request-id" {
		t.Errorf("Expected RequestId to be 'mock-request-id', got: %v", metadata["RequestId"])
	}

	if respMap["Result"] == nil {
		t.Fatalf("Expected Result in response, got: %v", respMap)
	}

	result := respMap["Result"].(map[string]any)
	if result["MessageId"] != "mock-message-id" {
		t.Errorf("Expected MessageId to be 'mock-message-id', got: %v", result["MessageId"])
	}

	if result["Code"] != "OK" {
		t.Errorf("Expected Code to be 'OK', got: %v", result["Code"])
	}
}

// TestVolcengineGatewayWithCustomData 测试火山引擎短信网关自定义数据
func TestVolcengineGatewayWithCustomData(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", `=~^https://sms.volcengineapi.com\?Action=SendSms&Version=2020-01-01`,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"ResponseMetadata": map[string]any{
				"RequestId": "mock-request-id",
				"Action":    "SendSms",
				"Version":   "2020-01-01",
				"Service":   "volcSMS",
			},
			"Result": map[string]any{
				"MessageId": "mock-message-id",
				"Code":      "OK",
				"Message":   "Success",
			},
		}))

	// 创建网关配置
	config := map[string]any{
		"access_key_id":     "mock-access-key-id",
		"access_key_secret": "mock-access-key-secret",
		"sign_name":         "default-sign-name",
		"sms_account":       "default-sms-account",
	}

	// 创建网关
	g := gateway.NewVolcengineGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-id").
		SetData(map[string]any{
			"code":          "1234",
			"sign_name":     "custom-sign-name",
			"sms_account":   "custom-sms-account",
			"phone_numbers": "18888888888,19999999999",
			"template_param": map[string]any{
				"custom": "value",
			},
			"tag": "custom-tag",
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

	// 验证响应内容
	if respMap["Result"] == nil {
		t.Fatalf("Expected Result in response, got: %v", respMap)
	}

	result := respMap["Result"].(map[string]any)
	if result["Code"] != "OK" {
		t.Errorf("Expected Code to be 'OK', got: %v", result["Code"])
	}
}

// TestVolcengineGatewayError 测试火山引擎短信网关错误响应
func TestVolcengineGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", `=~^https://sms.volcengineapi.com\?Action=SendSms&Version=2020-01-01`,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"ResponseMetadata": map[string]any{
				"RequestId": "mock-request-id",
				"Action":    "SendSms",
				"Version":   "2020-01-01",
				"Service":   "volcSMS",
				"Error": map[string]any{
					"Code":    "InvalidParameter",
					"Message": "Invalid parameter value",
				},
			},
		}))

	// 创建网关配置
	config := map[string]any{
		"access_key_id":     "mock-access-key-id",
		"access_key_secret": "mock-access-key-secret",
		"sign_name":         "mock-sign-name",
		"sms_account":       "mock-sms-account",
	}

	// 创建网关
	g := gateway.NewVolcengineGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-id").
		SetData(map[string]any{
			"code": "1234",
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误消息
	expectedError := "volcengine gateway error: Invalid parameter value"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestVolcengineGatewayWithCustomRegion 测试火山引擎短信网关自定义区域
func TestVolcengineGatewayWithCustomRegion(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", `=~^https://sms.byteplusapi.com\?Action=SendSms&Version=2020-01-01`,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"ResponseMetadata": map[string]any{
				"RequestId": "mock-request-id",
				"Action":    "SendSms",
				"Version":   "2020-01-01",
				"Service":   "volcSMS",
			},
			"Result": map[string]any{
				"MessageId": "mock-message-id",
				"Code":      "OK",
				"Message":   "Success",
			},
		}))

	// 创建网关配置
	config := map[string]any{
		"access_key_id":     "mock-access-key-id",
		"access_key_secret": "mock-access-key-secret",
		"sign_name":         "mock-sign-name",
		"sms_account":       "mock-sms-account",
		"region_id":         "ap-singapore-1",
	}

	// 创建网关
	g := gateway.NewVolcengineGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-id").
		SetData(map[string]any{
			"code": "1234",
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

	// 验证响应内容
	if respMap["Result"] == nil {
		t.Fatalf("Expected Result in response, got: %v", respMap)
	}

	result := respMap["Result"].(map[string]any)
	if result["Code"] != "OK" {
		t.Errorf("Expected Code to be 'OK', got: %v", result["Code"])
	}
}
