package gateway

import (
	"net/http"
	"strings"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
	"github.com/jarcoal/httpmock"
)

// setupAliyunMock sets up httpmock for Aliyun SMS API
func setupAliyunMock() string {
	// Define the base URL
	baseURL := "https://dysmsapi.aliyuncs.com"

	// Activate httpmock
	httpmock.Activate()

	// Register the request validator
	httpmock.RegisterResponder("GET", `=~^https://dysmsapi\.aliyuncs\.com/.*`, func(r *http.Request) (*http.Response, error) {
		// Parse query parameters
		query := r.URL.Query()

		// Validate required parameters
		if query.Get("AccessKeyId") == "" {
			return httpmock.NewJsonResponse(http.StatusOK, map[string]any{
				"Code":    "MissingAccessKeyId",
				"Message": "Specified access key is not found.",
			})
		}

		if query.Get("SignName") == "" {
			return httpmock.NewJsonResponse(http.StatusOK, map[string]any{
				"Code":    "SignNameError",
				"Message": "Specified signature is not found.",
			})
		}

		if query.Get("TemplateCode") == "" {
			return httpmock.NewJsonResponse(http.StatusOK, map[string]any{
				"Code":    "TemplateCodeError",
				"Message": "Specified template code is not found.",
			})
		}

		// Check if this is a test failure scenario
		if strings.Contains(query.Get("PhoneNumbers"), "error") {
			return httpmock.NewJsonResponse(http.StatusOK, map[string]any{
				"Code":    "isv.MOBILE_NUMBER_ILLEGAL",
				"Message": "Invalid phone number.",
			})
		}

		// Return success response
		return httpmock.NewJsonResponse(http.StatusOK, map[string]any{
			"Code":      "OK",
			"Message":   "OK",
			"RequestId": "F655A8D5-B967-440B-8683-DAD6FF8DE990",
			"BizId":     "900619746936498440^0",
		})
	})

	return baseURL
}

func TestAliyunGateway(t *testing.T) {
	// 创建一个阿里云网关
	config := map[string]any{
		"access_key_id":     "test_key_id",
		"access_key_secret": "test_key_secret",
		"sign_name":         "测试签名",
	}

	g := gateway.NewAliyunGateway(config)

	// 测试网关名称
	if g.GetName() != "aliyun" {
		t.Errorf("Expected name to be aliyun, got: %s", g.GetName())
	}

	// 测试配置
	if g.GetConfigString("access_key_id") != "test_key_id" {
		t.Errorf("Expected access_key_id to be test_key_id, got: %s", g.GetConfigString("access_key_id"))
	}

	if g.GetConfigString("access_key_secret") != "test_key_secret" {
		t.Errorf("Expected access_key_secret to be test_key_secret, got: %s", g.GetConfigString("access_key_secret"))
	}

	if g.GetConfigString("sign_name") != "测试签名" {
		t.Errorf("Expected sign_name to be 测试签名, got: %s", g.GetConfigString("sign_name"))
	}

	// 测试缺少必要配置时的错误
	badConfig := map[string]any{}
	badGateway := gateway.NewAliyunGateway(badConfig)

	phone := message.NewPhoneNumber("13800138000")
	msg := message.NewMessage().
		SetTemplate("SMS_12345678").
		SetData(map[string]any{
			"code": "123456",
		})

	_, err := badGateway.Send(phone, msg)
	if err == nil {
		t.Errorf("Expected error when missing required config")
	}

	// 测试缺少模板时的错误
	badMsg := message.NewMessage().
		SetData(map[string]any{
			"code": "123456",
		})

	_, err = g.Send(phone, badMsg)
	if err == nil {
		t.Errorf("Expected error when missing template")
	}
}

func TestAliyunGatewayWithMockServer(t *testing.T) {
	// 创建模拟服务器
	baseURL := setupAliyunMock()
	defer httpmock.DeactivateAndReset()

	// 创建一个阿里云网关，使用模拟服务器的URL
	config := map[string]any{
		"access_key_id":     "test_key_id",
		"access_key_secret": "test_key_secret",
		"sign_name":         "测试签名",
		"endpoint":          baseURL, // 使用模拟服务器的URL
	}

	g := gateway.NewAliyunGateway(config)

	// 测试成功发送短信
	phone := message.NewPhoneNumber("13800138000")
	msg := message.NewMessage().
		SetTemplate("SMS_12345678").
		SetData(map[string]any{
			"code": "123456",
		})

	resp, err := g.Send(phone, msg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// 验证响应
	respMap, ok := resp.(map[string]any)
	if !ok {
		t.Errorf("Expected response to be map[string]any, got: %T", resp)
	}

	if code, ok := respMap["Code"].(string); !ok || code != "OK" {
		t.Errorf("Expected Code to be OK, got: %v", code)
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

	if code, ok := respMap["Code"].(string); !ok || code != "isv.MOBILE_NUMBER_ILLEGAL" {
		t.Errorf("Expected Code to be isv.MOBILE_NUMBER_ILLEGAL, got: %v", code)
	}
}
