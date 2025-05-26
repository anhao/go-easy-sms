package easysms

import (
	"errors"
	"fmt"
	"sync"

	"github.com/anhao/go-easy-sms/config"
	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/logger"
	"github.com/anhao/go-easy-sms/message"
	"github.com/anhao/go-easy-sms/strategy"
)

// 状态常量
const (
	StatusSuccess = "success"
	StatusFailure = "failure"
)

// Result 表示发送短信的结果
type Result struct {
	Gateway string
	Status  string
	Data    any
	Error   error
}

// GatewayError 定义网关相关的错误类型（优先级4：错误处理优化）
type GatewayError struct {
	GatewayName string
	Operation   string
	Err         error
}

func (e *GatewayError) Error() string {
	return fmt.Sprintf("gateway %s %s failed: %v", e.GatewayName, e.Operation, e.Err)
}

func (e *GatewayError) Unwrap() error {
	return e.Err
}

// NewGatewayCreator 新的网关创建函数类型（优先级2：性能优化）
type NewGatewayCreator func(config map[string]any) (gateway.Gateway, error)

// GatewayRegistry 网关注册表（优先级2：性能优化）
type GatewayRegistry struct {
	creators map[string]NewGatewayCreator
	mu       sync.RWMutex
}

// NewGatewayRegistry 创建新的网关注册表
func NewGatewayRegistry() *GatewayRegistry {
	return &GatewayRegistry{
		creators: make(map[string]NewGatewayCreator),
	}
}

// Register 注册网关创建函数
func (r *GatewayRegistry) Register(name string, creator NewGatewayCreator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.creators[name] = creator
}

// Create 创建网关实例
func (r *GatewayRegistry) Create(name string, config map[string]any) (gateway.Gateway, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if creator, ok := r.creators[name]; ok {
		return creator(config)
	}

	return nil, &GatewayError{
		GatewayName: name,
		Operation:   "creation",
		Err:         errors.New("creator not found"),
	}
}

// HasCreator 检查是否存在指定名称的创建函数
func (r *GatewayRegistry) HasCreator(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.creators[name]
	return exists
}

// EasySms 是短信服务的主结构（优化后）
type EasySms struct {
	config   *config.Config
	gateways map[string]gateway.Gateway
	strategy strategy.Strategy
	registry *GatewayRegistry
	logger   *logger.Logger
	mu       sync.RWMutex // 优先级1：线程安全保护
}

// New 创建一个新的 EasySms 实例
func New(cfg *config.Config) *EasySms {
	if cfg == nil {
		cfg = config.NewConfig()
	}

	// 如果没有指定策略，使用默认的顺序策略
	if cfg.Strategy == nil {
		cfg.Strategy = strategy.NewOrderStrategy()
	}

	sms := &EasySms{
		config:   cfg,
		gateways: make(map[string]gateway.Gateway),
		strategy: cfg.Strategy,
		registry: NewGatewayRegistry(),
		logger:   logger.GetLogger(),
	}

	// 注册内置网关创建函数
	sms.registerBuiltinGatewayCreators()

	// 自动注册配置中的网关
	sms.autoRegisterGateways()

	return sms
}

// SetLogger 设置日志记录器
func (e *EasySms) SetLogger(l *logger.Logger) {
	e.logger = l
}

// RegisterGateway 注册一个网关实例（线程安全）
func (e *EasySms) RegisterGateway(name string, gw gateway.Gateway) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.logger.Debug("Registering gateway: %s", name)
	e.gateways[name] = gw
}

// RegisterGatewayCreator 注册网关创建函数（新接口）
func (e *EasySms) RegisterGatewayCreator(name string, creator NewGatewayCreator) {
	e.logger.Debug("Registering gateway creator: %s", name)
	e.registry.Register(name, creator)
}

