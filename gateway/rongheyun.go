package gateway

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/anhao/go-easy-sms/message"
)

// 融合云短信网关常量
const (
	// RongheyunEndpointURL 融合云短信 API 地址
	RongheyunEndpointURL = "https://api.mix2.zthysms.com/v2/sendSmsTp"
	// RongheyunSuccessCode 融合云短信 API 成功状态码
	RongheyunSuccessCode = 200
)

// RongheyunGateway 融合云短信网关
type RongheyunGateway struct {
	*BaseGateway
}

// NewRongheyunGateway 创建一个新的融合云短信网关
func NewRongheyunGateway(config map[string]any) *RongheyunGateway {
	return &RongheyunGateway{
		BaseGateway: NewBaseGateway("rongheyun", config),
	}
}

// Send 发送短信
func (g *RongheyunGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 获取当前时间戳
	tKey := time.Now().Unix()

	// 生成密码
	password := g.generatePassword(tKey)

	// 构建请求参数
	params := map[string]any{
		"username":  g.GetConfigString("username"),
		"password":  password,
		"tKey":      tKey,
		"signature": g.GetConfigString("signature"),
		"tpId":      msg.GetTemplate(),
		"ext":       "",
		"extend":    "",
		"records": []map[string]any{
			{
				"mobile":    to.GetNumber(),
				"tpContent": msg.GetData(),
			},
		},
	}

	// 发送请求
	result, err := g.postJSON(RongheyunEndpointURL, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if resultCode, ok := result["code"].(float64); ok && int(resultCode) != RongheyunSuccessCode {
		errorMsg := ""
		if msg, ok := result["msg"].(string); ok {
			errorMsg = msg
		}

		return result, fmt.Errorf("融合云短信发送失败: [%d] %s", int(resultCode), errorMsg)
	}

	return result, nil
}

// generatePassword 生成密码
func (g *RongheyunGateway) generatePassword(tKey int64) string {
	// 第一次 MD5
	h1 := md5.New()
	h1.Write([]byte(g.GetConfigString("password")))
	firstMD5 := fmt.Sprintf("%x", h1.Sum(nil))

	// 第二次 MD5
	h2 := md5.New()
	h2.Write([]byte(firstMD5 + fmt.Sprintf("%d", tKey)))
	return fmt.Sprintf("%x", h2.Sum(nil))
}

// postJSON 发送 JSON 请求
func (g *RongheyunGateway) postJSON(endpoint string, params map[string]any) (map[string]any, error) {
	// 使用 BaseGateway 的 PostJSON 方法发送请求
	return g.PostJSON(endpoint, params, map[string]string{
		"Content-Type": "application/json; charset=UTF-8",
	})
}
