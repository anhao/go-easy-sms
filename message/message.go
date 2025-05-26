package message

// MessageType 定义消息类型
type MessageType string

const (
	// TextMessage 文本消息
	TextMessage MessageType = "text"
	// VoiceMessage 语音消息
	VoiceMessage MessageType = "voice"
)

// Message 表示短信消息
type Message struct {
	// 消息类型
	Type MessageType

	// 消息内容
	Content string

	// 模板ID
	Template string

	// 模板数据
	Data map[string]any

	// 支持的网关
	Gateways []string
}

// NewMessage 创建一个新的消息
func NewMessage() *Message {
	return &Message{
		Type:     TextMessage,
		Data:     make(map[string]any),
		Gateways: []string{},
	}
}

// SetContent 设置消息内容
func (m *Message) SetContent(content string) *Message {
	m.Content = content
	return m
}

// GetContent 获取消息内容
func (m *Message) GetContent() string {
	return m.Content
}

// SetTemplate 设置模板ID
func (m *Message) SetTemplate(template string) *Message {
	m.Template = template
	return m
}

// GetTemplate 获取模板ID
func (m *Message) GetTemplate() string {
	return m.Template
}

// SetData 设置模板数据
func (m *Message) SetData(data map[string]any) *Message {
	m.Data = data
	return m
}

// GetData 获取模板数据
func (m *Message) GetData() map[string]any {
	return m.Data
}

// SetGateways 设置支持的网关
func (m *Message) SetGateways(gateways []string) *Message {
	m.Gateways = gateways
	return m
}

// GetGateways 获取支持的网关
func (m *Message) GetGateways() []string {
	return m.Gateways
}

// SetType 设置消息类型
func (m *Message) SetType(messageType MessageType) *Message {
	m.Type = messageType
	return m
}

// GetType 获取消息类型
func (m *Message) GetType() MessageType {
	return m.Type
}
