package gateway

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// TestMaapGateway 测试 MAAP 短信网关
func TestMaapGateway(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册成功响应
	httpmock.RegisterResponder("POST", gateway.MaapEndpointURL,
		func(req *http.Request) (*http.Response, error) {
			// 解析请求体
			var params map[string]any
			if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// 验证请求参数
			if cpcode, ok := params["cpcode"].(string); !ok || cpcode != "mock-cpcode" {
				t.Errorf("Expected cpcode to be 'mock-cpcode', got: %v", params["cpcode"])
			}

			if msg, ok := params["msg"].(string); !ok || msg != "1234" {
				t.Errorf("Expected msg to be '1234', got: %v", params["msg"])
			}

			if mobiles, ok := params["mobiles"].(string); !ok || mobiles != "18888888888" {
				t.Errorf("Expected mobiles to be '18888888888', got: %v", params["mobiles"])
			}

			if excode, ok := params["excode"].(string); !ok || excode != "mock-excode" {
				t.Errorf("Expected excode to be 'mock-excode', got: %v", params["excode"])
			}

			if templetid, ok := params["templetid"].(string); !ok || templetid != "123456" {
				t.Errorf("Expected templetid to be '123456', got: %v", params["templetid"])
			}

			// 验证签名
			signStr := "mock-cpcode123418888888888mock-excode123456mock-key"
			h := md5.New()
			h.Write([]byte(signStr))
			expectedSign := fmt.Sprintf("%x", h.Sum(nil))
			if sign, ok := params["sign"].(string); !ok || sign != expectedSign {
				t.Errorf("Expected sign to be '%s', got: %v (signStr: %s)", expectedSign, params["sign"], signStr)
			}

			// 返回成功响应
			return httpmock.NewJsonResponse(200, map[string]any{
				"resultcode": 0,
				"resultmsg":  "成功",
				"taskid":     "C20511170688217",
			})
		})

	// 创建网关配置
	config := map[string]any{
		"cpcode": "mock-cpcode",
		"key":    "mock-key",
		"excode": "mock-excode",
	}

	// 创建网关
	g := gateway.NewMaapGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("123456").
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

	resultCode, ok := respMap["resultcode"].(float64)
	if !ok || int(resultCode) != 0 {
		t.Errorf("Expected resultcode to be 0, got: %v", resultCode)
	}

	resultMsg, ok := respMap["resultmsg"].(string)
	if !ok || resultMsg != "成功" {
		t.Errorf("Expected resultmsg to be '成功', got: %v", resultMsg)
	}

	taskID, ok := respMap["taskid"].(string)
	if !ok || taskID != "C20511170688217" {
		t.Errorf("Expected taskid to be 'C20511170688217', got: %v", taskID)
	}
}

// TestMaapGatewayError 测试 MAAP 短信网关错误响应
func TestMaapGatewayError(t *testing.T) {
	// 设置模拟服务器
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 注册错误响应
	httpmock.RegisterResponder("POST", gateway.MaapEndpointURL,
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"resultcode": 301,
			"resultmsg":  "Error Message",
			"taskid":     "",
		}))

	// 创建网关配置
	config := map[string]any{
		"cpcode": "mock-cpcode",
		"key":    "mock-key",
		"excode": "mock-excode",
	}

	// 创建网关
	g := gateway.NewMaapGateway(config)

	// 创建消息
	msg := message.NewMessage().
		SetTemplate("123456").
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
	expectedError := "MAAP 短信发送失败: [301] Error Message"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be '%s', got: '%s'", expectedError, err.Error())
	}
}
