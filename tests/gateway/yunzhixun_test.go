package gateway

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestYunzhixunGateway 测试云之讯短信网关
func TestYunzhixunGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应 - sendsms
	httpmock.RegisterResponder("POST", "https://open.ucpaas.com/ol/sms/sendsms",
		func(req *http.Request) (*http.Response, error) {
			// 验证请求头
			if req.Header.Get("Content-Type") != "application/json;charset=utf-8" {
				t.Errorf("Expected Content-Type to be 'application/json;charset=utf-8', got: %s", req.Header.Get("Content-Type"))
			}

			if req.Header.Get("Accept") != "application/json" {
				t.Errorf("Expected Accept to be 'application/json', got: %s", req.Header.Get("Accept"))
			}

			// 解析请求体
			var params map[string]any
			decoder := json.NewDecoder(req.Body)
			if err := decoder.Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// 验证请求参数
			if params["sid"] != "mock-sid" {
				t.Errorf("Expected sid to be 'mock-sid', got: %v", params["sid"])
			}

			if params["token"] != "mock-token" {
				t.Errorf("Expected token to be 'mock-token', got: %v", params["token"])
			}

			if params["appid"] != "mock-app-id" {
				t.Errorf("Expected appid to be 'mock-app-id', got: %v", params["appid"])
			}

			if params["templateid"] != "mock-template-id" {
				t.Errorf("Expected templateid to be 'mock-template-id', got: %v", params["templateid"])
			}

			if params["mobile"] != "18888888888" {
				t.Errorf("Expected mobile to be '18888888888', got: %v", params["mobile"])
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code": gateway.YunzhixunSuccessCode,
				"msg":  "发送成功",
				"data": map[string]any{
					"count": 1,
					"smsid": "mock-smsid",
				},
			})
		})

	// 创建网关配置
	config := map[string]any{
		"sid":    "mock-sid",
		"token":  "mock-token",
		"app_id": "mock-app-id",
	}

	// 创建网关
	g := gateway.NewYunzhixunGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-id").
		SetData(map[string]any{
			"params": "1234,5",
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

	if respMap["code"].(string) != gateway.YunzhixunSuccessCode {
		t.Errorf("Expected code to be %s, got: %v", gateway.YunzhixunSuccessCode, respMap["code"])
	}

	if respMap["msg"].(string) != "发送成功" {
		t.Errorf("Expected msg to be '发送成功', got: %v", respMap["msg"])
	}

	data, ok := respMap["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data to be map[string]any, got: %T", respMap["data"])
	}

	if data["count"].(float64) != 1 {
		t.Errorf("Expected count to be 1, got: %v", data["count"])
	}

	if data["smsid"].(string) != "mock-smsid" {
		t.Errorf("Expected smsid to be 'mock-smsid', got: %v", data["smsid"])
	}
}

// TestYunzhixunGatewayBatchSend 测试云之讯短信网关批量发送
func TestYunzhixunGatewayBatchSend(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应 - sendsms_batch
	httpmock.RegisterResponder("POST", "https://open.ucpaas.com/ol/sms/sendsms_batch",
		func(req *http.Request) (*http.Response, error) {
			// 解析请求体
			var params map[string]any
			decoder := json.NewDecoder(req.Body)
			if err := decoder.Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// 验证请求参数
			if params["mobile"] != "18888888888,19999999999" {
				t.Errorf("Expected mobile to be '18888888888,19999999999', got: %v", params["mobile"])
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code": gateway.YunzhixunSuccessCode,
				"msg":  "批量发送成功",
				"data": map[string]any{
					"count": 2,
					"smsid": "mock-batch-smsid",
				},
			})
		})

	// 创建网关配置
	config := map[string]any{
		"sid":    "mock-sid",
		"token":  "mock-token",
		"app_id": "mock-app-id",
	}

	// 创建网关
	g := gateway.NewYunzhixunGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-id").
		SetData(map[string]any{
			"params":  "1234,5",
			"mobiles": "18888888888,19999999999",
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

	if respMap["code"].(string) != gateway.YunzhixunSuccessCode {
		t.Errorf("Expected code to be %s, got: %v", gateway.YunzhixunSuccessCode, respMap["code"])
	}

	if respMap["msg"].(string) != "批量发送成功" {
		t.Errorf("Expected msg to be '批量发送成功', got: %v", respMap["msg"])
	}

	data, ok := respMap["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data to be map[string]any, got: %T", respMap["data"])
	}

	if data["count"].(float64) != 2 {
		t.Errorf("Expected count to be 2, got: %v", data["count"])
	}
}

// TestYunzhixunGatewayError 测试云之讯短信网关错误响应
func TestYunzhixunGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", "https://open.ucpaas.com/ol/sms/sendsms",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"code": "100001",
			"msg":  "参数错误",
		}))

	// 创建网关配置
	config := map[string]any{
		"sid":    "mock-sid",
		"token":  "mock-token",
		"app_id": "mock-app-id",
	}

	// 创建网关
	g := gateway.NewYunzhixunGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-id")

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误消息
	expectedError := "yunzhixun gateway error: 参数错误 (code: 100001)"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}

// TestYunzhixunGatewayWithUID 测试云之讯短信网关带 UID
func TestYunzhixunGatewayWithUID(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", "https://open.ucpaas.com/ol/sms/sendsms",
		func(req *http.Request) (*http.Response, error) {
			// 解析请求体
			var params map[string]any
			decoder := json.NewDecoder(req.Body)
			if err := decoder.Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// 验证请求参数
			if params["uid"] != "mock-uid" {
				t.Errorf("Expected uid to be 'mock-uid', got: %v", params["uid"])
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code": gateway.YunzhixunSuccessCode,
				"msg":  "发送成功",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"sid":    "mock-sid",
		"token":  "mock-token",
		"app_id": "mock-app-id",
	}

	// 创建网关
	g := gateway.NewYunzhixunGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template-id").
		SetData(map[string]any{
			"uid": "mock-uid",
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

	if respMap["code"].(string) != gateway.YunzhixunSuccessCode {
		t.Errorf("Expected code to be %s, got: %v", gateway.YunzhixunSuccessCode, respMap["code"])
	}
}
