<h1 align="center">Go Easy SMS</h1>

<p align="center">:calling: 一款满足你的多种发送需求的 Go 语言短信发送组件</p>

## 特点

1. 支持目前市面多家服务商
2. 一套写法兼容所有平台
3. 简单配置即可灵活增减服务商
4. 内置多种服务商轮询策略、支持自定义轮询策略
5. 统一的返回值格式，便于日志与监控
6. 自动轮询选择可用的服务商
7. 完善的日志记录功能
8. 支持自动注册已有网关和自定义网关
9. 完整的单元测试覆盖
10. **高性能优化**：线程安全的网关注册机制，零内存分配，极低延迟（~15ns/op）
11. **结构化错误处理**：详细的错误信息和上下文，便于调试和监控

## 平台支持

目前已实现的平台：

- [阿里云](https://www.aliyun.com/)
- [阿里云国际](https://www.alibabacloud.com/help/zh/sms/list-of-operations-by-function/)
- [淘宝开放平台](https://developer.alibaba.com/docs/api.htm?apiId=25450&amp;scopeId=11872)
- [云片](https://www.yunpian.com)
- [腾讯云](https://cloud.tencent.com/product/sms)
- [创蓝](https://www.253.com/)
- [创蓝云智](https://www.chuanglan.com/)
- [UCloud](https://www.ucloud.cn/)
- [百度云](https://cloud.baidu.com/)
- [天翼云](https://www.ctyun.cn/document/10020426/10021544)
- [华信](http://www.ipyy.com/)
- [互亿无线](http://www.ihuyi.com/)
- [聚合数据](https://www.juhe.cn/)
- [凯信通](http://www.kingtto.cn/)
- [螺丝帽](https://luosimao.com/)
- [MAAP](https://maap.wo.cn/)
- [摩杜云](https://www.moduyun.com/)
- [时代互联](http://www.nowcn.com/)
- [七牛云](https://www.qiniu.com/)
- [融云](https://www.rongcloud.cn/)
- [助通融合云通信](https://zthysms.com)
- [SendCloud](https://www.sendcloud.net/)
- [Twilio](https://www.twilio.com/)
- [容联云通讯](https://www.yuntongxun.com/)
- [赛邮云](https://www.mysubmail.com/)
- [短信宝](http://www.smsbao.com/)
- [火山引擎](https://console.volcengine.com/sms/)
- [UE35.NET](http://uesms.ue35.cn/)
- [网易云信](https://yunxin.163.com/sms)
- [云之讯](https://www.ucpaas.com/index.html)
- [移动云MAS](https://mas.10086.cn/)

## 环境需求

- Go 1.18+

## 安装

```bash
go get github.com/anhao/go-easy-sms
```

## 使用

```go
package main

import (
	"fmt"
	"log"

	"github.com/anhao/go-easy-sms"
	"github.com/anhao/go-easy-sms/config"
	"github.com/anhao/go-easy-sms/message"
)

func main() {
	// 配置信息
	cfg := config.NewConfig()

	// HTTP 请求的超时时间（秒）
	cfg.Timeout = 5.0

	// 默认发送配置
	cfg.DefaultGateways = []string{"yunpian", "aliyun"}

	// 可用的网关配置
	cfg.GatewayConfigs = map[string]map[string]any{
		"errorlog": {
			"file": "/tmp/easy-sms.log",
		},
		"yunpian": {
			"api_key":   "824f0ff2f71cab52936axxxxxxxxxx",
			"signature": "【默认签名】",
		},
		"aliyun": {
			"access_key_id":     "your-access-key-id",
			"access_key_secret": "your-access-key-secret",
			"sign_name":         "your-sign-name",
		},
		// 更多网关配置...
	}

	// 创建 EasySms 实例
	sms := easysms.New(cfg)

	// 发送短信
	results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
		SetContent("您的验证码为: 6379").
		SetTemplate("SMS_001").
		SetData(map[string]any{
			"code": "6379",
		}))

	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// 处理结果
	for gateway, result := range results {
		if result.Status == easysms.StatusSuccess {
			fmt.Printf("Successfully sent message via %s\n", gateway)
		} else {
			fmt.Printf("Failed to send message via %s: %v\n", gateway, result.Error)
		}
	}
}
```

## 短信内容

由于使用多网关发送，所以一条短信要支持多平台发送，每家的发送方式不一样，但是我们抽象定义了以下公用属性：

- `content` 文字内容，使用在像云片类似的以文字内容发送的平台
- `template` 模板 ID，使用在以模板ID来发送短信的平台
- `data` 模板变量，使用在以模板ID来发送短信的平台

所以，在使用过程中你可以根据所要使用的平台定义发送的内容。

```go
msg := message.NewMessage().
	SetContent("您的验证码为: 6379").
	SetTemplate("SMS_001").
	SetData(map[string]any{
		"code": "6379",
	})
```

你也可以使用函数来返回对应的值（类似于 PHP 版本的闭包）：

```go
// 使用 SimpleSend 方法时可以使用函数返回不同网关的内容
results, err := sms.SimpleSend("13800138000", map[string]any{
    "content": func(gateway string) string {
        if gateway == "yunpian" {
            return "云片专用验证码：1235"
        }
        return "您的验证码为: 6379"
    },
    "template": func(gateway string) string {
        if gateway == "aliyun" {
            return "TP2818"
        }
        return "SMS_001"
    },
    "data": func(gateway string) map[string]any {
        return map[string]any{
            "code": "6379",
        }
    },
})
```

## 发送网关

默认使用配置中的 `DefaultGateways` 设置来发送，如果某一条短信你想要覆盖默认的设置，可以在消息中指定网关：

```go
// 在 Message 中设置网关
msg := message.NewMessage().
    SetContent("您的验证码为: 6379").
    SetTemplate("SMS_001").
    SetData(map[string]any{
        "code": "6379",
    }).
    SetGateways([]string{"yunpian", "aliyun"}) // 这里的网关配置将会覆盖全局默认值

// 发送短信
results, err := sms.Send(message.NewPhoneNumber("13800138000"), msg)

// 使用 SimpleSend 方法时也可以在数据中指定网关
results, err := sms.SimpleSend("13800138000", map[string]any{
    "content": "您的验证码为: 6379",
    "gateways": []string{"yunpian", "aliyun"}, // 指定网关
})
```

## 日志记录

```go
import (
	"os"
	"github.com/anhao/go-easy-sms/logger"
)

// 设置日志级别
logger.SetLevel(logger.DEBUG)

// 将日志输出到文件
logFile, _ := os.OpenFile("sms.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
logger.SetOutput(logFile)

// 禁用日志
logger.Disable()

// 启用日志
logger.Enable()
```

## 自定义网关

本组件已经支持用户自定义网关，你可以很方便地配置即可当成与其它组件一样使用。

### 基础自定义网关

```go
import (
	"fmt"
	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
)

// 自定义网关
type CustomGateway struct {
	name   string
	config map[string]any
}

func NewCustomGateway(config map[string]any) *CustomGateway {
	return &CustomGateway{
		name:   "custom",
		config: config,
	}
}

func (g *CustomGateway) GetName() string {
	return g.name
}

func (g *CustomGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 实现发送逻辑
	fmt.Printf("Custom gateway sending message to %s: %s\n", to.String(), msg.GetContent())
	return map[string]any{
		"message_id": "custom_12345",
		"status":     "sent",
		"gateway":    g.name,
	}, nil
}

// 注册自定义网关实例
sms.RegisterGateway("custom", NewCustomGateway(config))

// 或者注册自定义网关创建函数（推荐）
sms.RegisterGatewayCreator("custom", func(config map[string]any) (gateway.Gateway, error) {
	gw := NewCustomGateway(config)
	if gw == nil {
		return nil, fmt.Errorf("failed to create custom gateway")
	}
	return gw, nil
})
```

### 高级自定义网关（带配置验证和错误处理）

```go
import (
	"errors"
	"fmt"
	"github.com/anhao/go-easy-sms"
	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
)

// 高级自定义网关
type AdvancedCustomGateway struct {
	name     string
	endpoint string
	apiKey   string
	timeout  int
}

func NewAdvancedCustomGateway(config map[string]any) (*AdvancedCustomGateway, error) {
	// 验证必需的配置
	endpoint, ok := config["endpoint"].(string)
	if !ok || endpoint == "" {
		return nil, errors.New("endpoint is required")
	}

	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, errors.New("api_key is required")
	}

	// 可选配置
	timeout := 30
	if t, ok := config["timeout"].(int); ok {
		timeout = t
	}

	return &AdvancedCustomGateway{
		name:     "advanced_custom",
		endpoint: endpoint,
		apiKey:   apiKey,
		timeout:  timeout,
	}, nil
}

func (g *AdvancedCustomGateway) GetName() string {
	return g.name
}

func (g *AdvancedCustomGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 验证输入
	if to == nil {
		return nil, errors.New("phone number is required")
	}

	if msg == nil {
		return nil, errors.New("message is required")
	}

	content := msg.GetContent()
	if content == "" {
		return nil, errors.New("message content is required")
	}

	// 实现具体的发送逻辑
	fmt.Printf("Advanced Custom Gateway: Sending to %s via %s\n", to.String(), g.endpoint)
	fmt.Printf("API Key: %s, Timeout: %d\n", g.apiKey, g.timeout)
	fmt.Printf("Content: %s\n", content)

	// 模拟成功响应
	return map[string]any{
		"message_id": "advanced_12345",
		"status":     "queued",
		"gateway":    g.name,
		"endpoint":   g.endpoint,
	}, nil
}

// 注册高级自定义网关
sms.RegisterGatewayCreator("advanced_custom", func(config map[string]any) (gateway.Gateway, error) {
	gw, err := NewAdvancedCustomGateway(config)
	if err != nil {
		// 使用结构化错误处理
		return nil, &easysms.GatewayError{
			GatewayName: "advanced_custom",
			Operation:   "creation",
			Err:         err,
		}
	}
	return gw, nil
})
```

### 错误处理

自定义网关可以使用新的结构化错误处理机制：

```go
// 获取网关时的错误处理
gw, err := sms.Gateway("custom")
if err != nil {
	if gwErr, ok := err.(*easysms.GatewayError); ok {
		fmt.Printf("Gateway %s %s failed: %v\n",
			gwErr.GatewayName, gwErr.Operation, gwErr.Err)
	} else {
		fmt.Printf("Unknown error: %v\n", err)
	}
}

// 发送时的错误处理
results, err := sms.Send(phone, msg)
if err != nil {
	fmt.Printf("Send failed: %v\n", err)
}

for gatewayName, result := range results {
	if result.Status == easysms.StatusSuccess {
		fmt.Printf("Successfully sent via %s: %v\n", gatewayName, result.Data)
	} else {
		fmt.Printf("Failed to send via %s: %v\n", gatewayName, result.Error)
	}
}
```

## 简单发送方式

`SimpleSend` 方法提供了一个更简单的发送接口

```go
// 示例1：使用字符串内容
results, err := sms.SimpleSend("13800138000", map[string]any{
    "content": "您的验证码是：123456，有效期为5分钟。",
})

// 示例2：使用函数内容、模板和数据（类似于 PHP 版本的闭包）
results, err := sms.SimpleSend("13800138000", map[string]any{
    "content": func(gateway string) string {
        if gateway == "aliyun" {
            return "阿里云短信内容"
        }
        return "您的验证码为: 6379"
    },
    "template": func(gateway string) string {
        if gateway == "aliyun" {
            return "SMS_001"
        }
        return ""
    },
    "data": func(gateway string) map[string]any {
        if gateway == "aliyun" {
            return map[string]any{
                "code": 6379,
            }
        }
        return nil
    },
    "gateways": []string{"aliyun", "yunpian", "custom"},
})
```

## 国际短信

国际短信与国内短信的区别是号码前面需要加国际码，使用方法如下：

```go
// 发送到国际码为 31 的国际号码
phone := message.NewPhoneNumber("13800138000", 31)

// 发送短信
sms.Send(phone, msg)
```

## 返回值

由于使用多网关发送，所以返回值为一个 map，结构如下：

```go
map[string]Result{
	"aliyun": {
		Gateway: "aliyun",
		Status:  "success",
		Data:    {...}, // 平台返回值
	},
	"yunpian": {
		Gateway: "yunpian",
		Status:  "failure",
		Error:   error, // 错误信息
	},
}
```

如果所选网关列表均发送失败时，将会返回错误。

## 各平台配置说明

### [阿里云](https://www.aliyun.com/)

短信内容使用 `template` + `data`

```go
"aliyun": {
    "access_key_id":     "your-access-key-id",
    "access_key_secret": "your-access-key-secret",
    "sign_name":         "your-sign-name",
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("SMS_12345678").
    SetData(map[string]any{
        "code": "123456",
    }))
```

### [云片](https://www.yunpian.com)

短信内容使用 `content`

```go
"yunpian": {
    "api_key":   "your-api-key",
    "signature": "【默认签名】", // 内容中无签名时使用
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetContent("您的验证码为: 6379，有效期为5分钟。"))
```

### [腾讯云](https://cloud.tencent.com/product/sms)

短信内容使用 `template` + `data`

```go
"qcloud": {
    "sdk_app_id": "your-sdk-app-id",
    "secret_id":  "your-secret-id",
    "secret_key": "your-secret-key",
    "sign_name":  "your-sign-name",
    "region":     "ap-guangzhou", // 可选，默认为 ap-guangzhou
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("101234").
    SetData(map[string]any{
        "0": "a", // 按照模板参数顺序填充
        "1": "b",
        "2": "c",
    }))
```

### [创蓝](https://www.253.com/)

短信内容使用 `content`

```go
"chuanglan": {
    "account":        "your-account",
    "password":       "your-password",
    "channel":        "smsbj1", // 可选，验证码通道，默认为 smsbj1
    // "channel":     "smssh1", // 营销通道
    "sign":           "【签名】", // 使用营销通道时必填
    "unsubscribe":    "回TD退订", // 使用营销通道时必填
    "intel_account":  "your-international-account", // 可选，国际短信账号，默认使用 account
    "intel_password": "your-international-password", // 可选，国际短信密码，默认使用 password
},
```

发送示例：

```go
// 发送验证码短信
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetContent("您的验证码为: 6379，有效期为5分钟。"))

// 发送国际短信
phone := message.NewPhoneNumber("13800138000", 86)
results, err = sms.Send(phone, message.NewMessage().
    SetContent("Your verification code is: 6379, valid for 5 minutes."))
```

### [UCloud](https://www.ucloud.cn/)

短信内容使用 `template` + `data`

```go
"ucloud": {
    "private_key": "your-private-key", // 私钥
    "public_key":  "your-public-key",  // 公钥
    "sig_content": "your-sig-content", // 签名
    "project_id":  "your-project-id",  // 项目ID，默认不填，子账号才需要填
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("UTAXXXXX").
    SetData(map[string]any{
        "code": "123456", // 模板参数
        // 批量发送短信
        // "mobiles": []string{"13800138000", "13900139000"},
    }))
```

### [百度云](https://cloud.baidu.com/)

短信内容使用 `template` + `data`

```go
"baidu": {
    "ak":        "your-access-key",    // 百度云 AK
    "sk":        "your-secret-key",    // 百度云 SK
    "invoke_id": "your-invoke-id",     // 短信服务的调用ID
    "domain":    "smsv3.bj.baidubce.com", // 可选，默认为 smsv3.bj.baidubce.com
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("sms-tmpl-xxxxxx").
    SetData(map[string]any{
        "code": "123456",
        // 其他模板参数
    }))
```

### [天翼云](https://www.ctyun.cn/)

短信内容使用 `template` + `data`

```go
"ctyun": {
    "secret_key":    "your-secret-key",    // 天翼云 SecretKey
    "access_key":    "your-access-key",    // 天翼云 AccessKey
    "template_code": "your-template-code", // 短信模板ID
    "sign_name":     "your-sign-name",     // 短信签名
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("SMS64124870510").
    SetData(map[string]any{
        "code": "123456",
    }))
```

### [华信](http://www.ipyy.com/)

短信内容使用 `content`

```go
"huaxin": {
    "user_id":  "your-user-id",   // 用户ID
    "account":  "your-account",   // 账号
    "password": "your-password",  // 密码
    "ip":       "your-ip",        // IP地址
    "ext_no":   "your-ext-no",    // 扩展号码，可选
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetContent("您的验证码为: 6379，有效期为5分钟。"))
```

### [赛邮云](https://www.mysubmail.com/)

短信内容使用 `content` 或 `template` + `data`

```go
"submail": {
    "app_id":  "your-app-id",    // 应用ID
    "app_key": "your-app-key",   // 应用密钥
    "project": "your-project",   // 项目标识，使用模板发送时必填
},
```

发送示例：

```go
// 使用内容发送
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetContent("您的验证码为: 6379，有效期为5分钟。"))

// 使用模板发送
results, err = sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("your-template-id").
    SetData(map[string]any{
        "code": "123456",
        "project": "your-project", // 可以在发送时指定项目标识，覆盖配置中的值
    }))
```

### [短信宝](http://www.smsbao.com/)

短信内容使用 `content`

```go
"smsbao": {
    "user":     "your-username", // 用户名
    "password": "your-password", // 密码
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetContent("您的验证码为: 6379，有效期为5分钟。"))
```

### [阿里云国际](https://www.alibabacloud.com/help/zh/doc-detail/162279.htm)

短信内容使用 `template` + `data`

```go
"aliyun_intl": {
    "access_key_id":     "your-access-key-id",     // 访问密钥 ID
    "access_key_secret": "your-access-key-secret", // 访问密钥密钥
    "sign_name":         "your-sign-name",         // 短信签名
},
```

发送示例：

```go
// 使用国际电话号码
phone := message.NewPhoneNumber("13800138000", 86)

results, err := sms.Send(phone, message.NewMessage().
    SetTemplate("SMS_00000001").
    SetData(map[string]any{
        "code": "123456",
    }))
```

### [淘宝开放平台](https://developer.alibaba.com/docs/api.htm?apiId=25450&amp;scopeId=11872)

短信内容使用 `template` + `data`

```go
"aliyunrest": {
    "app_key":        "your-app-key",        // 应用密钥
    "app_secret_key": "your-app-secret-key", // 应用密钥密钥
    "sign_name":      "your-sign-name",      // 短信签名
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("SMS_00000001").
    SetData(map[string]any{
        "code": "123456",
    }))
```

### [创蓝云智](https://www.chuanglan.com/)

短信内容使用 `content` 或 `template` + `data`

```go
"chuanglanv1": {
    "account":        "your-account",        // 账号
    "password":       "your-password",       // 密码
    "channel":        "v1/send",             // 可选，通道，默认为 v1/send
    // "channel":     "variable",            // 变量通道
    "intel_account":  "your-intel-account",  // 可选，国际短信账号，默认使用 account
    "intel_password": "your-intel-password", // 可选，国际短信密码，默认使用 password
    "needstatus":     false,                 // 可选，是否需要状态报告，默认为 false
},
```

发送示例：

```go
// 普通短信发送
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetContent("您的验证码为: 6379，有效期为5分钟。"))

// 变量短信发送
results, err = sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("您的验证码为{$code}，有效期为{$time}分钟。").
    SetData(map[string]any{
        "phone": "15800000000,15900000000", // 多个手机号用逗号分隔
        "data": "code=1234,time=5;code=5678,time=10", // 多组参数用分号分隔，每组参数用逗号分隔
    }))
```

### [互亿无线](http://www.ihuyi.com/)

短信内容使用 `content`

```go
"huyi": {
    "api_id":    "your-api-id",    // API ID
    "api_key":   "your-api-key",   // API 密钥
    "signature": "your-signature", // 可选，短信签名
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetContent("您的验证码为: 6379，有效期为5分钟。"))
```

### [聚合数据](https://www.juhe.cn/)

短信内容使用 `template` + `data`

```go
"juhe": {
    "app_key": "your-app-key", // 应用密钥
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("TPL_123456").
    SetData(map[string]any{
        "#code#": "123456",
        "#time#": "5",
    }))
```

### [凯信通](http://www.kingtto.cn/)

短信内容使用 `content`

```go
"kingtto": {
    "userid":   "your-userid",   // 用户ID
    "account":  "your-account",  // 账号
    "password": "your-password", // 密码
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetContent("您的验证码为: 6379，有效期为5分钟。"))
```

### [螺丝帽](https://luosimao.com/)

短信内容使用 `content`

```go
"luosimao": {
    "api_key": "your-api-key", // API 密钥
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetContent("您的验证码为: 6379，有效期为5分钟。"))
```

### [MAAP](https://maap.wo.cn/)

短信内容使用 `template` + `data`

```go
"maap": {
    "cpcode": "your-cpcode", // 企业编码
    "key":    "your-key",    // 签名密钥
    "excode": "your-excode", // 可选，扩展码
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("356120").
    SetData(map[string]any{
        "0": "123456", // 按照模板参数顺序填充
    }))
```

### [摩杜云](https://www.moduyun.com/)

短信内容使用 `template` + `data`

```go
"moduyun": {
    "accesskey": "your-accesskey", // 访问密钥
    "secretkey": "your-secretkey", // 密钥
    "signId":    "your-signId",    // 签名 ID
    "type":      0,                // 可选，短信类型，默认为 0
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("5a95****b953").
    SetData(map[string]any{
        "0": "123456", // 对应模板的第一个参数
        "1": "5",      // 对应模板的第二个参数
    }))
```

### [时代互联](http://www.nowcn.com/)

短信内容使用 `content`

```go
"nowcn": {
    "key":      "your-key",      // 用户 ID
    "secret":   "your-secret",   // 密码
    "api_type": "your-api-type", // API 类型
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetContent("您的验证码为: 6379，有效期为5分钟。"))
```

### [七牛云](https://www.qiniu.com/)

短信内容使用 `template` + `data`

```go
"qiniu": {
    "access_key": "your-access-key", // 访问密钥
    "secret_key": "your-secret-key", // 密钥
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("1231234123412341234").
    SetData(map[string]any{
        "code": "123456",
    }))
```

### [融云](https://www.rongcloud.cn/)

短信内容使用 `template` + `data`

```go
"rongcloud": {
    "app_key":    "your-app-key",    // 应用密钥
    "app_secret": "your-app-secret", // 应用密钥密钥
},
```

支持的动作：
- `sendCode`：发送验证码（默认）
- `verifyCode`：验证验证码，需要提供 `code` 和 `sessionId`
- `sendNotify`：发送通知

发送示例：

```go
// 发送验证码
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("templateId").
    SetData(map[string]any{
        "action": "sendCode", // 默认为 sendCode
        "mobile": "13800138000",
        "region": "86",
    }))

// 验证验证码
results, err = sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetData(map[string]any{
        "action":    "verifyCode",
        "code":      "123456",
        "sessionId": "session_id",
    }))

// 发送通知
results, err = sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("templateId").
    SetData(map[string]any{
        "action": "sendNotify",
        "params": []string{"param1", "param2"}, // 模板参数列表
    }))
```

### [助通融合云通信](https://zthysms.com/)

短信内容使用 `template` + `data`

```go
"rongheyun": {
    "username":  "your-username",  // 用户名
    "password":  "your-password",  // 密码
    "signature": "your-signature", // 签名
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("31874").
    SetData(map[string]any{
        "valid_code": "888888", // 对应模板中的 {valid_code} 变量
    }))
```

### [SendCloud](https://www.sendcloud.net/)

短信内容使用 `template` + `data`

```go
"sendcloud": {
    "sms_user":  "your-sms-user", // 短信用户名
    "sms_key":   "your-sms-key",  // 短信密钥
    "timestamp": false,           // 可选，是否添加时间戳，默认为 false
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("template_id").
    SetData(map[string]any{
        "code": "123456",
        "time": "5",
    }))
```


### [Twilio](https://www.twilio.com/)

短信内容使用 `content`

```go
"twilio": {
    "account_sid": "your-account-sid", // 账号 SID
    "token":       "your-token",       // 令牌
    "from":        "your-from",        // 发送者
},
```

发送示例：

```go
// 注意：Twilio 需要使用国际格式的电话号码，带有 + 前缀
phone := message.NewPhoneNumber("13800138000", 86) // 将自动添加 + 前缀

results, err := sms.Send(phone, message.NewMessage().
    SetContent("Your verification code is: 6379, valid for 5 minutes."))
```

### [容联云通讯](https://www.yuntongxun.com/)

国内短信使用 `template` + `data`，国际短信使用 `content`

```go
"yuntongxun": {
    "debug":          false,           // 是否使用沙箱环境，默认为 false
    "is_sub_account": false,           // 是否使用子账号，默认为 false
    "account_sid":    "your-account-sid", // 主账号 ID
    "account_token":  "your-account-token", // 主账号令牌
    "app_id":         "your-app-id",   // 应用 ID
},
```

发送示例：

```go
// 国内短信
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("templateId").
    SetData(map[string]any{
        "params": []string{"123456", "5"}, // 模板参数列表
    }))

// 国际短信
phone := message.NewPhoneNumber("13800138000", 86)
results, err = sms.Send(phone, message.NewMessage().
    SetContent("Your verification code is: 6379, valid for 5 minutes."))
```

### [火山引擎](https://console.volcengine.com/sms/)

短信内容使用 `template` + `data`

```go
"volcengine": {
    "access_key_id":     "your-access-key-id",     // 平台分配的 access_key_id
    "access_key_secret": "your-access-key-secret", // 平台分配的 access_key_secret
    "region_id":         "cn-north-1",             // 国内节点 cn-north-1，国外节点 ap-singapore-1，不填或填错，默认使用国内节点
    "sign_name":         "your-sign-name",         // 平台上申请的接口短信签名，可不填，发送短信时 data 中指定
    "sms_account":       "your-sms-account",       // 消息组帐号，可不填，发送短信时 data 中指定
},
```

发送示例：

```go
// 示例1：基本使用
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("SMS_123456").
    SetData(map[string]any{
        "code": "1234", // 模板变量
    }))

// 示例2：高级使用
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("SMS_123456").
    SetData(map[string]any{
        "template_param": map[string]any{"code": "1234"}, // 模板变量参数
        "sign_name":      "your-sign-name",               // 签名，覆盖配置文件中的 sign_name
        "sms_account":    "your-sms-account",             // 消息组帐号，覆盖配置文件中的 sms_account
        "phone_numbers":  "13800138000,13900139000",      // 手机号，批量发送，英文逗号连接多个手机号
        "tag":            "your-tag",                     // 标签，可选
    }))
```



### [UE35](http://uesms.ue35.cn/)

短信内容使用 `content`

```go
"ue35": {
    "username": "your-username", // 用户名
    "userpwd":  "your-password", // 密码
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetContent("您的验证码为: 6379，有效期为5分钟。"))
```

### [网易云信](https://yunxin.163.com/sms)

短信内容使用 `template` + `data`

```go
"yunxin": {
    "app_key":     "your-app-key",     // 应用 AppKey
    "app_secret":  "your-app-secret",  // 应用 AppSecret
    "code_length": "4",                // 随机验证码长度，范围 4～10，默认为 4
    "need_up":     "false",            // 是否需要支持短信上行，默认为 false
},
```

发送示例：

```go
// 发送验证码
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("SMS_001").
    SetData(map[string]any{
        "code":      "8946",       // 如果设置了该参数，则 code_length 参数无效
        "device_id": "device-id",  // 设备 ID，可选
        "action":    "sendCode",   // 默认为 sendCode，校验短信验证码使用 verifyCode
    }))

// 校验验证码
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetData(map[string]any{
        "action": "verifyCode",
        "code":   "8946",
    }))

// 通知模板短信
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("templateid").
    SetData(map[string]any{
        "action": "sendTemplate",
        "params": []string{"param1", "param2"}, // 短信参数列表，用于依次填充模板
    }))
```

### [云之讯](https://www.ucpaas.com/index.html)

短信内容使用 `template` + `data`

```go
"yunzhixun": {
    "sid":    "your-sid",    // SID
    "token":  "your-token",  // 令牌
    "app_id": "your-app-id", // 应用 ID
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetTemplate("SMS_001").
    SetData(map[string]any{
        "params":  "8946,3",                    // 模板参数，多个参数使用逗号分割，模板无参数时可为空
        "uid":     "user-id",                   // 用户 ID，随状态报告返回，可为空
        "mobiles": "13800138000,13900139000",   // 批量发送短信，手机号使用逗号分割，不使用批量发送请不要设置该参数
    }))
```



### [移动云MAS（黑名单模式）](https://mas.10086.cn/)

短信内容使用 `content`

```go
"yidongmasblack": {
    "ecName":    "your-ec-name",    // 机构名称
    "secretKey": "your-secret-key", // 密钥
    "apId":      "your-ap-id",      // 应用 ID
    "sign":      "your-sign",       // 签名
    "addSerial": "",                // 通道号，默认为空
},
```

发送示例：

```go
results, err := sms.Send(message.NewPhoneNumber("13800138000"), message.NewMessage().
    SetContent("您的验证码为: 6379，有效期为5分钟。"))
```

## 性能和优化

### 高性能网关注册机制

go-easy-sms 采用了优化的网关注册机制，具有以下特点：

- **线程安全**：使用 `sync.RWMutex` 保护并发访问，支持高并发场景
- **零内存分配**：优化后的实现避免了不必要的内存分配
- **极低延迟**：网关访问延迟约 15ns/op，并发访问约 25ns/op
- **智能缓存**：网关实例创建后自动缓存，避免重复创建


### 并发安全

所有网关操作都是线程安全的，可以在高并发环境中安全使用：

```go
// 并发发送示例
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        phone := fmt.Sprintf("1380013%04d", id)
        results, err := sms.Send(message.NewPhoneNumber(phone),
            message.NewMessage().SetContent("并发测试消息"))
        if err != nil {
            log.Printf("发送失败: %v", err)
        }
        // 处理结果...
    }(i)
}
wg.Wait()
```

## 单元测试

所有测试文件都放在 `tests` 目录下，按照功能模块分类。运行单元测试：

```bash
go test ./tests/...
```


## 致谢

go-easy-sms 项目是参考 PHP 版本的 [overtrue/easy-sms](https://github.com/overtrue/easy-sms) 开发的 Go 语言实现。在此特别感谢 [overtrue](https://github.com/overtrue) 创建的优秀项目，为我们提供了清晰的设计思路和实现参考。

go-easy-sms 保持了与 PHP 版本相似的 API 设计和使用体验，同时充分利用了 Go 语言的特性进行了适当的优化和调整。

## 许可证

MIT
