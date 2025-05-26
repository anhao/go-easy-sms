package gateway

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestModuyunGateway 测试摩杜云短信网关
func TestModuyunGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", `=~^https://live\.moduyun\.com/sms/v2/sendsinglesms`, func(req *http.Request) (*http.Response, error) {
		// 解析请求体
		var params map[string]any
		if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// 验证请求参数
		tel, ok := params["tel"].(map[string]any)
		if !ok {
			t.Errorf("Expected tel to be map[string]any, got: %T", params["tel"])
		}

		mobile, ok := tel["mobile"].(string)
		if !ok || mobile != "18888888888" {
			t.Errorf("Expected tel.mobile to be '18888888888', got: %v", tel["mobile"])
		}

		nationcode, ok := tel["nationcode"].(string)
		if !ok || nationcode != "86" {
			t.Errorf("Expected tel.nationcode to be '86', got: %v", tel["nationcode"])
		}

		signId, ok := params["signId"].(string)
		if !ok || signId != "mock-signId" {
			t.Errorf("Expected signId to be 'mock-signId', got: %v", params["signId"])
		}

		templateId, ok := params["templateId"].(string)
		if !ok || templateId != "mock-template" {
			t.Errorf("Expected templateId to be 'mock-template', got: %v", params["templateId"])
		}

		// 验证模板参数
		templateParams, ok := params["params"].([]any)
		if !ok {
			t.Errorf("Expected params to be []any, got: %T", params["params"])
		}

		if len(templateParams) != 1 {
			t.Errorf("Expected params to have 1 element, got: %d", len(templateParams))
		}

		if templateParams[0] != "1234" {
			t.Errorf("Expected params[0] to be '1234', got: %v", templateParams[0])
		}

		// 验证签名
		sig, ok := params["sig"].(string)
		if !ok {
			t.Errorf("Expected sig to be string, got: %T", params["sig"])
		}

		// 获取 URL 参数
		query := req.URL.Query()
		accesskey := query.Get("accesskey")
		random := query.Get("random")
		randomInt, _ := strconv.Atoi(random)

		// 验证 URL 参数
		if accesskey != "mock-accesskey" {
			t.Errorf("Expected accesskey to be 'mock-accesskey', got: %s", accesskey)
		}

		// 验证签名
		timestamp, ok := params["time"].(float64)
		if !ok {
			t.Errorf("Expected time to be float64, got: %T", params["time"])
		}

		signStr := fmt.Sprintf("secretkey=%s&random=%d&time=%d&mobile=%s",
			"mock-secretkey",
			randomInt,
			int64(timestamp),
			mobile)

		h := sha256.New()
		h.Write([]byte(signStr))
		expectedSign := fmt.Sprintf("%x", h.Sum(nil))

		if sig != expectedSign {
			t.Errorf("Expected sig to be '%s', got: '%s'", expectedSign, sig)
		}

		// 返回成功响应
		return httpmock.NewJsonResponse(200, map[string]any{
			"result":  0,
			"errmsg":  "OK",
			"ext":     "",
			"sid":     "mock-sid",
			"surplus": 4,
			"balance": 0,
		})
	})

	// 创建网关配置
	config := map[string]any{
		"accesskey": "mock-accesskey",
		"secretkey": "mock-secretkey",
		"signId":    "mock-signId",
		"type":      0,
	}

	// 创建网关
	g := gateway.NewModuyunGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template").
		SetData(map[string]any{
			"0": "1234",
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

	result, ok := respMap["result"].(float64)
	if !ok || int(result) != 0 {
		t.Errorf("Expected result to be 0, got: %v", result)
	}

	errmsg, ok := respMap["errmsg"].(string)
	if !ok || errmsg != "OK" {
		t.Errorf("Expected errmsg to be 'OK', got: %v", errmsg)
	}

	sid, ok := respMap["sid"].(string)
	if !ok || sid != "mock-sid" {
		t.Errorf("Expected sid to be 'mock-sid', got: %v", sid)
	}
}

// TestModuyunGatewayError 测试摩杜云短信网关错误响应
func TestModuyunGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", `=~^https://live\.moduyun\.com/sms/v2/sendsinglesms`, httpmock.NewJsonResponderOrPanic(200, map[string]any{
		"result": 1001,
		"errmsg": "accesskey not exist.",
	}))

	// 创建网关配置
	config := map[string]any{
		"accesskey": "mock-accesskey",
		"secretkey": "mock-secretkey",
		"signId":    "mock-signId",
		"type":      0,
	}

	// 创建网关
	g := gateway.NewModuyunGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template").
		SetData(map[string]any{
			"0": "1234",
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误信息
	expectedError := "摩杜云短信发送失败: [1001] accesskey not exist."
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
