package gateway

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anhao/go-easy-sms/logger"
	"github.com/anhao/go-easy-sms/message"
)

// ErrorlogGateway 是一个将短信内容记录到错误日志文件的网关
type ErrorlogGateway struct {
	*BaseGateway
}

// NewErrorlogGateway 创建一个新的 ErrorlogGateway 实例
func NewErrorlogGateway(config map[string]any) *ErrorlogGateway {
	return &ErrorlogGateway{
		BaseGateway: NewBaseGateway("errorlog", config),
	}
}

// Send 将短信内容记录到错误日志文件
func (g *ErrorlogGateway) Send(to *message.PhoneNumber, msg *message.Message) (any, error) {
	// 获取日志文件路径，如果未指定则使用默认路径
	file, ok := g.Config["file"].(string)
	if !ok || file == "" {
		// 如果未指定文件，则使用默认路径
		tempDir := os.TempDir()
		file = filepath.Join(tempDir, "easy-sms-error.log")
	}

	// 格式化消息内容
	content := fmt.Sprintf(
		"[%s] to: %s | message: \"%s\" | template: \"%s\" | data: %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		to.String(),
		msg.GetContent(),
		msg.GetTemplate(),
		jsonEncode(msg.GetData()),
	)

	// 打开文件并写入内容
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logger.Warning("Failed to close error log file: %v", closeErr)
		}
	}()

	// 写入内容
	_, err = f.WriteString(content)
	if err != nil {
		return nil, err
	}

	// 返回结果
	return map[string]any{
		"status": true,
		"file":   file,
	}, nil
}

// jsonEncode 将数据编码为 JSON 字符串
func jsonEncode(data map[string]any) string {
	if data == nil {
		return "{}"
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}

	return string(bytes)
}
