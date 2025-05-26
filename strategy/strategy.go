package strategy

import (
	"math/rand"
)

// Strategy 定义了网关选择策略的接口
type Strategy interface {
	// Apply 应用策略，返回排序后的网关列表
	Apply(gateways []string) []string
}

// OrderStrategy 是按顺序调用网关的策略
type OrderStrategy struct{}

// NewOrderStrategy 创建一个新的顺序策略
func NewOrderStrategy() *OrderStrategy {
	return &OrderStrategy{}
}

// Apply 实现 Strategy 接口，按原始顺序返回网关列表
func (s *OrderStrategy) Apply(gateways []string) []string {
	// 直接返回原始顺序
	return gateways
}

// RandomStrategy 是随机选择网关的策略
type RandomStrategy struct{}

// NewRandomStrategy 创建一个新的随机策略
func NewRandomStrategy() *RandomStrategy {
	return &RandomStrategy{}
}

// Apply 实现 Strategy 接口，随机排序网关列表
func (s *RandomStrategy) Apply(gateways []string) []string {
	// 创建一个新的切片，避免修改原始切片
	result := make([]string, len(gateways))
	copy(result, gateways)

	// 使用 Fisher-Yates 洗牌算法随机排序
	for i := len(result) - 1; i > 0; i-- {
		j := int(float64(i+1) * rand.Float64())
		result[i], result[j] = result[j], result[i]
	}

	return result
}
