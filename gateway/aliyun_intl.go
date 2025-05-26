package gateway

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/anhao/go-easy-sms/message"
	"github.com/google/uuid"
)

// 阿里云国际短信网关常量
const (
	// AliyunIntlEndpointURL 阿里云国际短信 API 地址
	AliyunIntlEndpointURL = "https://dysmsapi.ap-southeast-1.aliyuncs.com"
	// AliyunIntlEndpointAction 阿里云国际短信 API 动作
	AliyunIntlEndpointAction = "SendMessageWithTemplate"
	// AliyunIntlEndpointVersion 阿里云国际短信 API 版本
	AliyunIntlEndpointVersion = "2018-05-01"
	// AliyunIntlEndpointFormat 阿里云国际短信 API 格式
	AliyunIntlEndpointFormat = "JSON"
	// AliyunIntlEndpointRegionID 阿里云国际短信 API 区域 ID
	AliyunIntlEndpointRegionID = "ap-southeast-1"
	// AliyunIntlEndpointSignatureMethod 阿里云国际短信 API 签名方法
	AliyunIntlEndpointSignatureMethod = "HMAC-SHA1"
	// AliyunIntlEndpointSignatureVersion 阿里云国际短信 API 签名版本
	AliyunIntlEndpointSignatureVersion = "1.0"
	// AliyunIntlSuccessCode 阿里云国际短信 API 成功状态码
	AliyunIntlSuccessCode = "OK"
)

// AliyunIntlGateway 阿里云国际短信网关
type AliyunIntlGateway struct {
	*BaseGateway
}

// NewAliyunIntlGateway 创建一个新的阿里云国际短信网关
func NewAliyunIntlGateway(config map[string]any) *AliyunIntlGateway {
	return &AliyunIntlGateway{
		BaseGateway: NewBaseGateway("aliyun_intl", config),
	}
}

// Send 发送短信
func (g *AliyunIntlGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 获取消息数据
	data := msg.GetData()

	// 获取签名
	signName := ""
	if signNameValue, ok := data["sign_name"]; ok {
		if signNameStr, ok := signNameValue.(string); ok {
			signName = signNameStr
			delete(data, "sign_name")
		}
	}
	if signName == "" {
		signName = g.GetConfigString("sign_name")
	}

	// 获取电话号码
	phoneNumber := ""
	if to.GetIDDCode() != 0 {
		phoneNumber = to.GetZeroPrefixedNumber()
	} else {
		phoneNumber = to.GetNumber()
	}

	// 构建请求参数
	params := map[string]string{
		"RegionId":         AliyunIntlEndpointRegionID,
		"AccessKeyId":      g.GetConfigString("access_key_id"),
		"Format":           AliyunIntlEndpointFormat,
		"SignatureMethod":  AliyunIntlEndpointSignatureMethod,
		"SignatureVersion": AliyunIntlEndpointSignatureVersion,
		"SignatureNonce":   uuid.New().String(),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          AliyunIntlEndpointVersion,
		"To":               phoneNumber,
		"Action":           AliyunIntlEndpointAction,
		"From":             signName,
		"TemplateCode":     msg.GetTemplate(),
	}

	// 将模板参数转换为 JSON 字符串
	templateParamJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	params["TemplateParam"] = string(templateParamJSON)

	// 生成签名
	params["Signature"] = g.generateSign(params)

	// 发送请求
	result, err := g.get(AliyunIntlEndpointURL, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if responseCode, ok := result["ResponseCode"].(string); !ok || responseCode != AliyunIntlSuccessCode {
		errorMsg := ""
		if msg, ok := result["ResponseDescription"].(string); ok {
			errorMsg = msg
		}

		errorCode := 0
		if responseCode != "" {
			// 尝试将 ResponseCode 转换为整数
			_, _ = fmt.Sscanf(responseCode, "%d", &errorCode)
		}

		return result, fmt.Errorf("阿里云国际短信发送失败: [%d] %s", errorCode, errorMsg)
	}

	return result, nil
}

// generateSign 生成签名
func (g *AliyunIntlGateway) generateSign(params map[string]string) string {
	// 按照键名排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建规范化请求字符串
	var canonicalizedQueryString strings.Builder
	for i, k := range keys {
		if i > 0 {
			canonicalizedQueryString.WriteString("&")
		}
		canonicalizedQueryString.WriteString(url.QueryEscape(k))
		canonicalizedQueryString.WriteString("=")
		canonicalizedQueryString.WriteString(url.QueryEscape(params[k]))
	}

	// 构建待签名字符串
	stringToSign := "GET&%2F&" + url.QueryEscape(canonicalizedQueryString.String())

	// 计算签名
	key := g.GetConfigString("access_key_secret") + "&"
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature
}

// get 发送 GET 请求
func (g *AliyunIntlGateway) get(endpoint string, params map[string]string) (map[string]any, error) {
	// 使用 BaseGateway 的 Get 方法发送请求
	return g.Get(endpoint, params, nil)
}
