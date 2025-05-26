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

// TestQiniuGateway 测试七牛云短信网关
func TestQiniuGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", `=~^https://sms\.qiniuapi\.com/v1/message/single`,
		func(req *http.Request) (*http.Response, error) {
			// 解析请求体
			var params map[string]any
			if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// 验证请求参数
			if templateID, ok := params["template_id"].(string); !ok || templateID != "mock-tpl-id" {
				t.Errorf("Expected template_id to be 'mock-tpl-id', got: %v", params["template_id"])
			}

			if mobile, ok := params["mobile"].(string); !ok || mobile != "18888888888" {
				t.Errorf("Expected mobile to be '18888888888', got: %v", params["mobile"])
			}

			// 验证模板参数
			parameters, ok := params["parameters"].(map[string]any)
			if !ok {
				t.Errorf("Expected parameters to be map[string]any, got: %T", params["parameters"])
			}

			if code, ok := parameters["code"].(string); !ok || code != "1234" {
				t.Errorf("Expected parameters.code to be '1234', got: %v", parameters["code"])
			}

			// 验证请求头
			if contentType := req.Header.Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Expected Content-Type to be 'application/json', got: %s", contentType)
			}

			if authorization := req.Header.Get("Authorization"); !strings.HasPrefix(authorization, "Qiniu mock-access-key:") {
				t.Errorf("Expected Authorization to start with 'Qiniu mock-access-key:', got: %s", authorization)
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"message_id": "21321974632178",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"access_key": "mock-access-key",
		"secret_key": "mock-secret-key",
	}

	// 创建网关
	g := gateway.NewQiniuGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-tpl-id").
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

	messageID, ok := respMap["message_id"].(string)
	if !ok || messageID != "21321974632178" {
		t.Errorf("Expected message_id to be '21321974632178', got: %v", messageID)
	}
}

// TestQiniuGatewayError 测试七牛云短信网关错误响应
func TestQiniuGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", `=~^https://sms\.qiniuapi\.com/v1/message/single`,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"error":      "BadToken",
			"message":    "Your authorization token is invalid",
			"request_id": "VEc9f6W1guxye94V",
		}))

	// 创建网关配置
	config := map[string]any{
		"access_key": "mock-access-key",
		"secret_key": "mock-secret-key",
	}

	// 创建网关
	g := gateway.NewQiniuGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-tpl-id").
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

	// 验证错误信息
	expectedError := "七牛云短信发送失败: Your authorization token is invalid"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
