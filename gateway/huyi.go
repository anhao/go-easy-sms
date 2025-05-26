package gateway

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"time"

	"github.com/anhao/go-easy-sms/message"
)

// 互亿无线短信网关常量
const (
	// HuyiEndpointURL 互亿无线短信 API 地址
	HuyiEndpointURL = "http://106.ihuyi.com/webservice/sms.php?method=Submit"
	// HuyiEndpointFormat 互亿无线短信 API 格式
	HuyiEndpointFormat = "json"
	// HuyiSuccessCode 互亿无线短信 API 成功状态码
	HuyiSuccessCode = 2
)

// HuyiGateway 互亿无线短信网关
type HuyiGateway struct {
	*BaseGateway
}

// NewHuyiGateway 创建一个新的互亿无线短信网关
func NewHuyiGateway(config map[string]any) *HuyiGateway {
	return &HuyiGateway{
		BaseGateway: NewBaseGateway("huyi", config),
	}
}

// Send 发送短信
func (g *HuyiGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 获取电话号码
	mobile := ""
	if to.GetIDDCode() != 0 {
		mobile = fmt.Sprintf("%d %s", to.GetIDDCode(), to.GetNumber())
	} else {
		mobile = to.GetNumber()
	}

	// 获取签名
	signature := g.GetConfigString("signature", "")

	// 获取当前时间戳
	timestamp := time.Now().Unix()

	// 构建请求参数
	params := map[string]string{
		"account": g.GetConfigString("api_id"),
		"mobile":  mobile,
		"content": msg.GetContent(),
		"time":    strconv.FormatInt(timestamp, 10),
		"format":  HuyiEndpointFormat,
		"sign":    signature,
	}

	// 生成密码签名
	params["password"] = g.generateSign(params)

	// 发送请求
	result, err := g.post(HuyiEndpointURL, params)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if code, ok := result["code"].(float64); !ok || int(code) != HuyiSuccessCode {
		errorMsg := ""
		if msg, ok := result["msg"].(string); ok {
			errorMsg = msg
		}

		errorCode := 0
		if code, ok := result["code"].(float64); ok {
			errorCode = int(code)
		}

		return result, fmt.Errorf("互亿无线短信发送失败: [%d] %s", errorCode, errorMsg)
	}

	return result, nil
}

// generateSign 生成签名
func (g *HuyiGateway) generateSign(params map[string]string) string {
	// 构建签名字符串
	signStr := params["account"] + g.GetConfigString("api_key") + params["mobile"] + params["content"] + params["time"]

	// 计算 MD5 哈希
	h := md5.New()
	h.Write([]byte(signStr))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// post 发送 POST 请求
func (g *HuyiGateway) post(endpoint string, params map[string]string) (map[string]any, error) {
	// 使用 BaseGateway 的 Post 方法发送请求
	return g.Post(endpoint, params, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
}
