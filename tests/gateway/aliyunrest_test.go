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

// TestAliyunrestGateway 测试阿里云 REST API 短信网关
func TestAliyunrestGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", gateway.AliyunrestEndpointURL,
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("sms_free_sign_name") != "mock-app-sign-name" {
				t.Errorf("Expected sms_free_sign_name to be 'mock-app-sign-name', got: %s", req.Form.Get("sms_free_sign_name"))
			}

			if req.Form.Get("sms_template_code") != "mock-template-code" {
				t.Errorf("Expected sms_template_code to be 'mock-template-code', got: %s", req.Form.Get("sms_template_code"))
			}

			if req.Form.Get("rec_num") != "18888888888" {
				t.Errorf("Expected rec_num to be '18888888888', got: %s", req.Form.Get("rec_num"))
			}

			// 验证模板参数
			smsParam := req.Form.Get("sms_param")
			var smsParamData map[string]any
			if err := json.Unmarshal([]byte(smsParam), &smsParamData); err != nil {
				t.Errorf("Failed to parse sms_param: %v", err)
			}

			if code, ok := smsParamData["code"].(string); !ok || code != "123456" {
				t.Errorf("Expected sms_param.code to be '123456', got: %v", smsParamData["code"])
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"alibaba_aliqin_fc_sms_num_send_response": map[string]any{
					"result": map[string]any{
						"err_code": "0",
						"msg":      "mock-result",
						"success":  true,
					},
				},
			})
		})

	// 创建网关配置
	config := map[string]any{
		"app_key":        "mock-app-key",
		"app_secret_key": "mock-app-secret",
		"sign_name":      "mock-app-sign-name",
	}

	// 创建网关
	g := gateway.NewAliyunrestGateway(config)

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

	alibabaResponse, ok := respMap["alibaba_aliqin_fc_sms_num_send_response"].(map[string]any)
	if !ok {
		t.Fatalf("Expected alibaba_aliqin_fc_sms_num_send_response to be map[string]any, got: %T", respMap["alibaba_aliqin_fc_sms_num_send_response"])
	}

	result, ok := alibabaResponse["result"].(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be map[string]any, got: %T", alibabaResponse["result"])
	}

	errCode, ok := result["err_code"].(string)
	if !ok || errCode != "0" {
		t.Errorf("Expected err_code to be '0', got: %v", errCode)
	}

	resultMsg, ok := result["msg"].(string)
	if !ok || resultMsg != "mock-result" {
		t.Errorf("Expected msg to be 'mock-result', got: %v", resultMsg)
	}
}

// TestAliyunrestGatewayError 测试阿里云 REST API 短信网关错误响应
func TestAliyunrestGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", gateway.AliyunrestEndpointURL,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"error_response": map[string]any{
				"code": 15,
				"msg":  "mock-err-msg",
			},
		}))

	// 创建网关配置
	config := map[string]any{
		"app_key":        "mock-app-key",
		"app_secret_key": "mock-app-secret",
		"sign_name":      "mock-app-sign-name",
	}

	// 创建网关
	g := gateway.NewAliyunrestGateway(config)

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
	expectedError := "阿里云 REST API 短信发送失败: [15] mock-err-msg"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', got: '%s'", expectedError, err.Error())
	}
}