// Gateway 获取指定名称的网关（线程安全，优化性能）
func (e *EasySms) Gateway(name string) (gateway.Gateway, error) {
	// 首先尝试从缓存中获取（读锁）
	e.mu.RLock()
	if gw, ok := e.gateways[name]; ok {
		e.mu.RUnlock()
		return gw, nil
	}
	e.mu.RUnlock()

	// 检查配置是否存在
	config, hasConfig := e.config.GatewayConfigs[name]
	if !hasConfig {
		return nil, &GatewayError{
			GatewayName: name,
			Operation:   "lookup",
			Err:         errors.New("config not found"),
		}
	}

	// 尝试创建网关
	if e.registry.HasCreator(name) {
		e.logger.Debug("Creating gateway: %s", name)
		gw, err := e.registry.Create(name, config)
		if err != nil {
			return nil, err
		}

		// 缓存创建的网关（写锁）
		e.RegisterGateway(name, gw)
		return gw, nil
	}

	return nil, &GatewayError{
		GatewayName: name,
		Operation:   "lookup",
		Err:         errors.New("gateway not found"),
	}
}

// Send 发送短信
func (e *EasySms) Send(to *message.PhoneNumber, msg *message.Message) (map[string]Result, error) {
	// 如果消息中没有指定网关，使用默认网关
	gateways := msg.GetGateways()
	if len(gateways) == 0 {
		gateways = e.config.DefaultGateways
	}

	if len(gateways) == 0 {
		return nil, errors.New("no gateway available")
	}

	e.logger.Info("Sending message to %s using gateways: %v", to.String(), gateways)

	// 使用策略确定网关顺序
	orderedGateways := e.strategy.Apply(gateways)

	results := make(map[string]Result)
	var lastErr error
	success := false

	// 尝试每个网关，直到一个成功
	for _, gatewayName := range orderedGateways {
		e.logger.Debug("Trying gateway: %s", gatewayName)

		gateway, err := e.Gateway(gatewayName)
		if err != nil {
			e.logger.Error("Gateway %s not available: %v", gatewayName, err)
			results[gatewayName] = Result{
				Gateway: gatewayName,
				Status:  StatusFailure,
				Error:   err,
			}
			lastErr = err
			continue
		}

		// 尝试发送消息
		e.logger.Debug("Sending message via gateway: %s", gatewayName)
		resp, err := gateway.Send(to, msg)
		if err != nil {
			e.logger.Error("Failed to send message via gateway %s: %v", gatewayName, err)
			results[gatewayName] = Result{
				Gateway: gatewayName,
				Status:  StatusFailure,
				Error:   err,
			}
			lastErr = err
			continue
		}

		// 成功
		e.logger.Info("Successfully sent message via gateway: %s", gatewayName)
		results[gatewayName] = Result{
			Gateway: gatewayName,
			Status:  StatusSuccess,
			Data:    resp,
		}
		success = true
		break
	}

	if !success {
		e.logger.Error("All gateways failed: %v", lastErr)
		return results, fmt.Errorf("all gateways failed: %v", lastErr)
	}

	return results, nil
}

