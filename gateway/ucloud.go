package gateway

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/anhao/go-easy-sms/message"
)

// UCloud 短信网关常量
const (
	// UcloudEndpointURL UCloud 短信 API 地址
	UcloudEndpointURL = "https://api.ucloud.cn"
	// UcloudEndpointAction UCloud 短信 API 动作
	UcloudEndpointAction = "SendUSMSMessage"
	// UcloudSuccessCode UCloud 短信 API 成功状态码
	UcloudSuccessCode = 0
)

// UcloudGateway UCloud 短信网关
type UcloudGateway struct {
	*BaseGateway
}

// NewUcloudGateway 创建一个新的 UCloud 短信网关
func NewUcloudGateway(config map[string]any) *UcloudGateway {
	return &UcloudGateway{
		BaseGateway: NewBaseGateway("ucloud", config),
	}
}

// Send 发送短信
func (g *UcloudGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 构建请求参数
	params := g.buildParams(to, msg)

	// 发送请求
	result, err := g.request(UcloudEndpointURL, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if retCode, ok := result["RetCode"].(float64); !ok || int(retCode) != UcloudSuccessCode {
		errorMsg := ""
		if msg, ok := result["Message"].(string); ok {
			errorMsg = msg
		}

		return result, fmt.Errorf("UCloud 短信发送失败: [%d] %s", int(retCode), errorMsg)
	}

	return result, nil
}

// buildParams 构建请求参数
func (g *UcloudGateway) buildParams(to *message.PhoneNumber, msg *message.Message) map[string]string {
	data := msg.GetData()
	params := map[string]string{
		"Action":     UcloudEndpointAction,
		"PublicKey":  g.GetConfigString("public_key"),
		"TemplateId": msg.GetTemplate(),
	}

	// 处理签名内容
	sigContent, ok := data["sig_content"].(string)
	if !ok || sigContent == "" {
		sigContent = g.GetConfigString("sig_content", "")
	}
	params["SigContent"] = sigContent

	// 处理模板参数
	code, ok := data["code"]
	if ok {
		switch v := code.(type) {
		case map[string]any:
			// 如果是多个参数，使用 TemplateParams.0, TemplateParams.1 等格式
			for key, value := range v {
				params[fmt.Sprintf("TemplateParams.%s", key)] = fmt.Sprintf("%v", value)
			}
		case []any:
			// 如果是数组，使用 TemplateParams.0, TemplateParams.1 等格式
			for i, value := range v {
				params[fmt.Sprintf("TemplateParams.%d", i)] = fmt.Sprintf("%v", value)
			}
		default:
			// 如果是单个值，使用 TemplateParams.0
			if code != nil && fmt.Sprintf("%v", code) != "" {
				params["TemplateParams.0"] = fmt.Sprintf("%v", code)
			}
		}
	}

	// 处理手机号码
	mobiles, ok := data["mobiles"]
	if ok && mobiles != nil {
		switch v := mobiles.(type) {
		case []any:
			// 如果是数组，使用 PhoneNumbers.0, PhoneNumbers.1 等格式
			for i, value := range v {
				params[fmt.Sprintf("PhoneNumbers.%d", i)] = fmt.Sprintf("%v", value)
			}
		default:
			// 如果是单个值，使用 PhoneNumbers.0
			if fmt.Sprintf("%v", mobiles) != "" {
				params["PhoneNumbers.0"] = fmt.Sprintf("%v", mobiles)
			}
		}
	} else {
		// 使用传入的手机号码
		params["PhoneNumbers.0"] = to.String()
	}

	// 处理项目 ID
	projectID := g.GetConfigString("project_id", "")
	if projectID != "" {
		params["ProjectId"] = projectID
	}

	// 生成签名
	privateKey := g.GetConfigString("private_key")
	signature := g.getSignature(params, privateKey)
	params["Signature"] = signature

	return params
}

// getSignature 生成签名
func (g *UcloudGateway) getSignature(params map[string]string, privateKey string) string {
	// 按照键名排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接参数
	var paramsData strings.Builder
	for _, k := range keys {
		paramsData.WriteString(k)
		paramsData.WriteString(params[k])
	}
	paramsData.WriteString(privateKey)

	// 计算 SHA1 哈希
	h := sha1.New()
	h.Write([]byte(paramsData.String()))
	return hex.EncodeToString(h.Sum(nil))
}

// request 发送 HTTP 请求
func (g *UcloudGateway) request(endpoint string, params map[string]string) (map[string]any, error) {
	// 使用 BaseGateway 的 Get 方法发送请求
	return g.Get(endpoint, params, nil)
}
