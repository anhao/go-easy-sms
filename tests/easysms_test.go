package tests

import (
	"errors"
	"sync"
	"testing"

	easysms "github.com/anhao/go-easy-sms"
	"github.com/anhao/go-easy-sms/config"
	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
)

// 创建一个模拟网关用于测试（优化后版本）
type MockGateway struct {
	name       string
	config     map[string]any
	shouldFail bool
}

func NewMockGateway(config map[string]any, shouldFail bool) *MockGateway {
	return &MockGateway{
		name:       "mock",
		config:     config,
		shouldFail: shouldFail,
	}
}

func (g *MockGateway) GetName() string {
	return g.name
}

func (g *MockGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	if g.shouldFail {
		return nil, errors.New("mock gateway error")
	}
	return map[string]any{
		"success":    true,
		"message_id": "mock_12345",
		"gateway":    g.name,
	}, nil
}

func TestEasySms(t *testing.T) {
	// 创建配置
	cfg := config.NewConfig()
	cfg.DefaultGateways = []string{"mock"}
	cfg.GatewayConfigs = map[string]map[string]any{
		"mock": {"timeout": 5.0},
	}

	// 创建 EasySms 实例
	sms := easysms.New(cfg)

	// 注册模拟网关
	sms.RegisterGateway("mock", NewMockGateway(cfg.GatewayConfigs["mock"], false))

	// 创建消息
	phone := message.NewPhoneNumber("13800138000")
	msg := message.NewMessage()
	msg.SetContent("测试消息")

	// 发送消息
	results, err := sms.Send(phone, msg)
	if err != nil {
		t.Fatalf("发送失败: %v", err)
	}

	// 验证结果
	if len(results) != 1 {
		t.Errorf("期望1个结果，得到%d个", len(results))
	}

	result, ok := results["mock"]
	if !ok {
		t.Error("期望有mock网关的结果")
	}

	if result.Status != easysms.StatusSuccess {
		t.Errorf("期望状态为%s，得到%s", easysms.StatusSuccess, result.Status)
	}
}

func TestEasySmsWithFailingGateway(t *testing.T) {
	// 创建配置
	cfg := config.NewConfig()
	cfg.DefaultGateways = []string{"failing"}
	cfg.GatewayConfigs = map[string]map[string]any{
		"failing": {"timeout": 5.0},
	}

	// 创建 EasySms 实例
	sms := easysms.New(cfg)

	// 注册失败的模拟网关
	sms.RegisterGateway("failing", NewMockGateway(cfg.GatewayConfigs["failing"], true))

	// 创建消息
	phone := message.NewPhoneNumber("13800138000")
	msg := message.NewMessage()
	msg.SetContent("测试消息")

	// 发送消息
	results, err := sms.Send(phone, msg)
	if err == nil {
		t.Error("期望发送失败")
	}

	// 验证结果
	if len(results) != 1 {
		t.Errorf("期望1个结果，得到%d个", len(results))
	}

	result, ok := results["failing"]
	if !ok {
		t.Error("期望有failing网关的结果")
	}

	if result.Status != easysms.StatusFailure {
		t.Errorf("期望状态为%s，得到%s", easysms.StatusFailure, result.Status)
	}
}

// 测试自定义网关创建函数（使用新的优化接口）
func TestCustomGatewayCreator(t *testing.T) {
	cfg := config.NewConfig()
	cfg.DefaultGateways = []string{"custom"}
	cfg.GatewayConfigs = map[string]map[string]any{
		"custom": {"key": "value"},
	}

	sms := easysms.New(cfg)

	// 注册自定义网关创建函数（使用新的优化接口）
	sms.RegisterGatewayCreator("custom", func(config map[string]any) (gateway.Gateway, error) {
		gw := NewMockGateway(config, false)
		if gw == nil {
			return nil, errors.New("failed to create mock gateway")
		}
		return gw, nil
	})

	// 测试获取网关
	gw, err := sms.Gateway("custom")
	if err != nil {
		t.Fatalf("获取自定义网关失败: %v", err)
	}

	if gw.GetName() != "mock" {
		t.Errorf("期望网关名称为mock，得到%s", gw.GetName())
	}

	// 测试发送消息
	phone := message.NewPhoneNumber("13800138000")
	msg := message.NewMessage().SetContent("自定义网关测试")

	results, err := sms.Send(phone, msg)
	if err != nil {
		t.Fatalf("通过自定义网关发送失败: %v", err)
	}

	result, ok := results["custom"]
	if !ok {
		t.Error("期望有custom网关的结果")
	}

	if result.Status != easysms.StatusSuccess {
		t.Errorf("期望状态为%s，得到%s", easysms.StatusSuccess, result.Status)
	}
}

