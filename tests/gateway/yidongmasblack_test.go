package gateway

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestYidongmasblackGateway 测试移动MAS黑名单模式短信网关
func TestYidongmasblackGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", gateway.YidongmasblackEndpointURL,
		func(req *http.Request) (*http.Response, error) {
			// 验证请求头
			if req.Header.Get("Content-Type") != "application/json;charset=utf-8" {
				t.Errorf("Expected Content-Type to be 'application/json;charset=utf-8', got: %s", req.Header.Get("Content-Type"))
			}

			if req.Header.Get("Accept") != "application/json" {
				t.Errorf("Expected Accept to be 'application/json', got: %s", req.Header.Get("Accept"))
			}

			// 读取请求体
			buf := make([]byte, req.ContentLength)
			_, err := req.Body.Read(buf)
			if err != nil && err.Error() != "EOF" {
				t.Errorf("Failed to read request body: %v", err)
			}

			// 解码 Base64
			decoded, err := base64.StdEncoding.DecodeString(string(buf))
			if err != nil {
				t.Errorf("Failed to decode base64: %v", err)
			}

			// 解析 JSON
			var params map[string]any
			if err := json.Unmarshal(decoded, &params); err != nil {
				t.Errorf("Failed to unmarshal JSON: %v", err)
			}

			// 验证请求参数
			if params["ecName"] != "mock-ec-name" {
				t.Errorf("Expected ecName to be 'mock-ec-name', got: %v", params["ecName"])
			}

			if params["apId"] != "mock-ap-id" {
				t.Errorf("Expected apId to be 'mock-ap-id', got: %v", params["apId"])
			}

			if params["sign"] != "mock-sign" {
				t.Errorf("Expected sign to be 'mock-sign', got: %v", params["sign"])
			}

			if params["addSerial"] != "mock-add-serial" {
				t.Errorf("Expected addSerial to be 'mock-add-serial', got: %v", params["addSerial"])
			}

			if params["mobiles"] != "18888888888" {
				t.Errorf("Expected mobiles to be '18888888888', got: %v", params["mobiles"])
			}

			if params["content"] != "This is a test message." {
				t.Errorf("Expected content to be 'This is a test message.', got: %v", params["content"])
			}

			// 验证 MAC
			h := md5.New()
			_, _ = fmt.Fprintf(h, "%s%s%s%s%s%s%s",
				params["ecName"],
				params["apId"],
				"mock-secret-key",
				params["mobiles"],
				params["content"],
				params["sign"],
				params["addSerial"],
			)
			expectedMac := fmt.Sprintf("%x", h.Sum(nil))

			if params["mac"] != expectedMac {
				t.Errorf("Expected mac to be '%s', got: '%s'", expectedMac, params["mac"])
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"success": gateway.YidongmasblackSuccessStatus,
				"rspcod":  "1234",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"ecName":    "mock-ec-name",
		"secretKey": "mock-secret-key",
		"apId":      "mock-ap-id",
		"sign":      "mock-sign",
		"addSerial": "mock-add-serial",
	}

	// 创建网关
	g := gateway.NewYidongmasblackGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("This is a test message.")

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

	if respMap["success"].(string) != gateway.YidongmasblackSuccessStatus {
		t.Errorf("Expected success to be %s, got: %v", gateway.YidongmasblackSuccessStatus, respMap["success"])
	}

	if respMap["rspcod"].(string) != "1234" {
		t.Errorf("Expected rspcod to be '1234', got: %v", respMap["rspcod"])
	}
}

// TestYidongmasblackGatewayError 测试移动MAS黑名单模式短信网关错误响应
func TestYidongmasblackGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", gateway.YidongmasblackEndpointURL,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"success": "mock-err-msg",
			"rspcod":  "1234",
		}))

	// 创建网关配置
	config := map[string]any{
		"ecName":    "mock-ec-name",
		"secretKey": "mock-secret-key",
		"apId":      "mock-ap-id",
		"sign":      "mock-sign",
		"addSerial": "mock-add-serial",
	}

	// 创建网关
	g := gateway.NewYidongmasblackGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("This is a test message.")

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误消息
	expectedError := "yidongmasblack gateway error: mock-err-msg (code: 1234)"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestYidongmasblackGatewayGenerateContent 测试移动MAS黑名单模式短信网关生成内容
func TestYidongmasblackGatewayGenerateContent(t *testing.T) {
	// 创建网关配置
	config := map[string]any{
		"ecName":    "mock-ec-name",
		"secretKey": "mock-secret-key",
		"apId":      "mock-ap-id",
		"sign":      "mock-sign",
		"addSerial": "mock-add-serial",
	}

	// 创建网关配置
	_ = config

	// 构建参数
	params := map[string]any{
		"ecName":    "mock-ec-name",
		"apId":      "mock-ap-id",
		"sign":      "mock-sign",
		"addSerial": "mock-add-serial",
		"mobiles":   "18888888888",
		"content":   "This is a test message.",
	}

	// 创建一个新的网关实例，直接使用 GenerateContent 方法
	newGateway := gateway.NewYidongmasblackGateway(config)
	// 生成内容
	content := newGateway.GenerateContent(params)

	// 解码 Base64
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	// 解析 JSON
	var decodedParams map[string]any
	if err := json.Unmarshal(decoded, &decodedParams); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// 验证参数
	if decodedParams["ecName"] != "mock-ec-name" {
		t.Errorf("Expected ecName to be 'mock-ec-name', got: %v", decodedParams["ecName"])
	}

	if decodedParams["apId"] != "mock-ap-id" {
		t.Errorf("Expected apId to be 'mock-ap-id', got: %v", decodedParams["apId"])
	}

	// 验证 MAC
	h := md5.New()
	_, _ = fmt.Fprintf(h, "%s%s%s%s%s%s%s",
		params["ecName"],
		params["apId"],
		"mock-secret-key",
		params["mobiles"],
		params["content"],
		params["sign"],
		params["addSerial"],
	)
	expectedMac := fmt.Sprintf("%x", h.Sum(nil))

	if decodedParams["mac"] != expectedMac {
		t.Errorf("Expected mac to be '%s', got: '%s'", expectedMac, decodedParams["mac"])
	}
}
