package gateway

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestJuheGateway 测试聚合数据短信网关
func TestJuheGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("GET", gateway.JuheEndpointURL,
		func(req *http.Request) (*http.Response, error) {
			// 验证请求参数
			query := req.URL.Query()

			if query.Get("mobile") != "18888888888" {
				t.Errorf("Expected mobile to be '18888888888', got: %s", query.Get("mobile"))
			}

			if query.Get("tpl_id") != "mock-tpl-id" {
				t.Errorf("Expected tpl_id to be 'mock-tpl-id', got: %s", query.Get("tpl_id"))
			}

			if query.Get("key") != "mock-app-key" {
				t.Errorf("Expected key to be 'mock-app-key', got: %s", query.Get("key"))
			}

			if query.Get("dtype") != gateway.JuheEndpointFormat {
				t.Errorf("Expected dtype to be '%s', got: %s", gateway.JuheEndpointFormat, query.Get("dtype"))
			}

			// 验证模板变量
			tplValue := query.Get("tpl_value")
			parsedTplValue, err := url.ParseQuery(tplValue)
			if err != nil {
				t.Errorf("Failed to parse tpl_value: %v", err)
			}

			if parsedTplValue.Get("#code#") != "1234" {
				t.Errorf("Expected tpl_value to contain '#code#=1234', got: %s", tplValue)
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"reason":     "操作成功",
				"error_code": 0,
			})
		})

	// 创建网关配置
	config := map[string]any{
		"app_key": "mock-app-key",
	}

	// 创建网关
	g := gateway.NewJuheGateway(config)

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

	reason, ok := respMap["reason"].(string)
	if !ok || reason != "操作成功" {
		t.Errorf("Expected reason to be '操作成功', got: %v", reason)
	}

	errorCode, ok := respMap["error_code"].(float64)
	if !ok || int(errorCode) != 0 {
		t.Errorf("Expected error_code to be 0, got: %v", errorCode)
	}
}

// TestJuheGatewayError 测试聚合数据短信网关错误响应
func TestJuheGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("GET", gateway.JuheEndpointURL,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"reason":     "操作失败",
			"error_code": 21000,
		}))

	// 创建网关配置
	config := map[string]any{
		"app_key": "mock-app-key",
	}

	// 创建网关
	g := gateway.NewJuheGateway(config)

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
	expectedError := "聚合数据短信发送失败: [21000] 操作失败"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