// registerBuiltinGatewayCreators 注册内置网关创建函数（优先级3：消除硬编码）
func (e *EasySms) registerBuiltinGatewayCreators() {
	// 使用 map 批量注册，减少重复代码
	builtinCreators := map[string]NewGatewayCreator{
		"aliyun": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewAliyunGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "aliyun",
					Operation:   "creation",
					Err:         errors.New("failed to create aliyun gateway"),
				}
			}
			return gw, nil
		},
		"yunpian": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewYunpianGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "yunpian",
					Operation:   "creation",
					Err:         errors.New("failed to create yunpian gateway"),
				}
			}
			return gw, nil
		},
		"errorlog": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewErrorlogGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "errorlog",
					Operation:   "creation",
					Err:         errors.New("failed to create errorlog gateway"),
				}
			}
			return gw, nil
		},
		"qcloud": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewQcloudGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "qcloud",
					Operation:   "creation",
					Err:         errors.New("failed to create qcloud gateway"),
				}
			}
			return gw, nil
		},
		"chuanglan": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewChuanglanGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "chuanglan",
					Operation:   "creation",
					Err:         errors.New("failed to create chuanglan gateway"),
				}
			}
			return gw, nil
		},
		"ucloud": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewUcloudGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "ucloud",
					Operation:   "creation",
					Err:         errors.New("failed to create ucloud gateway"),
				}
			}
			return gw, nil
		},
		"baidu": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewBaiduGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "baidu",
					Operation:   "creation",
					Err:         errors.New("failed to create baidu gateway"),
				}
			}
			return gw, nil
		},
		"ctyun": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewCtyunGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "ctyun",
					Operation:   "creation",
					Err:         errors.New("failed to create ctyun gateway"),
				}
			}
			return gw, nil
		},
		"huaxin": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewHuaxinGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "huaxin",
					Operation:   "creation",
					Err:         errors.New("failed to create huaxin gateway"),
				}
			}
			return gw, nil
		},
		"submail": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewSubmailGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "submail",
					Operation:   "creation",
					Err:         errors.New("failed to create submail gateway"),
				}
			}
			return gw, nil
		},
		"smsbao": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewSmsbaoGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "smsbao",
					Operation:   "creation",
					Err:         errors.New("failed to create smsbao gateway"),
				}
			}
			return gw, nil
		},
		"aliyun_intl": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewAliyunIntlGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "aliyun_intl",
					Operation:   "creation",
					Err:         errors.New("failed to create aliyun_intl gateway"),
				}
			}
			return gw, nil
		},
		"aliyunrest": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewAliyunrestGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "aliyunrest",
					Operation:   "creation",
					Err:         errors.New("failed to create aliyunrest gateway"),
				}
			}
			return gw, nil
		},
		"chuanglanv1": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewChuanglanv1Gateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "chuanglanv1",
					Operation:   "creation",
					Err:         errors.New("failed to create chuanglanv1 gateway"),
				}
			}
			return gw, nil
		},
		"huyi": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewHuyiGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "huyi",
					Operation:   "creation",
					Err:         errors.New("failed to create huyi gateway"),
				}
			}
			return gw, nil
		},
		"juhe": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewJuheGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "juhe",
					Operation:   "creation",
					Err:         errors.New("failed to create juhe gateway"),
				}
			}
			return gw, nil
		},
		"kingtto": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewKingttoGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "kingtto",
					Operation:   "creation",
					Err:         errors.New("failed to create kingtto gateway"),
				}
			}
			return gw, nil
		},
		"luosimao": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewLuosimaoGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "luosimao",
					Operation:   "creation",
					Err:         errors.New("failed to create luosimao gateway"),
				}
			}
			return gw, nil
		},
		"maap": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewMaapGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "maap",
					Operation:   "creation",
					Err:         errors.New("failed to create maap gateway"),
				}
			}
			return gw, nil
		},
		"moduyun": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewModuyunGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "moduyun",
					Operation:   "creation",
					Err:         errors.New("failed to create moduyun gateway"),
				}
			}
			return gw, nil
		},
		"nowcn": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewNowcnGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "nowcn",
					Operation:   "creation",
					Err:         errors.New("failed to create nowcn gateway"),
				}
			}
			return gw, nil
		},
		"qiniu": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewQiniuGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "qiniu",
					Operation:   "creation",
					Err:         errors.New("failed to create qiniu gateway"),
				}
			}
			return gw, nil
		},
		"rongcloud": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewRongcloudGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "rongcloud",
					Operation:   "creation",
					Err:         errors.New("failed to create rongcloud gateway"),
				}
			}
			return gw, nil
		},
		"rongheyun": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewRongheyunGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "rongheyun",
					Operation:   "creation",
					Err:         errors.New("failed to create rongheyun gateway"),
				}
			}
			return gw, nil
		},
		"sendcloud": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewSendcloudGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "sendcloud",
					Operation:   "creation",
					Err:         errors.New("failed to create sendcloud gateway"),
				}
			}
			return gw, nil
		},
		"twilio": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewTwilioGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "twilio",
					Operation:   "creation",
					Err:         errors.New("failed to create twilio gateway"),
				}
			}
			return gw, nil
		},
		"yuntongxun": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewYuntongxunGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "yuntongxun",
					Operation:   "creation",
					Err:         errors.New("failed to create yuntongxun gateway"),
				}
			}
			return gw, nil
		},
		"volcengine": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewVolcengineGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "volcengine",
					Operation:   "creation",
					Err:         errors.New("failed to create volcengine gateway"),
				}
			}
			return gw, nil
		},
		"ue35": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewUe35Gateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "ue35",
					Operation:   "creation",
					Err:         errors.New("failed to create ue35 gateway"),
				}
			}
			return gw, nil
		},
		"yunxin": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewYunxinGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "yunxin",
					Operation:   "creation",
					Err:         errors.New("failed to create yunxin gateway"),
				}
			}
			return gw, nil
		},
		"yunzhixun": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewYunzhixunGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "yunzhixun",
					Operation:   "creation",
					Err:         errors.New("failed to create yunzhixun gateway"),
				}
			}
			return gw, nil
		},
		"yidongmasblack": func(config map[string]any) (gateway.Gateway, error) {
			gw := gateway.NewYidongmasblackGateway(config)
			if gw == nil {
				return nil, &GatewayError{
					GatewayName: "yidongmasblack",
					Operation:   "creation",
					Err:         errors.New("failed to create yidongmasblack gateway"),
				}
			}
			return gw, nil
		},
	}

	// 批量注册内置网关
	for name, creator := range builtinCreators {
		e.registry.Register(name, creator)
	}
}

