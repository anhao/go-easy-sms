package gateway

import (
	"net/http"
	"strings"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestQcloudGateway 测试腾讯云短信网关
func TestQcloudGateway(t *testing.T) {
	// 设置模拟服务器
	baseURL := "https://sms.tencentcloudapi.com"

	// 激活 httpmock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 重置 httpmock 并设置成功响应
	httpmock.Reset()
	httpmock.RegisterResponder("POST", baseURL,
		httpmock.NewJsonResponderOrPanic(http.StatusOK, map[string]any{
			"Response": map[string]any{
				"SendStatusSet": []map[string]any{
					{
						"SerialNo":       "2028:f825e6b16e23f73f4123",
						"PhoneNumber":    "8618888888888",
						"Fee":            1,
						"SessionContext": "",
						"Code":           "Ok",
						"Message":        "send success",
						"IsoCode":        "CN",
					},
				},
			},
			"RequestId": "0dc99542-c61a-4a16-9545-ec8ec202c543",
		}))

	// 创建网关配置
	config := map[string]any{
		"sdk_app_id": "mock-sdk-app-id",
		"secret_key": "mock-secret-key",
		"secret_id":  "mock-secret-id",
		"sign_name":  "mock-api-sign-name",
		"endpoint":   baseURL,
	}

	// 创建网关
	g := gateway.NewQcloudGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("template-id").
		SetData(map[string]any{
			"0": "888888",
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送成功
	resp, err := g.Send(phone, msg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 验证响应
	respMap, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("Expected response to be map[string]any, got: %T", resp)
	}

	response, ok := respMap["Response"].(map[string]any)
	if !ok {
		t.Fatalf("Expected Response to be map[string]any, got: %T", respMap["Response"])
	}

	statusSet, ok := response["SendStatusSet"].([]any)
	if !ok || len(statusSet) == 0 {
		t.Fatalf("Expected SendStatusSet to be non-empty array, got: %T", response["SendStatusSet"])
	}

	status, ok := statusSet[0].(map[string]any)
	if !ok {
		t.Fatalf("Expected status to be map[string]any, got: %T", statusSet[0])
	}

	code, ok := status["Code"].(string)
	if !ok || code != "Ok" {
		t.Errorf("Expected Code to be 'Ok', got: %v", code)
	}

	// 重置 httpmock 并设置错误响应
	httpmock.Reset()
	httpmock.RegisterResponder("POST", baseURL,
		httpmock.NewJsonResponderOrPanic(http.StatusOK, map[string]any{
			"Response": map[string]any{
				"Error": map[string]any{
					"Code":    "AuthFailure.SignatureFailure",
					"Message": "The provided credentials could not be validated. Please check your signature is correct.",
				},
			},
			"RequestId": "0dc99542-c61a-4a16-9545-2b967e2c980a",
		}))

	// 测试发送失败 - 认证错误
	_, err = g.Send(phone, msg)
	if err == nil {
		t.Errorf("Expected error for invalid credentials")
	} else if !contains(err.Error(), "AuthFailure.SignatureFailure") && !contains(err.Error(), "credentials could not be validated") {
		t.Errorf("Expected error message to contain 'AuthFailure.SignatureFailure' or 'credentials could not be validated', got: %v", err)
	}

	// 重置 httpmock 并设置模板参数错误响应
	httpmock.Reset()
	httpmock.RegisterResponder("POST", baseURL,
		httpmock.NewJsonResponderOrPanic(http.StatusOK, map[string]any{
			"Response": map[string]any{
				"SendStatusSet": []map[string]any{
					{
						"SerialNo":       "2028:f825e6b16e23f73f4123",
						"PhoneNumber":    "8618888888888",
						"Fee":            1,
						"SessionContext": "",
						"Code":           "InvalidParameterValue.TemplateParameterFormatError",
						"Message":        "Verification code template parameter format error",
						"IsoCode":        "CN",
					},
				},
			},
			"RequestId": "0dc99542-c61a-4a16-9545-ec8ec202c543",
		}))

	// 测试发送失败 - 模板参数错误
	_, err = g.Send(phone, msg)
	if err == nil {
		t.Errorf("Expected error for template parameter format error")
	} else if !contains(err.Error(), "InvalidParameterValue.TemplateParameterFormatError") && !contains(err.Error(), "template parameter format error") {
		t.Errorf("Expected error message to contain 'InvalidParameterValue.TemplateParameterFormatError' or 'template parameter format error', got: %v", err)
	}
}

// TestQcloudGatewayWithSignName 测试使用消息中的签名
func TestQcloudGatewayWithSignName(t *testing.T) {
	// 设置模拟服务器
	baseURL := "https://sms.tencentcloudapi.com"

	// 激活 httpmock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 重置 httpmock 并设置成功响应
	httpmock.Reset()
	httpmock.RegisterResponder("POST", baseURL,
		httpmock.NewJsonResponderOrPanic(http.StatusOK, map[string]any{
			"Response": map[string]any{
				"SendStatusSet": []map[string]any{
					{
						"SerialNo":       "2028:f825e6b16e23f73f4123",
						"PhoneNumber":    "8618888888888",
						"Fee":            1,
						"SessionContext": "",
						"Code":           "Ok",
						"Message":        "send success",
						"IsoCode":        "CN",
					},
				},
			},
			"RequestId": "0dc99542-c61a-4a16-9545-ec8ec202c543",
		}))

	// 创建网关配置
	config := map[string]any{
		"sdk_app_id": "mock-sdk-app-id",
		"secret_key": "mock-secret-key",
		"secret_id":  "mock-secret-id",
		"sign_name":  "mock-api-sign-name",
		"endpoint":   baseURL,
	}

	// 创建网关
	g := gateway.NewQcloudGateway(config)

	// 创建消息，包含自定义签名
	msg := message.NewMessage().
		SetTemplate("template-id").
		SetData(map[string]any{
			"0":         "888888",
			"sign_name": "custom-sign-name",
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 验证 sign_name 字段已从数据中删除
	if _, ok := msg.GetData()["sign_name"]; ok {
		t.Errorf("Expected sign_name to be removed from message data")
	}
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
