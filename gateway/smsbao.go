package gateway

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/anhao/go-easy-sms/message"
)

// 短信宝网关常量
const (
	// SmsbaoEndpointURL 短信宝 API 地址模板
	SmsbaoEndpointURL = "http://api.smsbao.com/%s"
	// SmsbaoSuccessCode 短信宝 API 成功状态码
	SmsbaoSuccessCode = "0"
)

// SmsbaoGateway 短信宝网关
type SmsbaoGateway struct {
	*BaseGateway
	errorStatuses map[string]string
}

// NewSmsbaoGateway 创建一个新的短信宝网关
func NewSmsbaoGateway(config map[string]any) *SmsbaoGateway {
	return &SmsbaoGateway{
		BaseGateway: NewBaseGateway("smsbao", config),
		errorStatuses: map[string]string{
			"0":  "短信发送成功",
			"-1": "参数不全",
			"-2": "服务器空间不支持,请确认支持curl或者fsocket，联系您的空间商解决或者更换空间！",
			"30": "密码错误",
			"40": "账号不存在",
			"41": "余额不足",
			"42": "帐户已过期",
			"43": "IP地址限制",
			"50": "内容含有敏感词",
		},
	}
}

// Send 发送短信
func (g *SmsbaoGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 获取短信内容
	content := msg.GetContent()

	// 判断是国内短信还是国际短信
	var number string
	var action string

	if to.GetIDDCode() == 0 || to.GetIDDCode() == 86 {
		number = to.GetNumber()
		action = "sms"
	} else {
		number = to.GetUniversalNumber()
		action = "wsms"
	}

	// 构建请求参数
	params := map[string]string{
		"u": g.GetConfigString("user"),
		"p": g.md5(g.GetConfigString("password")),
		"m": number,
		"c": content,
	}

	// 构建请求地址
	endpoint := g.buildEndpoint(action)

	// 发送请求
	result, err := g.get(endpoint, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if result != SmsbaoSuccessCode {
		errorMsg := g.errorStatuses[result]
		if errorMsg == "" {
			errorMsg = "未知错误"
		}
		return result, fmt.Errorf("短信宝短信发送失败: [%s] %s", result, errorMsg)
	}

	return result, nil
}

// buildEndpoint 构建请求地址
func (g *SmsbaoGateway) buildEndpoint(action string) string {
	return fmt.Sprintf(SmsbaoEndpointURL, action)
}

// md5 计算 MD5 哈希值
func (g *SmsbaoGateway) md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// get 发送 GET 请求
func (g *SmsbaoGateway) get(endpoint string, params map[string]string) (string, error) {
	// 使用 BaseGateway 的 Get 方法发送请求
	// 创建上下文，设置超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(g.GetConfigFloat("timeout", 5.0))*time.Second)
	defer cancel()

	// 使用 BaseGateway 的 HTTP 客户端发送请求
	httpClient := g.GetHTTPClient()
	body, err := httpClient.Get(ctx, endpoint, params, nil)
	if err != nil {
		return "", err
	}

	// 短信宝返回的是字符串，不是 JSON，所以需要特殊处理
	return string(body), nil
}
