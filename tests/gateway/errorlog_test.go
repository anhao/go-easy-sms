package gateway

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
	"github.com/anhao/go-easy-sms/message"
)

func TestErrorlogGateway(t *testing.T) {
	// 创建临时日志文件
	logFile := filepath.Join(os.TempDir(), "easy-sms-error-log-test.log")

	// 确保测试结束后删除日志文件
	defer func() {
		if err := os.Remove(logFile); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove test log file: %v", err)
		}
	}()

	// 创建 ErrorlogGateway 实例
	config := map[string]any{
		"file": logFile,
	}
	g := gateway.NewErrorlogGateway(config)

	// 创建测试消息
	phone := message.NewPhoneNumber("13800138000")
	msg := message.NewMessage().
		SetContent("This is a test message.").
		SetData(map[string]any{
			"foo": "bar",
		})

	// 发送消息
	resp, err := g.Send(phone, msg)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 验证返回结果
	respMap, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("Expected response to be map[string]any, got: %T", resp)
	}

	status, ok := respMap["status"].(bool)
	if !ok || !status {
		t.Errorf("Expected status to be true, got: %v", status)
	}

	file, ok := respMap["file"].(string)
	if !ok || file != logFile {
		t.Errorf("Expected file to be %s, got: %s", logFile, file)
	}

	// 验证日志文件内容
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	expectedContent := []string{
		"to: 13800138000",
		"message: \"This is a test message.\"",
		"template: \"\"",
		"data: {\"foo\":\"bar\"}",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Expected log file to contain '%s', but it doesn't", expected)
		}
	}
}

func TestErrorlogGatewayDefaultFile(t *testing.T) {
	// 创建 ErrorlogGateway 实例，不指定文件路径
	config := map[string]any{}
	g := gateway.NewErrorlogGateway(config)

	// 创建测试消息
	phone := message.NewPhoneNumber("13800138000")
	msg := message.NewMessage().
		SetContent("This is a test message.").
		SetData(map[string]any{
			"foo": "bar",
		})

	// 发送消息
	resp, err := g.Send(phone, msg)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 验证返回结果
	respMap, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("Expected response to be map[string]any, got: %T", resp)
	}

	status, ok := respMap["status"].(bool)
	if !ok || !status {
		t.Errorf("Expected status to be true, got: %v", status)
	}

	file, ok := respMap["file"].(string)
	if !ok {
		t.Errorf("Expected file to be a string, got: %T", respMap["file"])
	} else {
		// 清理测试文件
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove test log file: %v", err)
		}
	}
}
