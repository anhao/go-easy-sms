package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/anhao/go-easy-sms/logger"
)

func TestLogger(t *testing.T) {
	// 创建一个缓冲区来捕获日志输出
	buf := new(bytes.Buffer)

	// 创建一个新的日志记录器
	log := logger.NewLogger(buf, logger.DEBUG)

	// 测试不同级别的日志
	log.Debug("Debug message")
	if !strings.Contains(buf.String(), "[DEBUG] Debug message") {
		t.Errorf("Expected debug message, got: %s", buf.String())
	}

	buf.Reset()
	log.Info("Info message")
	if !strings.Contains(buf.String(), "[INFO] Info message") {
		t.Errorf("Expected info message, got: %s", buf.String())
	}

	buf.Reset()
	log.Warning("Warning message")
	if !strings.Contains(buf.String(), "[WARNING] Warning message") {
		t.Errorf("Expected warning message, got: %s", buf.String())
	}

	buf.Reset()
	log.Error("Error message")
	if !strings.Contains(buf.String(), "[ERROR] Error message") {
		t.Errorf("Expected error message, got: %s", buf.String())
	}

	// 测试日志级别过滤
	buf.Reset()
	log.SetLevel(logger.WARNING)
	log.Debug("Debug message") // 不应该记录
	log.Info("Info message")   // 不应该记录
	if buf.String() != "" {
		t.Errorf("Expected no output, got: %s", buf.String())
	}

	// 测试禁用日志
	buf.Reset()
	log.Disable()
	log.Warning("Warning message") // 不应该记录
	log.Error("Error message")     // 不应该记录
	if buf.String() != "" {
		t.Errorf("Expected no output, got: %s", buf.String())
	}

	// 测试启用日志
	buf.Reset()
	log.Enable()
	log.Error("Error message")
	if !strings.Contains(buf.String(), "[ERROR] Error message") {
		t.Errorf("Expected error message, got: %s", buf.String())
	}
}

func TestGlobalLoggerFunctions(t *testing.T) {
	// 测试全局日志函数
	// 注意：这些测试可能会影响其他测试，因为它们修改全局状态

	// 我们不需要保存原始输出，因为这只是测试

	// 创建一个缓冲区来捕获日志输出
	buf := new(bytes.Buffer)
	logger.SetOutput(buf)

	// 测试全局日志函数
	logger.SetLevel(logger.DEBUG)
	logger.Debug("Global debug message")
	if !strings.Contains(buf.String(), "[DEBUG] Global debug message") {
		t.Errorf("Expected global debug message, got: %s", buf.String())
	}

	buf.Reset()
	logger.Info("Global info message")
	if !strings.Contains(buf.String(), "[INFO] Global info message") {
		t.Errorf("Expected global info message, got: %s", buf.String())
	}

	buf.Reset()
	logger.Warning("Global warning message")
	if !strings.Contains(buf.String(), "[WARNING] Global warning message") {
		t.Errorf("Expected global warning message, got: %s", buf.String())
	}

	buf.Reset()
	logger.Error("Global error message")
	if !strings.Contains(buf.String(), "[ERROR] Global error message") {
		t.Errorf("Expected global error message, got: %s", buf.String())
	}

	// 测试禁用和启用全局日志
	buf.Reset()
	logger.Disable()
	logger.Error("Should not be logged")
	if buf.String() != "" {
		t.Errorf("Expected no output after disable, got: %s", buf.String())
	}

	buf.Reset()
	logger.Enable()
	logger.Error("Should be logged")
	if !strings.Contains(buf.String(), "[ERROR] Should be logged") {
		t.Errorf("Expected error message after enable, got: %s", buf.String())
	}

	// 恢复原始输出
	// 不需要恢复，因为我们只是在测试中临时修改
}
