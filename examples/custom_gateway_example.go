package main

import (
	"fmt"
	"log"

	easysms "github.com/anhao/go-easy-sms"
	"github.com/anhao/go-easy-sms/config"
	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
)

// CustomGateway 自定义网关示例
type CustomGateway struct {
	name   string
	config map[string]any
}

// NewCustomGateway 创建自定义网关
func NewCustomGateway(config map[string]any) *CustomGateway {
	return &CustomGateway{
		name:   "custom",
		config: config,
	}
}

// GetName 获取网关名称
func (g *CustomGateway) GetName() string {
	return g.name
}

// Send 发送短信
func (g *CustomGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 这里实现自定义的短信发送逻辑
	fmt.Printf("Custom Gateway: Sending message to %s\n", to.String())
	fmt.Printf("Content: %s\n", msg.GetContent())
	fmt.Printf("Template: %s\n", msg.GetTemplate())
	fmt.Printf("Data: %v\n", msg.GetData())

	// 模拟发送成功
	return map[string]any{
		"message_id": "custom_12345",
		"status":     "sent",
		"gateway":    g.name,
	}, nil
}

// AnotherCustomGateway 另一个自定义网关示例
type AnotherCustomGateway struct {
	name   string
	apiKey string
}

// NewAnotherCustomGateway 创建另一个自定义网关
func NewAnotherCustomGateway(config map[string]any) *AnotherCustomGateway {
	apiKey := ""
	if key, ok := config["api_key"].(string); ok {
		apiKey = key
	}

	return &AnotherCustomGateway{
		name:   "another_custom",
		apiKey: apiKey,
	}
}

// GetName 获取网关名称
func (g *AnotherCustomGateway) GetName() string {
	return g.name
}

// Send 发送短信
func (g *AnotherCustomGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 这里实现另一个自定义的短信发送逻辑
	fmt.Printf("Another Custom Gateway: Sending message to %s with API Key: %s\n", to.String(), g.apiKey)
	fmt.Printf("Content: %s\n", msg.GetContent())

	// 模拟发送成功
	return map[string]any{
		"message_id": "another_custom_67890",
		"status":     "delivered",
		"gateway":    g.name,
		"api_key":    g.apiKey,
	}, nil
}

func main() {
	fmt.Println("=== go-easy-sms 自定义网关注册示例 ===")

	// 创建配置
	cfg := config.NewConfig()

	// 配置自定义网关
	cfg.GatewayConfigs = map[string]map[string]any{
		"custom": {
			"timeout": 30,
		},
		"another_custom": {
			"api_key": "your_api_key_here",
			"timeout": 60,
		},
	}

	// 设置默认网关
	cfg.DefaultGateways = []string{"custom", "another_custom"}

	// 创建 EasySms 实例
	sms := easysms.New(cfg)

	// 注册自定义网关创建函数（新的优化接口）
	fmt.Println("1. 注册自定义网关...")

	// 注册第一个自定义网关
	sms.RegisterGatewayCreator("custom", func(config map[string]any) (gateway.Gateway, error) {
		gw := NewCustomGateway(config)
		if gw == nil {
			return nil, fmt.Errorf("failed to create custom gateway")
		}
		return gw, nil
	})

	// 注册第二个自定义网关
	sms.RegisterGatewayCreator("another_custom", func(config map[string]any) (gateway.Gateway, error) {
		gw := NewAnotherCustomGateway(config)
		if gw == nil {
			return nil, fmt.Errorf("failed to create another custom gateway")
		}
		return gw, nil
	})

	fmt.Println("自定义网关注册完成！")

	// 测试获取网关
	fmt.Println("2. 测试获取自定义网关...")

	customGw, err := sms.Gateway("custom")
	if err != nil {
		log.Fatalf("Failed to get custom gateway: %v", err)
	}
	fmt.Printf("成功获取自定义网关: %s\n", customGw.GetName())

	anotherGw, err := sms.Gateway("another_custom")
	if err != nil {
		log.Fatalf("Failed to get another custom gateway: %v", err)
	}
	fmt.Printf("成功获取另一个自定义网关: %s\n\n", anotherGw.GetName())

	// 测试发送短信
	fmt.Println("3. 测试发送短信...")

	// 创建电话号码
	phone := message.NewPhoneNumber("13800138000")

	// 创建消息
	msg := message.NewMessage()
	msg.SetContent("这是一条测试短信")
	msg.SetTemplate("test_template")
	msg.SetData(map[string]any{
		"name": "张三",
		"code": "123456",
	})

	// 发送短信
	results, err := sms.Send(phone, msg)
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// 显示结果
	fmt.Println("\n发送结果:")
	for gatewayName, result := range results {
		fmt.Printf("网关: %s\n", gatewayName)
		fmt.Printf("状态: %s\n", result.Status)
		if result.Error != nil {
			fmt.Printf("错误: %v\n", result.Error)
		} else {
			fmt.Printf("响应: %v\n", result.Data)
		}
		fmt.Println()
	}

	// 测试 SimpleSend
	fmt.Println("4. 测试 SimpleSend 接口...")

	simpleResults, err := sms.SimpleSend("13900139000", map[string]any{
		"content":  "这是通过 SimpleSend 发送的消息",
		"gateways": []string{"custom"},
	})

	if err != nil {
		log.Fatalf("Failed to send simple message: %v", err)
	}

	fmt.Println("SimpleSend 结果:")
	for gatewayName, result := range simpleResults {
		fmt.Printf("网关: %s, 状态: %s\n", gatewayName, result.Status)
		if result.Data != nil {
			fmt.Printf("响应: %v\n", result.Data)
		}
	}

	fmt.Println("\n=== 示例完成 ===")
}

