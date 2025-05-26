package gateway

import (
	"net/http"
	"strings"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// setupYunpianMock sets up httpmock for Yunpian SMS API
func setupYunpianMock() string {
	// Define the base URL
	baseURL := "https://sms.yunpian.com"
	endpoint := baseURL + "/v2/sms/single_send.json"

	// Activate httpmock
	httpmock.Activate()

	// Register the request validator
	httpmock.RegisterResponder("POST", endpoint, func(r *http.Request) (*http.Response, error) {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			return httpmock.NewJsonResponse(http.StatusBadRequest, map[string]any{
				"code": 400,
				"msg":  "Bad Request",
			})
		}

		// Validate required parameters
		apikey := r.PostForm.Get("apikey")
		if apikey == "" {
			return httpmock.NewJsonResponse(http.StatusOK, map[string]any{
				"code": 1,
				"msg":  "非法的apikey",
			})
		}

		mobile := r.PostForm.Get("mobile")
		if mobile == "" {
			return httpmock.NewJsonResponse(http.StatusOK, map[string]any{
				"code": 2,
				"msg":  "手机号码为空",
			})
		}

		text := r.PostForm.Get("text")
		if text == "" {
			return httpmock.NewJsonResponse(http.StatusOK, map[string]any{
				"code": 3,
				"msg":  "短信内容为空",
			})
		}

		// Check if this is a test failure scenario
		if strings.Contains(mobile, "error") {
			return httpmock.NewJsonResponse(http.StatusOK, map[string]any{
				"code": 4,
				"msg":  "无效的手机号码",
			})
		}

		// Return success response
		return httpmock.NewJsonResponse(http.StatusOK, map[string]any{
			"code":   0,
			"msg":    "发送成功",
			"count":  1,
			"fee":    0.05,
			"unit":   "RMB",
			"mobile": mobile,
			"sid":    3310228982,
		})
	})

	return baseURL
}

func TestYunpianGateway(t *testing.T) {
	// 创建一个云片网关
	config := map[string]any{
		"api_key":   "test_api_key",
		"signature": "【测试签名】",
	}

	g := gateway.NewYunpianGateway(config)

	// 测试网关名称
	if g.GetName() != "yunpian" {
		t.Errorf("Expected name to be yunpian, got: %s", g.GetName())
	}

	// 测试配置
	if g.GetConfigString("api_key") != "test_api_key" {
		t.Errorf("Expected api_key to be test_api_key, got: %s", g.GetConfigString("api_key"))
	}

	if g.GetConfigString("signature") != "【测试签名】" {
		t.Errorf("Expected signature to be 【测试签名】, got: %s", g.GetConfigString("signature"))
	}

	// 测试缺少必要配置时的错误
	badConfig := map[string]any{}
	badGateway := gateway.NewYunpianGateway(badConfig)

	phone := message.NewPhoneNumber("13800138000")
	msg := message.NewMessage().
		SetContent("您的验证码是：123456，有效期为5分钟。")

	_, err := badGateway.Send(phone, msg)
	if err == nil {
		t.Errorf("Expected error when missing required config")
	}

	// 测试缺少内容时的错误
	badMsg := message.NewMessage()

	_, err = g.Send(phone, badMsg)
	if err == nil {
		t.Errorf("Expected error when missing content")
	}
}

func TestYunpianGatewayWithMockServer(t *testing.T) {
	// 创建模拟服务器
	baseURL := setupYunpianMock()
	defer httpmock.DeactivateAndReset()

	// 创建一个云片网关，使用模拟服务器的URL
	config := map[string]any{
		"api_key":   "test_api_key",
		"signature": "【测试签名】",
		"endpoint":  baseURL, // 使用模拟服务器的URL
	}

	g := gateway.NewYunpianGateway(config)

	// 测试成功发送短信
	phone := message.NewPhoneNumber("13800138000")
	msg := message.NewMessage().
		SetContent("您的验证码是：123456，有效期为5分钟。")

	resp, err := g.Send(phone, msg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// 验证响应
	respMap, ok := resp.(map[string]any)
	if !ok {
		t.Errorf("Expected response to be map[string]any, got: %T", resp)
	}

	if code, ok := respMap["code"].(float64); !ok || code != 0 {
		t.Errorf("Expected code to be 0, got: %v", code)
	}

	// 测试发送失败 - 无效的手机号码
	errorPhone := message.NewPhoneNumber("error_13800138000")
	resp, err = g.Send(errorPhone, msg)
	if err == nil {
		t.Errorf("Expected error for invalid phone number")
	}

	// 验证错误响应
	respMap, ok = resp.(map[string]any)
	if !ok {
		t.Errorf("Expected response to be map[string]any, got: %T", resp)
	}

	if code, ok := respMap["code"].(float64); !ok || code != 4 {
		t.Errorf("Expected code to be 4, got: %v", code)
	}
}

// TestYunpianGatewayWithCustomResponses tests the Yunpian gateway with custom mock responses
func TestYunpianGatewayWithCustomResponses(t *testing.T) {
	// 设置 httpmock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 定义基础 URL
	baseURL := "https://sms.yunpian.com"
	endpoint := baseURL + "/v2/sms/single_send.json"

	// 添加自定义响应
	httpmock.RegisterResponder("POST", endpoint,
		httpmock.NewJsonResponderOrPanic(http.StatusOK, map[string]any{
			"code":   0,
			"msg":    "Custom success response",
			"count":  2,
			"fee":    0.1,
			"unit":   "USD",
			"mobile": "13800138000",
			"sid":    12345678,
		}))

	// 创建一个云片网关，使用模拟服务器的URL
	config := map[string]any{
		"api_key":   "test_api_key",
		"signature": "【测试签名】",
		"endpoint":  baseURL, // 使用模拟服务器的URL
	}

	g := gateway.NewYunpianGateway(config)

	// 测试成功发送短信
	phone := message.NewPhoneNumber("13800138000")
	msg := message.NewMessage().
		SetContent("您的验证码是：123456，有效期为5分钟。")

	resp, err := g.Send(phone, msg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// 验证响应
	respMap, ok := resp.(map[string]any)
	if !ok {
		t.Errorf("Expected response to be map[string]any, got: %T", resp)
	}

	if code, ok := respMap["code"].(float64); !ok || code != 0 {
		t.Errorf("Expected code to be 0, got: %v", code)
	}

	if message, ok := respMap["msg"].(string); !ok || message != "Custom success response" {
		t.Errorf("Expected msg to be 'Custom success response', got: %v", message)
	}
}
