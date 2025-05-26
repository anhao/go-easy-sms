package gateway

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestRongheyunGateway 测试融合云短信网关
func TestRongheyunGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", gateway.RongheyunEndpointURL,
		func(req *http.Request) (*http.Response, error) {
			// 解析请求体
			var params map[string]any
			if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// 验证请求参数
			if username, ok := params["username"].(string); !ok || username != "mock-username" {
				t.Errorf("Expected username to be 'mock-username', got: %v", params["username"])
			}

			if signature, ok := params["signature"].(string); !ok || signature != "mock-signature" {
				t.Errorf("Expected signature to be 'mock-signature', got: %v", params["signature"])
			}

			if tpId, ok := params["tpId"].(string); !ok || tpId != "mock-template" {
				t.Errorf("Expected tpId to be 'mock-template', got: %v", params["tpId"])
			}

			// 验证密码
			password, ok := params["password"].(string)
			if !ok {
				t.Errorf("Expected password to be string, got: %T", params["password"])
			}

			tKey, ok := params["tKey"].(float64)
			if !ok {
				t.Errorf("Expected tKey to be float64, got: %T", params["tKey"])
			}

			// 验证密码生成
			h1 := md5.New()
			h1.Write([]byte("mock-password"))
			firstMD5 := fmt.Sprintf("%x", h1.Sum(nil))

			h2 := md5.New()
			h2.Write([]byte(firstMD5 + fmt.Sprintf("%d", int64(tKey))))
			expectedPassword := fmt.Sprintf("%x", h2.Sum(nil))

			if password != expectedPassword {
				t.Errorf("Expected password to be '%s', got: '%s'", expectedPassword, password)
			}

			// 验证 records
			records, ok := params["records"].([]any)
			if !ok {
				t.Errorf("Expected records to be []any, got: %T", params["records"])
			}

			if len(records) != 1 {
				t.Errorf("Expected records to have 1 element, got: %d", len(records))
			}

			record, ok := records[0].(map[string]any)
			if !ok {
				t.Errorf("Expected record to be map[string]any, got: %T", records[0])
			}

			if mobile, ok := record["mobile"].(string); !ok || mobile != "18888888888" {
				t.Errorf("Expected mobile to be '18888888888', got: %v", record["mobile"])
			}

			tpContent, ok := record["tpContent"].(map[string]any)
			if !ok {
				t.Errorf("Expected tpContent to be map[string]any, got: %T", record["tpContent"])
			}

			if validCode, ok := tpContent["valid_code"].(string); !ok || validCode != "888888" {
				t.Errorf("Expected valid_code to be '888888', got: %v", tpContent["valid_code"])
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"code":        gateway.RongheyunSuccessCode,
				"msg":         "success",
				"tpId":        "31874",
				"msgId":       "161553136878837480961",
				"invalidList": []any{},
			})
		})

	// 创建网关配置
	config := map[string]any{
		"username":  "mock-username",
		"password":  "mock-password",
		"signature": "mock-signature",
	}

	// 创建网关
	g := gateway.NewRongheyunGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template").
		SetData(map[string]any{
			"valid_code": "888888",
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

	code, ok := respMap["code"].(float64)
	if !ok || int(code) != gateway.RongheyunSuccessCode {
		t.Errorf("Expected code to be %d, got: %v", gateway.RongheyunSuccessCode, code)
	}

	respMsg, ok := respMap["msg"].(string)
	if !ok || respMsg != "success" {
		t.Errorf("Expected msg to be 'success', got: %v", respMsg)
	}

	tpId, ok := respMap["tpId"].(string)
	if !ok || tpId != "31874" {
		t.Errorf("Expected tpId to be '31874', got: %v", tpId)
	}

	msgId, ok := respMap["msgId"].(string)
	if !ok || msgId != "161553136878837480961" {
		t.Errorf("Expected msgId to be '161553136878837480961', got: %v", msgId)
	}
}

// TestRongheyunGatewayError 测试融合云短信网关错误响应
func TestRongheyunGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", gateway.RongheyunEndpointURL,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"code":        4025,
			"msg":         "template records null",
			"tpId":        "31874",
			"msgId":       "161553131051357039361",
			"invalidList": nil,
		}))

	// 创建网关配置
	config := map[string]any{
		"username":  "mock-username",
		"password":  "mock-password",
		"signature": "mock-signature",
	}

	// 创建网关
	g := gateway.NewRongheyunGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("mock-template").
		SetData(map[string]any{
			"valid_code": "888888",
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误信息
	expectedError := "融合云短信发送失败: [4025] template records null"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', got: '%s'", expectedError, err.Error())
	}
}
