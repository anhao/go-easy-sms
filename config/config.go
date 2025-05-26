package config

import (
	"github.com/anhao/go-easy-sms/strategy"
)

// Config 是短信服务的配置结构
type Config struct {
	// 超时时间（秒）
	Timeout float64

	// 默认策略
	Strategy strategy.Strategy

	// 默认可用的网关
	DefaultGateways []string

	// 网关配置
	GatewayConfigs map[string]map[string]any
}

// NewConfig 创建一个新的配置实例
func NewConfig() *Config {
	return &Config{
		Timeout:         5.0,
		DefaultGateways: []string{},
		GatewayConfigs:  make(map[string]map[string]any),
	}
}
