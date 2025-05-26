package gateway

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/anhao/go-easy-sms/message"
)

// 容联云通讯短信网关常量
const (
	// YuntongxunEndpointTemplate 容联云通讯短信 API 地址模板
	YuntongxunEndpointTemplate = "https://%s:%s/%s/%s/%s/%s/%s?sig=%s"
	// YuntongxunServerIP 容联云通讯短信 API 服务器地址
	YuntongxunServerIP = "app.cloopen.com"
	// YuntongxunDebugServerIP 容联云通讯短信 API 沙箱服务器地址
	YuntongxunDebugServerIP = "sandboxapp.cloopen.com"
	// YuntongxunDebugTemplateID 容联云通讯短信 API 沙箱模板 ID
	YuntongxunDebugTemplateID = 1
	// YuntongxunServerPort 容联云通讯短信 API 服务器端口
	YuntongxunServerPort = "8883"
	// YuntongxunSDKVersion 容联云通讯短信 API SDK 版本
	YuntongxunSDKVersion = "2013-12-26"
	// YuntongxunSDKVersionInt 容联云通讯短信 API SDK 国际版本
	YuntongxunSDKVersionInt = "v2"
	// YuntongxunSuccessCode 容联云通讯短信 API 成功状态码
	YuntongxunSuccessCode = "000000"
)

// YuntongxunGateway 容联云通讯短信网关
type YuntongxunGateway struct {
	*BaseGateway
	international bool
}

// NewYuntongxunGateway 创建一个新的容联云通讯短信网关
func NewYuntongxunGateway(config map[string]any) *YuntongxunGateway {
	return &YuntongxunGateway{
		BaseGateway:   NewBaseGateway("yuntongxun", config),
		international: false,
	}
}

// Send 发送短信
func (g *YuntongxunGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 重置国际短信标志
	g.international = false

	// 获取当前时间
	datetime := time.Now().Format("20060102150405")

	// 构建请求参数
	data := map[string]any{
		"appId": g.GetConfigString("app_id"),
	}

	var typeStr, resource string

	if to.InChineseMainland() {
		// 国内短信
		typeStr = "SMS"
		resource = "TemplateSMS"
		data["to"] = to.GetNumber()

		// 获取模板 ID
		templateID := msg.GetTemplate()
		if g.GetConfigBool("debug", false) {
			templateID = fmt.Sprintf("%d", YuntongxunDebugTemplateID)
		}
		data["templateId"], _ = parseTemplateID(templateID)

		// 转换数据格式
		msgData := msg.GetData()
		datas := make([]any, 0)

		// 尝试按索引排序
		for i := 0; i < len(msgData); i++ {
			key := fmt.Sprintf("%d", i)
			if val, ok := msgData[key]; ok {
				datas = append(datas, val)
			}
		}

		// 如果没有按索引排序的数据，则直接使用所有值
		if len(datas) == 0 {
			for _, val := range msgData {
				datas = append(datas, val)
			}
		}

		data["datas"] = datas
	} else {
		// 国际短信
		typeStr = "international"
		resource = "send"
		g.international = true
		// 格式化为 00 + 国际区号 + 号码
		data["mobile"] = fmt.Sprintf("00%d%s", to.GetIDDCode(), to.GetNumber())
		data["content"] = msg.GetContent()
	}

	// 构建请求地址
	endpoint := g.buildEndpoint(typeStr, resource, datetime)

	// 构建请求头
	headers := map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json;charset=utf-8",
		"Authorization": base64.StdEncoding.EncodeToString([]byte(g.GetConfigString("account_sid") + ":" + datetime)),
	}

	// 发送请求
	result, err := g.postJSON(endpoint, data, headers)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if statusCode, ok := result["statusCode"].(string); !ok || statusCode != YuntongxunSuccessCode {
		errorMsg := ""
		if statusCode != "" {
			errorMsg = statusCode
		}

		return result, fmt.Errorf("容联云通讯短信发送失败: %s", errorMsg)
	}

	return result, nil
}

// buildEndpoint 构建请求地址
func (g *YuntongxunGateway) buildEndpoint(typeStr, resource, datetime string) string {
	// 获取服务器地址
	serverIP := YuntongxunServerIP
	if g.GetConfigBool("debug", false) {
		serverIP = YuntongxunDebugServerIP
	}

	// 获取账号类型和 SDK 版本
	var accountType, sdkVersion string
	if g.international {
		accountType = "account"
		sdkVersion = YuntongxunSDKVersionInt
	} else {
		if g.GetConfigBool("is_sub_account", false) {
			accountType = "SubAccounts"
		} else {
			accountType = "Accounts"
		}
		sdkVersion = YuntongxunSDKVersion
	}

	// 生成签名
	sig := g.generateSignature(datetime)

	// 构建请求地址
	return fmt.Sprintf(YuntongxunEndpointTemplate,
		serverIP,
		YuntongxunServerPort,
		sdkVersion,
		accountType,
		g.GetConfigString("account_sid"),
		typeStr,
		resource,
		sig,
	)
}

// generateSignature 生成签名
func (g *YuntongxunGateway) generateSignature(datetime string) string {
	// 计算 MD5 哈希
	h := md5.New()
	h.Write([]byte(g.GetConfigString("account_sid") + g.GetConfigString("account_token") + datetime))
	return strings.ToUpper(fmt.Sprintf("%x", h.Sum(nil)))
}

// postJSON 发送 JSON 请求
func (g *YuntongxunGateway) postJSON(endpoint string, data map[string]any, headers map[string]string) (map[string]any, error) {
	// 使用 BaseGateway 的 PostJSON 方法发送请求
	return g.PostJSON(endpoint, data, headers)
}

// parseTemplateID 解析模板 ID
func parseTemplateID(templateID string) (int, error) {
	var id int
	_, err := fmt.Sscanf(templateID, "%d", &id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
