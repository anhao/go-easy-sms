package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// verifyFeigeSign 根据请求头校验签名是否正确
func verifyFeigeSign(t *testing.T, req *http.Request, apiKey, secret string) {
	t.Helper()

	if got := req.Header.Get("X-Api-Key"); got != apiKey {
		t.Errorf("Expected X-Api-Key to be '%s', got: '%s'", apiKey, got)
	}

	timestamp := req.Header.Get("X-Timestamp")
	if _, err := strconv.ParseInt(timestamp, 10, 64); err != nil {
		t.Errorf("Expected X-Timestamp to be a unix timestamp, got: '%s'", timestamp)
	}

	nonce := req.Header.Get("X-Nonce")
	if nonce == "" {
		t.Errorf("Expected X-Nonce to be present")
	}
	if len(nonce) > 40 {
		t.Errorf("Expected X-Nonce length <= 40, got: %d", len(nonce))
	}

	raw := fmt.Sprintf("%s%s%s%s", apiKey, secret, timestamp, nonce)
	sum := sha256.Sum256([]byte(raw))
	expectedSign := hex.EncodeToString(sum[:])

	if got := req.Header.Get("X-Sign"); got != expectedSign {
		t.Errorf("Expected X-Sign to be '%s', got: '%s'", expectedSign, got)
	}
}

// TestFeigeGatewayTemplate 测试飞鸽云模板短信发送
func TestFeigeGatewayTemplate(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", gateway.FeigeEndpointTemplate,
		func(req *http.Request) (*http.Response, error) {
			// 校验签名与请求头
			verifyFeigeSign(t, req, "mock-api-key", "mock-api-secret")

			// 解析请求体
			var payload map[string]any
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// 模板 ID 应为数字类型
			if templateID, ok := payload["template_id"].(float64); !ok || int(templateID) != 212 {
				t.Errorf("Expected template_id to be 212, got: %v", payload["template_id"])
			}

			// data 中的 sms_sign_id 应覆盖配置中的值
			if signID, ok := payload["sms_sign_id"].(float64); !ok || int(signID) != 5 {
				t.Errorf("Expected sms_sign_id to be 5, got: %v", payload["sms_sign_id"])
			}

			// user_data 应原样带上
			if userData, ok := payload["user_data"].(string); !ok || userData != "biz-123" {
				t.Errorf("Expected user_data to be 'biz-123', got: %v", payload["user_data"])
			}

			// 校验 messages
			messages, ok := payload["messages"].([]any)
			if !ok || len(messages) != 1 {
				t.Fatalf("Expected messages to have 1 element, got: %v", payload["messages"])
			}

			record, ok := messages[0].(map[string]any)
			if !ok {
				t.Fatalf("Expected message to be map[string]any, got: %T", messages[0])
			}

			if phone, ok := record["phone_number"].(string); !ok || phone != "18190678838" {
				t.Errorf("Expected phone_number to be '18190678838', got: %v", record["phone_number"])
			}

			params, ok := record["params"].(map[string]any)
			if !ok {
				t.Fatalf("Expected params to be map[string]any, got: %T", record["params"])
			}

			if name, ok := params["name"].(string); !ok || name != "12345678" {
				t.Errorf("Expected params.name to be '12345678', got: %v", params["name"])
			}

			// 保留字段不应出现在模板变量中
			if _, ok := params["sms_sign_id"]; ok {
				t.Errorf("Expected params to not contain sms_sign_id")
			}
			if _, ok := params["user_data"]; ok {
				t.Errorf("Expected params to not contain user_data")
			}

			return httpmock.NewJsonResponse(200, map[string]any{
				"code":     0,
				"msg":      "OK",
				"count":    1,
				"batch_id": "18918743061041152",
			})
		})

	config := map[string]any{
		"api_key":     "mock-api-key",
		"api_secret":  "mock-api-secret",
		"sms_sign_id": 3,
	}

	g := gateway.NewFeigeGateway(config)

	msg := message.NewMessage().
		SetTemplate("212").
		SetData(map[string]any{
			"name":        "12345678",
			"sms_sign_id": 5,
			"user_data":   "biz-123",
		})

	phone := message.NewPhoneNumber("18190678838")

	resp, err := g.Send(phone, msg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	respMap, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("Expected response to be map[string]any, got: %T", resp)
	}

	if code, ok := respMap["code"].(float64); !ok || int(code) != gateway.FeigeSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.FeigeSuccessCode, respMap["code"])
	}

	if batchID, ok := respMap["batch_id"].(string); !ok || batchID != "18918743061041152" {
		t.Errorf("Expected batch_id to be '18918743061041152', got: %v", respMap["batch_id"])
	}
}