// autoRegisterGateways 自动注册配置中的网关（优化版本）
func (e *EasySms) autoRegisterGateways() {
	for name, config := range e.config.GatewayConfigs {
		if e.registry.HasCreator(name) {
			e.logger.Debug("Auto registering gateway: %s", name)
			gw, err := e.registry.Create(name, config)
			if err != nil {
				e.logger.Error("Failed to auto register gateway %s: %v", name, err)
				continue
			}
			e.RegisterGateway(name, gw)
		}
	}
}

// SimpleSend 提供一个简单的发送接口
func (e *EasySms) SimpleSend(phone string, messageData map[string]any) (map[string]Result, error) {
	// 创建电话号码对象
	phoneNumber := message.NewPhoneNumber(phone)

	// 创建消息对象
	msg := message.NewMessage()

	// 处理网关列表
	if gateways, ok := messageData["gateways"]; ok {
		if gatewaysList, ok := gateways.([]string); ok {
			msg.SetGateways(gatewaysList)
		}
	}

	// 处理内容
	if content, ok := messageData["content"]; ok {
		switch c := content.(type) {
		case string:
			msg.SetContent(c)
		case func(string) string:
			// 如果有指定网关，则使用第一个网关来获取内容
			gatewayName := ""
			if len(msg.GetGateways()) > 0 {
				gatewayName = msg.GetGateways()[0]
			} else if len(e.config.DefaultGateways) > 0 {
				gatewayName = e.config.DefaultGateways[0]
			}
			msg.SetContent(c(gatewayName))
		}
	}

	// 处理模板
	if template, ok := messageData["template"]; ok {
		switch t := template.(type) {
		case string:
			msg.SetTemplate(t)
		case func(string) string:
			// 如果有指定网关，则使用第一个网关来获取模板
			gatewayName := ""
			if len(msg.GetGateways()) > 0 {
				gatewayName = msg.GetGateways()[0]
			} else if len(e.config.DefaultGateways) > 0 {
				gatewayName = e.config.DefaultGateways[0]
			}
			msg.SetTemplate(t(gatewayName))
		}
	}

	// 处理数据
	if data, ok := messageData["data"]; ok {
		switch d := data.(type) {
		case map[string]any:
			msg.SetData(d)
		case func(string) map[string]any:
			// 如果有指定网关，则使用第一个网关来获取数据
			gatewayName := ""
			if len(msg.GetGateways()) > 0 {
				gatewayName = msg.GetGateways()[0]
			} else if len(e.config.DefaultGateways) > 0 {
				gatewayName = e.config.DefaultGateways[0]
			}
			msg.SetData(d(gatewayName))
		}
	}

	// 发送消息
	return e.Send(phoneNumber, msg)
}
