package gateway

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestSendcloudGateway 测试 SendCloud 短信网关
func TestSendcloudGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", fmt.Sprintf(gateway.SendcloudEndpointTemplate, "send"),
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			if req.Form.Get("smsUser") != "mock-sms-user" {
				t.Errorf("Expected smsUser to be 'mock-sms-user', got: %s", req.Form.Get("smsUser"))
			}

			if req.Form.Get("templateId") != "mock-template" {
				t.Errorf("Expected templateId to be 'mock-template', got: %s", req.Form.Get("templateId"))
			}

			if req.Form.Get("msgType") != "0" {
				t.Errorf("Expected msgType to be '0', got: %s", req.Form.Get("msgType"))
			}

			if req.Form.Get("phone") != "18888888888" {
				t.Errorf("Expected phone to be '18888888888', got: %s", req.Form.Get("phone"))
			}

			// 验证模板变量
			vars := req.Form.Get("vars")
			var varsMap map[string]any
			if err := json.Unmarshal([]byte(vars), &varsMap); err != nil {
				t.Errorf("Failed to parse vars: %v", err)
			}

			if code, ok := varsMap["%code%"].(string); !ok || code != "1234" {
				t.Errorf("Expected vars to contain '%%code%%': '1234', got: %v", varsMap)
			}

			// 验证签名
			signature := req.Form.Get("signature")
			if signature == "" {
				t.Errorf("Expected signature to be set")
			}

			// 构建签名字符串
			params := map[string]string{
				"smsUser":    "mock-sms-user",
				"templateId": "mock-template",
				"msgType":    "0",
				"phone":      "18888888888",
				"vars":       vars,
			}

			// 按照键名排序
			keys := make([]string, 0, len(params))
			for k := range params {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			// 构建查询字符串
			var queryString strings.Builder
			for i, k := range keys {
				if i > 0 {
					queryString.WriteString("&")
				}
				queryString.WriteString(fmt.Sprintf("%s=%s", k, params[k]))
			}

			// 构建签名字符串
			smsKey := "mock-sms-key"
			signStr := fmt.Sprintf("%s&%s&%s", smsKey, queryString.String(), smsKey)

			// 计算 MD5 哈希
			h := md5.New()
			h.Write([]byte(signStr))
			expectedSignature := fmt.Sprintf("%x", h.Sum(nil))

			if signature != expectedSignature {
				t.Errorf("Expected signature to be '%s', got: '%s'", expectedSignature, signature)
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"message":    "操作成功",
				"result":     true,
				"statusCode": 200,
			})
		})

	// 创建网关配置
	config := map[string]any{
		"sms_user": "mock-sms-user",
		"sms_key":  "mock-sms-key",
	}

	// 创建网关
	g := gateway.NewSendcloudGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template").
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

	result, ok := respMap["result"].(bool)
	if !ok || !result {
		t.Errorf("Expected result to be true, got: %v", result)
	}

	message, ok := respMap["message"].(string)
	if !ok || message != "操作成功" {
		t.Errorf("Expected message to be '操作成功', got: %v", message)
	}

	statusCode, ok := respMap["statusCode"].(float64)
	if !ok || int(statusCode) != 200 {
		t.Errorf("Expected statusCode to be 200, got: %v", statusCode)
	}
}

// TestSendcloudGatewayWithTimestamp 测试 SendCloud 短信网关带时间戳
func TestSendcloudGatewayWithTimestamp(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", fmt.Sprintf(gateway.SendcloudEndpointTemplate, "send"),
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			err := req.ParseForm()
			if err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}

			// 验证时间戳
			timestamp := req.Form.Get("timestamp")
			if timestamp == "" {
				t.Errorf("Expected timestamp to be set")
			}

			// 验证时间戳长度
			if len(timestamp) != 13 {
				t.Errorf("Expected timestamp to be 13 digits, got: %s", timestamp)
			}

			// 验证时间戳值
			timestampInt := int64(0)
			_, _ = fmt.Sscanf(timestamp, "%d", &timestampInt)
			if timestampInt > time.Now().UnixMilli() {
				t.Errorf("Expected timestamp to be less than current time, got: %s", timestamp)
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"message":    "操作成功",
				"result":     true,
				"statusCode": 200,
			})
		})

	// 创建网关配置
	config := map[string]any{
		"sms_user":  "mock-sms-user",
		"sms_key":   "mock-sms-key",
		"timestamp": true,
	}

	// 创建网关
	g := gateway.NewSendcloudGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template").
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

	result, ok := respMap["result"].(bool)
	if !ok || !result {
		t.Errorf("Expected result to be true, got: %v", result)
	}
}

// TestSendcloudGatewayError 测试 SendCloud 短信网关错误响应
func TestSendcloudGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", fmt.Sprintf(gateway.SendcloudEndpointTemplate, "send"),
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"message":    "手机号不存在",
			"result":     false,
			"statusCode": 400,
		}))

	// 创建网关配置
	config := map[string]any{
		"sms_user": "mock-sms-user",
		"sms_key":  "mock-sms-key",
	}

	// 创建网关
	g := gateway.NewSendcloudGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template").
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
	expectedError := "SendCloud 短信发送失败: [400] 手机号不存在"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
