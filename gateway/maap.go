package gateway

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/anhao/go-easy-sms/message"
)

// MAAP 短信网关常量
const (
	// MaapEndpointURL MAAP 短信 API 地址
	MaapEndpointURL = "http://rcsapi.wo.cn:8000/umcinterface/sendtempletmsg"
)

// MaapGateway MAAP 短信网关
type MaapGateway struct {
	*BaseGateway
}

// NewMaapGateway 创建一个新的 MAAP 短信网关
func NewMaapGateway(config map[string]any) *MaapGateway {
	return &MaapGateway{
		BaseGateway: NewBaseGateway("maap", config),
	}
}

// Send 发送短信
func (g *MaapGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 获取消息数据
	data := msg.GetData()

	// 将数据转换为逗号分隔的字符串
	dataValues := make([]string, 0, len(data))
	for _, v := range data {
		dataValues = append(dataValues, fmt.Sprintf("%v", v))
	}
	dataStr := strings.Join(dataValues, ",")

	// 构建请求参数
	params := map[string]any{
		"cpcode":    g.GetConfigString("cpcode"),
		"msg":       dataStr,
		"mobiles":   to.GetNumber(),
		"excode":    g.GetConfigString("excode", ""),
		"templetid": msg.GetTemplate(),
	}

	// 生成签名
	params["sign"] = g.generateSign(params)

	// 发送请求
	result, err := g.postJSON(MaapEndpointURL, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if resultCode, ok := result["resultcode"].(float64); ok && resultCode != 0 {
		errorMsg := ""
		if msg, ok := result["resultmsg"].(string); ok {
			errorMsg = msg
		}

		return result, fmt.Errorf("MAAP 短信发送失败: [%d] %s", int(resultCode), errorMsg)
	}

	return result, nil
}

// generateSign 生成签名
func (g *MaapGateway) generateSign(params map[string]any) string {
	// 构建签名字符串
	signStr := fmt.Sprintf("%s%s%s%s%s%s",
		params["cpcode"],
		params["msg"],
		params["mobiles"],
		params["excode"],
		params["templetid"],
		g.GetConfigString("key"))

	// 计算 MD5 哈希
	h := md5.New()
	h.Write([]byte(signStr))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// postJSON 发送 JSON 请求
func (g *MaapGateway) postJSON(endpoint string, params map[string]any) (map[string]any, error) {
	// 使用 BaseGateway 的 PostJSON 方法发送请求
	return g.PostJSON(endpoint, params, map[string]string{
		"Content-Type": "application/json",
	})
}
