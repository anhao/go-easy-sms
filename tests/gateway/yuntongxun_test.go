package gateway

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestYuntongxunGateway 测试容联云通讯短信网关
func TestYuntongxunGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 获取当前时间
	datetime := time.Now().Format("20060102150405")

	// 生成签名
	h := md5.New()
	h.Write([]byte("mock-account-sid" + "mock-account-token" + datetime))
	sig := strings.ToUpper(fmt.Sprintf("%x", h.Sum(nil)))

	// 构建请求地址 - 仅用于参考
	url := fmt.Sprintf(gateway.YuntongxunEndpointTemplate,
		gateway.YuntongxunServerIP,
		gateway.YuntongxunServerPort,
		gateway.YuntongxunSDKVersion,
		"Accounts",
		"mock-account-sid",
		"SMS",
		"TemplateSMS",
		sig,
	)

	// 注册成功响应
	httpmock.RegisterResponder("POST", url,
		func(req *http.Request) (*http.Response, error) {
			// 解析请求体
			var params map[string]any
			if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// 验证请求参数
			if appID, ok := params["appId"].(string); !ok || appID != "mock-app-id" {
				t.Errorf("Expected appId to be 'mock-app-id', got: %v", params["appId"])
			}

			if to, ok := params["to"].(string); !ok || to != "18888888888" {
				t.Errorf("Expected to to be '18888888888', got: %v", params["to"])
			}

			if templateID, ok := params["templateId"].(float64); !ok || int(templateID) != 5589 {
				t.Errorf("Expected templateId to be 5589, got: %v", params["templateId"])
			}

			// 验证请求头
			if accept := req.Header.Get("Accept"); accept != "application/json" {
				t.Errorf("Expected Accept to be 'application/json', got: %s", accept)
			}

			if contentType := req.Header.Get("Content-Type"); contentType != "application/json;charset=utf-8" {
				t.Errorf("Expected Content-Type to be 'application/json;charset=utf-8', got: %s", contentType)
			}

			// 验证 Authorization 头
			expectedAuth := base64.StdEncoding.EncodeToString([]byte("mock-account-sid:" + datetime))
			if auth := req.Header.Get("Authorization"); auth != expectedAuth {
				t.Errorf("Expected Authorization to be '%s', got: '%s'", expectedAuth, auth)
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"statusCode": gateway.YuntongxunSuccessCode,
				"statusMsg":  "Success",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"debug":          false,
		"is_sub_account": false,
		"account_sid":    "mock-account-sid",
		"account_token":  "mock-account-token",
		"app_id":         "mock-app-id",
	}

	// 创建网关
	g := gateway.NewYuntongxunGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("5589").
		SetData(map[string]any{
			"0": "mock-data-1",
			"1": "mock-data-2",
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

	statusCode, ok := respMap["statusCode"].(string)
	if !ok || statusCode != gateway.YuntongxunSuccessCode {
		t.Errorf("Expected statusCode to be '%s', got: %v", gateway.YuntongxunSuccessCode, statusCode)
	}
}

// TestYuntongxunGatewayIntl 测试容联云通讯国际短信网关
func TestYuntongxunGatewayIntl(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 获取当前时间
	datetime := time.Now().Format("20060102150405")

	// 生成签名
	h := md5.New()
	h.Write([]byte("mock-account-sid" + "mock-account-token" + datetime))
	sig := strings.ToUpper(fmt.Sprintf("%x", h.Sum(nil)))

	// 构建请求地址 - 仅用于参考
	url := fmt.Sprintf(gateway.YuntongxunEndpointTemplate,
		gateway.YuntongxunServerIP,
		gateway.YuntongxunServerPort,
		gateway.YuntongxunSDKVersionInt,
		"account",
		"mock-account-sid",
		"international",
		"send",
		sig,
	)

	// 注册成功响应
	httpmock.RegisterResponder("POST", url,
		func(req *http.Request) (*http.Response, error) {
			// 解析请求体
			var params map[string]any
			if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// 验证请求参数
			if appID, ok := params["appId"].(string); !ok || appID != "mock-app-id" {
				t.Errorf("Expected appId to be 'mock-app-id', got: %v", params["appId"])
			}

			if mobile, ok := params["mobile"].(string); !ok || mobile != "006018888888888" {
				t.Errorf("Expected mobile to be '006018888888888', got: %v", params["mobile"])
			}

			if content, ok := params["content"].(string); !ok || content != "This is a test message." {
				t.Errorf("Expected content to be 'This is a test message.', got: %v", params["content"])
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"statusCode": gateway.YuntongxunSuccessCode,
				"statusMsg":  "Success",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"debug":          false,
		"is_sub_account": false,
		"account_sid":    "mock-account-sid",
		"account_token":  "mock-account-token",
		"app_id":         "mock-app-id",
	}

	// 创建网关
	g := gateway.NewYuntongxunGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetContent("This is a test message.")

	// 创建电话号码（国际号码）
	phone := message.NewPhoneNumber("18888888888", 60)

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

	statusCode, ok := respMap["statusCode"].(string)
	if !ok || statusCode != gateway.YuntongxunSuccessCode {
		t.Errorf("Expected statusCode to be '%s', got: %v", gateway.YuntongxunSuccessCode, statusCode)
	}
}

// TestYuntongxunGatewayError 测试容联云通讯短信网关错误响应
func TestYuntongxunGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 获取当前时间
	datetime := time.Now().Format("20060102150405")

	// 生成签名
	h := md5.New()
	h.Write([]byte("mock-account-sid" + "mock-account-token" + datetime))
	sig := strings.ToUpper(fmt.Sprintf("%x", h.Sum(nil)))

	// 构建请求地址 - 仅用于参考
	url := fmt.Sprintf(gateway.YuntongxunEndpointTemplate,
		gateway.YuntongxunServerIP,
		gateway.YuntongxunServerPort,
		gateway.YuntongxunSDKVersion,
		"Accounts",
		"mock-account-sid",
		"SMS",
		"TemplateSMS",
		sig,
	)

	// 注册错误响应
	httpmock.RegisterResponder("POST", url,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"statusCode": "100",
			"statusMsg":  "Error",
		}))

	// 创建网关配置
	config := map[string]any{
		"debug":          false,
		"is_sub_account": false,
		"account_sid":    "mock-account-sid",
		"account_token":  "mock-account-token",
		"app_id":         "mock-app-id",
	}

	// 创建网关
	g := gateway.NewYuntongxunGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("5589").
		SetData(map[string]any{
			"0": "mock-data-1",
			"1": "mock-data-2",
		})

	// 创建电话号码
	phone := message.NewPhoneNumber("18888888888")

	// 测试发送
	_, err := g.Send(phone, msg)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// 验证错误信息
	expectedError := "容联云通讯短信发送失败: 100"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