// TestFeigeGatewayContent 测试飞鸽云自定义短信发送
func TestFeigeGatewayContent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", gateway.FeigeEndpointCustom,
		func(req *http.Request) (*http.Response, error) {
			verifyFeigeSign(t, req, "mock-api-key", "mock-api-secret")

			var payload map[string]any
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if content, ok := payload["content"].(string); !ok || content != "您的验证码为: 6379，有效期为5分钟。" {
				t.Errorf("Expected content to match, got: %v", payload["content"])
			}

			// 未在 data 中指定时，签名 ID 取自配置
			if signID, ok := payload["sms_sign_id"].(float64); !ok || int(signID) != 3 {
				t.Errorf("Expected sms_sign_id to be 3, got: %v", payload["sms_sign_id"])
			}

			phoneNumbers, ok := payload["phone_numbers"].([]any)
			if !ok || len(phoneNumbers) != 1 {
				t.Fatalf("Expected phone_numbers to have 1 element, got: %v", payload["phone_numbers"])
			}

			if phone, ok := phoneNumbers[0].(string); !ok || phone != "13035678838" {
				t.Errorf("Expected phone_numbers[0] to be '13035678838', got: %v", phoneNumbers[0])
			}

			return httpmock.NewJsonResponse(200, map[string]any{
				"code":     0,
				"msg":      "OK",
				"count":    1,
				"batch_id": "18918743061041152",
			})
		})

	config := map[string]any{
		"api_key":     "mock-api-key",
		"api_secret":  "mock-api-secret",
		"sms_sign_id": 3,
	}

	g := gateway.NewFeigeGateway(config)

	msg := message.NewMessage().SetContent("您的验证码为: 6379，有效期为5分钟。")
	phone := message.NewPhoneNumber("13035678838")

	resp, err := g.Send(phone, msg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	respMap, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("Expected response to be map[string]any, got: %T", resp)
	}

	if code, ok := respMap["code"].(float64); !ok || int(code) != gateway.FeigeSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.FeigeSuccessCode, respMap["code"])
	}
}

// TestFeigeGatewayError 测试飞鸽云短信错误响应
func TestFeigeGatewayError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", gateway.FeigeEndpointTemplate,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"code": 1020,
			"msg":  "余额不足",
		}))

	config := map[string]any{
		"api_key":     "mock-api-key",
		"api_secret":  "mock-api-secret",
		"sms_sign_id": 3,
	}

	g := gateway.NewFeigeGateway(config)

	msg := message.NewMessage().
		SetTemplate("212").
		SetData(map[string]any{"name": "12345678"})

	phone := message.NewPhoneNumber("18190678838")

	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	expectedError := "飞鸽云短信发送失败: [1020] 余额不足"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestFeigeGatewayInvalidTemplate 测试非数字模板 ID 时返回错误
func TestFeigeGatewayInvalidTemplate(t *testing.T) {
	config := map[string]any{
		"api_key":    "mock-api-key",
		"api_secret": "mock-api-secret",
	}

	g := gateway.NewFeigeGateway(config)

	msg := message.NewMessage().
		SetTemplate("SMS_001").
		SetData(map[string]any{"name": "12345678"})

	phone := message.NewPhoneNumber("18190678838")

	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error for non-numeric template id, got nil")
	}

	if !strings.Contains(err.Error(), "飞鸽云短信模板ID必须为数字") {
		t.Errorf("Expected template id error, got: '%s'", err.Error())
	}
}
