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

// TestAliyunIntlGateway 测试阿里云国际短信网关
func TestAliyunIntlGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("GET", gateway.AliyunIntlEndpointURL,
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			query := req.URL.Query()

			if query.Get("AccessKeyId") != "mock-api-key" {
				t.Errorf("Expected AccessKeyId to be 'mock-api-key', got: %s", query.Get("AccessKeyId"))
			}

			if query.Get("Action") != gateway.AliyunIntlEndpointAction {
				t.Errorf("Expected Action to be '%s', got: %s", gateway.AliyunIntlEndpointAction, query.Get("Action"))
			}

			if query.Get("Format") != gateway.AliyunIntlEndpointFormat {
				t.Errorf("Expected Format to be '%s', got: %s", gateway.AliyunIntlEndpointFormat, query.Get("Format"))
			}

			if query.Get("RegionId") != gateway.AliyunIntlEndpointRegionID {
				t.Errorf("Expected RegionId to be '%s', got: %s", gateway.AliyunIntlEndpointRegionID, query.Get("RegionId"))
			}

			if query.Get("SignatureMethod") != gateway.AliyunIntlEndpointSignatureMethod {
				t.Errorf("Expected SignatureMethod to be '%s', got: %s", gateway.AliyunIntlEndpointSignatureMethod, query.Get("SignatureMethod"))
			}

			if query.Get("SignatureVersion") != gateway.AliyunIntlEndpointSignatureVersion {
				t.Errorf("Expected SignatureVersion to be '%s', got: %s", gateway.AliyunIntlEndpointSignatureVersion, query.Get("SignatureVersion"))
			}

			if query.Get("To") != "18888888888" {
				t.Errorf("Expected To to be '18888888888', got: %s", query.Get("To"))
			}

			if query.Get("From") != "mock-api-sign-name" {
				t.Errorf("Expected From to be 'mock-api-sign-name', got: %s", query.Get("From"))
			}

			if query.Get("TemplateCode") != "mock-template-code" {
				t.Errorf("Expected TemplateCode to be 'mock-template-code', got: %s", query.Get("TemplateCode"))
			}

			// 验证模板参数
			templateParam := query.Get("TemplateParam")
			var templateData map[string]any
			if err := json.Unmarshal([]byte(templateParam), &templateData); err != nil {
				t.Errorf("Failed to parse TemplateParam: %v", err)
			}

			if code, ok := templateData["code"].(string); !ok || code != "123456" {
				t.Errorf("Expected TemplateParam.code to be '123456', got: %v", templateData["code"])
			}

			// 验证签名
			if query.Get("Signature") == "" {
				t.Errorf("Expected Signature to be set")
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"ResponseCode":        gateway.AliyunIntlSuccessCode,
				"ResponseDescription": "mock-result",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"access_key_id":     "mock-api-key",
		"access_key_secret": "mock-api-secret",
		"sign_name":         "mock-api-sign-name",
	}

	// 创建网关
	g := gateway.NewAliyunIntlGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-code").
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

	responseCode, ok := respMap["ResponseCode"].(string)
	if !ok || responseCode != gateway.AliyunIntlSuccessCode {
		t.Errorf("Expected ResponseCode to be '%s', got: %v", gateway.AliyunIntlSuccessCode, responseCode)
	}

	responseDescription, ok := respMap["ResponseDescription"].(string)
	if !ok || responseDescription != "mock-result" {
		t.Errorf("Expected ResponseDescription to be 'mock-result', got: %v", responseDescription)
	}
}

// TestAliyunIntlGatewayError 测试阿里云国际短信网关错误响应
func TestAliyunIntlGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("GET", gateway.AliyunIntlEndpointURL,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"ResponseCode":        "1234",
			"ResponseDescription": "mock-err-msg",
		}))

	// 创建网关配置
	config := map[string]any{
		"access_key_id":     "mock-api-key",
		"access_key_secret": "mock-api-secret",
		"sign_name":         "mock-api-sign-name",
	}

	// 创建网关
	g := gateway.NewAliyunIntlGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-code").
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
	expectedError := "阿里云国际短信发送失败: [1234] mock-err-msg"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestAliyunIntlGatewayWithCustomSignName 测试阿里云国际短信网关自定义签名
func TestAliyunIntlGatewayWithCustomSignName(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("GET", gateway.AliyunIntlEndpointURL,
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			query := req.URL.Query()

			if query.Get("From") != "custom-sign-name" {
				t.Errorf("Expected From to be 'custom-sign-name', got: %s", query.Get("From"))
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"ResponseCode":        gateway.AliyunIntlSuccessCode,
				"ResponseDescription": "mock-result",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"access_key_id":     "mock-api-key",
		"access_key_secret": "mock-api-secret",
		"sign_name":         "mock-api-sign-name",
	}

	// 创建网关
	g := gateway.NewAliyunIntlGateway(config)

	// 创建消息，包含自定义签名
	msg := message.NewMessage().
		SetTemplate("mock-template-code").
		SetData(map[string]any{
			"code":      "123456",
			"sign_name": "custom-sign-name",
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

	responseCode, ok := respMap["ResponseCode"].(string)
	if !ok || responseCode != gateway.AliyunIntlSuccessCode {
		t.Errorf("Expected ResponseCode to be '%s', got: %v", gateway.AliyunIntlSuccessCode, responseCode)
	}
}
