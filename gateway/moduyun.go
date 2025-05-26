package gateway

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/anhao/go-easy-sms/message"
)

// 摩杜云短信网关常量
const (
	// ModuyunEndpointURL 摩杜云短信 API 地址
	ModuyunEndpointURL = "https://live.moduyun.com/sms/v2/sendsinglesms"
)

// ModuyunGateway 摩杜云短信网关
type ModuyunGateway struct {
	*BaseGateway
}

// NewModuyunGateway 创建一个新的摩杜云短信网关
func NewModuyunGateway(config map[string]any) *ModuyunGateway {
	return &ModuyunGateway{
		BaseGateway: NewBaseGateway("moduyun", config),
	}
}

// Send 发送短信
func (g *ModuyunGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 生成随机数
	// 使用 Go 1.20+ 推荐的方式生成随机数
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	random := r.Intn(900000) + 100000 // 生成 100000-999999 之间的随机数

	// 构建 URL 参数
	urlParams := map[string]string{
		"accesskey": g.GetConfigString("accesskey"),
		"random":    fmt.Sprintf("%d", random),
	}

	// 获取国际区号
	nationcode := "86"
	if to.GetIDDCode() != 0 {
		nationcode = fmt.Sprintf("%d", to.GetIDDCode())
	}

	// 获取当前时间戳
	timestamp := time.Now().Unix()

	// 获取消息数据
	data := msg.GetData()
	dataValues := make([]any, 0, len(data))
	for _, v := range data {
		dataValues = append(dataValues, v)
	}

	// 构建请求参数
	params := map[string]any{
		"tel": map[string]any{
			"mobile":     to.GetNumber(),
			"nationcode": nationcode,
		},
		"signId":     g.GetConfigString("signId", ""),
		"templateId": msg.GetTemplate(),
		"time":       timestamp,
		"type":       g.GetConfigInt("type", 0),
		"params":     dataValues,
		"ext":        "",
		"extend":     "",
	}

	// 生成签名
	params["sig"] = g.generateSign(params, random)

	// 发送请求
	result, err := g.postJSON(g.getEndpointURL(urlParams), params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if resultCode, ok := result["result"].(float64); ok && resultCode != 0 {
		errorMsg := ""
		if msg, ok := result["errmsg"].(string); ok {
			errorMsg = msg
		}

		return result, fmt.Errorf("摩杜云短信发送失败: [%d] %s", int(resultCode), errorMsg)
	}

	return result, nil
}

// getEndpointURL 构建请求地址
func (g *ModuyunGateway) getEndpointURL(params map[string]string) string {
	// 构建查询字符串
	query := url.Values{}
	for k, v := range params {
		query.Add(k, v)
	}

	return fmt.Sprintf("%s?%s", ModuyunEndpointURL, query.Encode())
}

// generateSign 生成签名
func (g *ModuyunGateway) generateSign(params map[string]any, random int) string {
	// 获取手机号码
	tel, _ := params["tel"].(map[string]any)
	mobile, _ := tel["mobile"].(string)

	// 构建签名字符串
	signStr := fmt.Sprintf("secretkey=%s&random=%d&time=%d&mobile=%s",
		g.GetConfigString("secretkey"),
		random,
		params["time"],
		mobile)

	// 计算 SHA-256 哈希
	h := sha256.New()
	h.Write([]byte(signStr))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// postJSON 发送 JSON 请求
func (g *ModuyunGateway) postJSON(endpoint string, params map[string]any) (map[string]any, error) {
	// 使用 BaseGateway 的 PostJSON 方法发送请求
	return g.PostJSON(endpoint, params, map[string]string{
		"Content-Type": "application/json",
	})
}