// 测试错误处理（新的结构化错误）
func TestGatewayError(t *testing.T) {
	cfg := config.NewConfig()
	sms := easysms.New(cfg)

	// 测试获取不存在的网关
	_, err := sms.Gateway("nonexistent")
	if err == nil {
		t.Error("期望获取不存在网关时出错")
	}

	// 验证错误类型
	if gwErr, ok := err.(*easysms.GatewayError); ok {
		if gwErr.GatewayName != "nonexistent" {
			t.Errorf("期望网关名称为nonexistent，得到%s", gwErr.GatewayName)
		}
		if gwErr.Operation != "lookup" {
			t.Errorf("期望操作为lookup，得到%s", gwErr.Operation)
		}
	} else {
		t.Error("期望GatewayError类型")
	}
}

// 测试网关创建错误
func TestGatewayCreationError(t *testing.T) {
	cfg := config.NewConfig()
	cfg.GatewayConfigs = map[string]map[string]any{
		"failing_creator": {"key": "value"},
	}

	sms := easysms.New(cfg)

	// 注册一个会失败的网关创建函数
	sms.RegisterGatewayCreator("failing_creator", func(config map[string]any) (gateway.Gateway, error) {
		return nil, errors.New("creation failed")
	})

	// 测试创建失败的情况
	_, err := sms.Gateway("failing_creator")
	if err == nil {
		t.Error("期望网关创建失败")
	}

	if err.Error() != "creation failed" {
		t.Errorf("期望错误信息为'creation failed'，得到'%s'", err.Error())
	}
}

// 测试并发安全性
func TestConcurrentSafety(t *testing.T) {
	cfg := config.NewConfig()
	cfg.GatewayConfigs = map[string]map[string]any{
		"concurrent_test": {"timeout": 5.0},
	}

	sms := easysms.New(cfg)

	// 注册测试网关
	sms.RegisterGatewayCreator("concurrent_test", func(config map[string]any) (gateway.Gateway, error) {
		return NewMockGateway(config, false), nil
	})

	var wg sync.WaitGroup
	numGoroutines := 10
	iterations := 10

	// 并发访问测试
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_, err := sms.Gateway("concurrent_test")
				if err != nil {
					t.Errorf("Goroutine %d 获取网关失败: %v", id, err)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestSimpleSend(t *testing.T) {
	// 创建配置
	cfg := config.NewConfig()
	cfg.DefaultGateways = []string{"simple"}
	cfg.GatewayConfigs = map[string]map[string]any{
		"simple": {"timeout": 5.0},
	}

	// 创建 EasySms 实例
	sms := easysms.New(cfg)

	// 注册模拟网关
	sms.RegisterGateway("simple", NewMockGateway(cfg.GatewayConfigs["simple"], false))

	// 使用 SimpleSend 发送消息
	results, err := sms.SimpleSend("13800138000", map[string]any{
		"content": "SimpleSend测试消息",
	})

	if err != nil {
		t.Fatalf("SimpleSend发送失败: %v", err)
	}

	// 验证结果
	if len(results) != 1 {
		t.Errorf("期望1个结果，得到%d个", len(results))
	}

	result, ok := results["simple"]
	if !ok {
		t.Error("期望有simple网关的结果")
	}

	if result.Status != easysms.StatusSuccess {
		t.Errorf("期望状态为%s，得到%s", easysms.StatusSuccess, result.Status)
	}
}

func TestMultipleGateways(t *testing.T) {
	// 创建配置
	cfg := config.NewConfig()
	cfg.DefaultGateways = []string{"first", "second"}
	cfg.GatewayConfigs = map[string]map[string]any{
		"first":  {"timeout": 5.0},
		"second": {"timeout": 5.0},
	}

	// 创建 EasySms 实例
	sms := easysms.New(cfg)

	// 注册第一个网关（会失败）
	sms.RegisterGateway("first", NewMockGateway(cfg.GatewayConfigs["first"], true))
	// 注册第二个网关（会成功）
	sms.RegisterGateway("second", NewMockGateway(cfg.GatewayConfigs["second"], false))

	// 创建消息
	phone := message.NewPhoneNumber("13800138000")
	msg := message.NewMessage()
	msg.SetContent("多网关测试消息")

	// 发送消息
	results, err := sms.Send(phone, msg)
	if err != nil {
		t.Fatalf("发送失败: %v", err)
	}

	// 验证结果
	if len(results) != 2 {
		t.Errorf("期望2个结果，得到%d个", len(results))
	}

	// 第一个网关应该失败
	firstResult, ok := results["first"]
	if !ok {
		t.Error("期望有first网关的结果")
	}
	if firstResult.Status != easysms.StatusFailure {
		t.Errorf("期望first网关状态为%s，得到%s", easysms.StatusFailure, firstResult.Status)
	}

	// 第二个网关应该成功
	secondResult, ok := results["second"]
	if !ok {
		t.Error("期望有second网关的结果")
	}
	if secondResult.Status != easysms.StatusSuccess {
		t.Errorf("期望second网关状态为%s，得到%s", easysms.StatusSuccess, secondResult.Status)
	}
}
