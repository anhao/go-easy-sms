package gateway

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/anhao/go-easy-sms/message"
)

// 阿里云 REST API 短信网关常量
const (
	// AliyunrestEndpointURL 阿里云 REST API 地址
	AliyunrestEndpointURL = "http://gw.api.taobao.com/router/rest"
	// AliyunrestEndpointVersion 阿里云 REST API 版本
	AliyunrestEndpointVersion = "2.0"
	// AliyunrestEndpointFormat 阿里云 REST API 格式
	AliyunrestEndpointFormat = "json"
	// AliyunrestEndpointMethod 阿里云 REST API 方法
	AliyunrestEndpointMethod = "alibaba.aliqin.fc.sms.num.send"
	// AliyunrestEndpointSignatureMethod 阿里云 REST API 签名方法
	AliyunrestEndpointSignatureMethod = "md5"
	// AliyunrestEndpointPartnerID 阿里云 REST API 合作伙伴 ID
	AliyunrestEndpointPartnerID = "EasySms"
)

// AliyunrestGateway 阿里云 REST API 短信网关
type AliyunrestGateway struct {
	*BaseGateway
}

// NewAliyunrestGateway 创建一个新的阿里云 REST API 短信网关
func NewAliyunrestGateway(config map[string]any) *AliyunrestGateway {
	return &AliyunrestGateway{
		BaseGateway: NewBaseGateway("aliyunrest", config),
	}
}

// Send 发送短信
func (g *AliyunrestGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 构建 URL 参数
	urlParams := map[string]string{
		"app_key":     g.GetConfigString("app_key"),
		"v":           AliyunrestEndpointVersion,
		"format":      AliyunrestEndpointFormat,
		"sign_method": AliyunrestEndpointSignatureMethod,
		"method":      AliyunrestEndpointMethod,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"partner_id":  AliyunrestEndpointPartnerID,
	}

	// 获取电话号码
	phoneNumber := ""
	if to.GetIDDCode() != 0 {
		phoneNumber = to.GetZeroPrefixedNumber()
	} else {
		phoneNumber = to.GetNumber()
	}

	// 构建请求参数
	data := msg.GetData()
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	params := map[string]string{
		"extend":             "",
		"sms_type":           "normal",
		"sms_free_sign_name": g.GetConfigString("sign_name"),
		"sms_param":          string(dataJSON),
		"rec_num":            phoneNumber,
		"sms_template_code":  msg.GetTemplate(),
	}

	// 生成签名
	urlParams["sign"] = g.generateSign(mergeStringMaps(params, urlParams))

	// 构建请求 URL
	endpoint := g.getEndpointURL(urlParams)

	// 发送请求
	result, err := g.post(endpoint, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if errorResponse, ok := result["error_response"].(map[string]any); ok {
		code := 0
		if codeValue, ok := errorResponse["code"].(float64); ok {
			code = int(codeValue)
		}

		errorMsg := ""
		if msgValue, ok := errorResponse["msg"].(string); ok {
			errorMsg = msgValue
		}

		return result, fmt.Errorf("阿里云 REST API 短信发送失败: [%d] %s", code, errorMsg)
	}

	return result, nil
}

// getEndpointURL 构建请求地址
func (g *AliyunrestGateway) getEndpointURL(params map[string]string) string {
	// 构建查询字符串
	query := url.Values{}
	for k, v := range params {
		query.Add(k, v)
	}

	return fmt.Sprintf("%s?%s", AliyunrestEndpointURL, query.Encode())
}

// generateSign 生成签名
func (g *AliyunrestGateway) generateSign(params map[string]string) string {
	// 按照键名排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建待签名字符串
	secretKey := g.GetConfigString("app_secret_key")
	var stringToBeSigned strings.Builder
	stringToBeSigned.WriteString(secretKey)

	for _, k := range keys {
		v := params[k]
		if !strings.HasPrefix(v, "@") {
			stringToBeSigned.WriteString(k)
			stringToBeSigned.WriteString(v)
		}
	}

	stringToBeSigned.WriteString(secretKey)

	// 计算 MD5 哈希
	h := md5.New()
	h.Write([]byte(stringToBeSigned.String()))
	return strings.ToUpper(fmt.Sprintf("%x", h.Sum(nil)))
}

// post 发送 POST 请求
func (g *AliyunrestGateway) post(endpoint string, params map[string]string) (map[string]any, error) {
	// 使用 BaseGateway 的 Post 方法发送请求
	return g.Post(endpoint, params, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
}

// mergeStringMaps 合并两个字符串映射
func mergeStringMaps(m1, m2 map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range m1 {
		result[k] = v
	}
	for k, v := range m2 {
		result[k] = v
	}
	return result
}
