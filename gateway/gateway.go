package gateway

import (
	"context"
	"strconv"
	"time"

	"github.com/anhao/go-easy-sms/http"
	"github.com/anhao/go-easy-sms/message"
)

// Gateway 定义了短信网关的接口
type Gateway interface {
	// GetName 获取网关名称
	GetName() string

	// Send 发送短信
	Send(to *message.PhoneNumber, msg *message.Message) (any, error)
}

// BaseGateway 提供了网关的基本实现
type BaseGateway struct {
	Name       string
	Config     map[string]any
	httpClient *http.Client
}

// NewBaseGateway 创建一个新的基础网关
func NewBaseGateway(name string, config map[string]any) *BaseGateway {
	// 获取超时配置
	timeout := 5.0
	if timeoutVal, ok := config["timeout"]; ok {
		if t, ok := timeoutVal.(float64); ok {
			timeout = t
		}
	}

	// 创建HTTP客户端
	client := http.NewClient(
		http.WithTimeout(time.Duration(timeout) * time.Second),
	)

	return &BaseGateway{
		Name:       name,
		Config:     config,
		httpClient: client,
	}
}

// GetName 获取网关名称
func (g *BaseGateway) GetName() string {
	return g.Name
}

// GetConfig 获取网关配置
func (g *BaseGateway) GetConfig() map[string]any {
	return g.Config
}

// GetConfigString 获取字符串类型的配置项
func (g *BaseGateway) GetConfigString(key string, defaultValue ...string) string {
	if val, ok := g.Config[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// GetConfigInt 获取整数类型的配置项
func (g *BaseGateway) GetConfigInt(key string, defaultValue ...int) int {
	if val, ok := g.Config[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal
			}
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetConfigFloat 获取浮点数类型的配置项
func (g *BaseGateway) GetConfigFloat(key string, defaultValue ...float64) float64 {
	if val, ok := g.Config[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
				return floatVal
			}
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0.0
}

// GetConfigBool 获取布尔类型的配置项
func (g *BaseGateway) GetConfigBool(key string, defaultValue ...bool) bool {
	if val, ok := g.Config[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			switch v {
			case "true", "1", "yes":
				return true
			case "false", "0", "no":
				return false
			}
		case int:
			return v != 0
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

// Get 发送GET请求
func (g *BaseGateway) Get(endpoint string, params map[string]string, headers map[string]string) (map[string]any, error) {
	// 创建上下文，设置超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(g.GetConfigFloat("timeout", 5.0))*time.Second)
	defer cancel()

	// 发送请求
	resp, err := g.httpClient.Get(ctx, endpoint, params, headers)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	result, err := http.ParseJSONResponse(resp)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Post 发送POST请求（表单数据）
func (g *BaseGateway) Post(endpoint string, params map[string]string, headers map[string]string) (map[string]any, error) {
	// 创建上下文，设置超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(g.GetConfigFloat("timeout", 5.0))*time.Second)
	defer cancel()

	// 发送请求
	resp, err := g.httpClient.Post(ctx, endpoint, params, headers)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	result, err := http.ParseJSONResponse(resp)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// PostJSON 发送POST请求（JSON数据）
func (g *BaseGateway) PostJSON(endpoint string, params map[string]any, headers map[string]string) (map[string]any, error) {
	// 创建上下文，设置超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(g.GetConfigFloat("timeout", 5.0))*time.Second)
	defer cancel()

	// 发送请求
	resp, err := g.httpClient.PostJSON(ctx, endpoint, params, headers)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	result, err := http.ParseJSONResponse(resp)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetHTTPClient 获取原始 HTTP 客户端
// 仅在需要特殊定制 HTTP 请求时使用
func (g *BaseGateway) GetHTTPClient() *http.Client {
	return g.httpClient
}