// 高级自定义网关示例：带有错误处理和配置验证
type AdvancedCustomGateway struct {
	name     string
	endpoint string
	apiKey   string
	timeout  int
}

// NewAdvancedCustomGateway 创建高级自定义网关
func NewAdvancedCustomGateway(config map[string]any) (*AdvancedCustomGateway, error) {
	// 验证必需的配置
	endpoint, ok := config["endpoint"].(string)
	if !ok || endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("api_key is required")
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

// GetName 获取网关名称
func (g *AdvancedCustomGateway) GetName() string {
	return g.name
}

// Send 发送短信
func (g *AdvancedCustomGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 验证输入
	if to == nil {
		return nil, fmt.Errorf("phone number is required")
	}

	if msg == nil {
		return nil, fmt.Errorf("message is required")
	}

	content := msg.GetContent()
	if content == "" {
		return nil, fmt.Errorf("message content is required")
	}

	// 模拟 HTTP 请求
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

// ExampleAdvancedCustomGateway 使用高级自定义网关的示例函数
func ExampleAdvancedCustomGateway() {
	fmt.Println("\n=== 高级自定义网关示例 ===")

	cfg := config.NewConfig()
	cfg.GatewayConfigs = map[string]map[string]any{
		"advanced_custom": {
			"endpoint": "https://api.example.com/sms",
			"api_key":  "your_secret_api_key",
			"timeout":  45,
		},
	}
	cfg.DefaultGateways = []string{"advanced_custom"}

	sms := easysms.New(cfg)

	// 注册高级自定义网关
	sms.RegisterGatewayCreator("advanced_custom", func(config map[string]any) (gateway.Gateway, error) {
		return NewAdvancedCustomGateway(config)
	})

	// 测试发送
	phone := message.NewPhoneNumber("13700137000")
	msg := message.NewMessage().SetContent("高级自定义网关测试消息")

	results, err := sms.Send(phone, msg)
	if err != nil {
		fmt.Printf("发送失败: %v\n", err)
		return
	}

	for gatewayName, result := range results {
		fmt.Printf("网关: %s, 状态: %s\n", gatewayName, result.Status)
		if result.Data != nil {
			fmt.Printf("响应: %v\n", result.Data)
		}
	}
}
