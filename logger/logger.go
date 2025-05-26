package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// LogLevel 定义日志级别
type LogLevel int

const (
	// DEBUG 调试级别
	DEBUG LogLevel = iota
	// INFO 信息级别
	INFO
	// WARNING 警告级别
	WARNING
	// ERROR 错误级别
	ERROR
	// FATAL 致命错误级别
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG:   "DEBUG",
	INFO:    "INFO",
	WARNING: "WARNING",
	ERROR:   "ERROR",
	FATAL:   "FATAL",
}

// Logger 是日志记录器
type Logger struct {
	mu       sync.Mutex
	level    LogLevel
	logger   *log.Logger
	disabled bool
}

var (
	// 默认日志记录器
	defaultLogger = NewLogger(os.Stdout, INFO)
)

// NewLogger 创建一个新的日志记录器
func NewLogger(out io.Writer, level LogLevel) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(out, "", log.LstdFlags),
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput 设置日志输出
func (l *Logger) SetOutput(out io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.SetOutput(out)
}

// GetOutput 获取日志输出
func (l *Logger) GetOutput() io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.logger.Writer()
}

// Disable 禁用日志
func (l *Logger) Disable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.disabled = true
}

// Enable 启用日志
func (l *Logger) Enable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.disabled = false
}

// log 记录日志
func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	if l.disabled || level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, v...)
	l.logger.Printf("[%s] %s", levelNames[level], msg)
}

// Debug 记录调试级别日志
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(DEBUG, format, v...)
}

// Info 记录信息级别日志
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(INFO, format, v...)
}

// Warning 记录警告级别日志
func (l *Logger) Warning(format string, v ...interface{}) {
	l.log(WARNING, format, v...)
}

// Error 记录错误级别日志
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ERROR, format, v...)
}

// Fatal 记录致命错误级别日志
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(FATAL, format, v...)
	os.Exit(1)
}

// 全局函数

// SetLevel 设置默认日志记录器的级别
func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// SetOutput 设置默认日志记录器的输出
func SetOutput(out io.Writer) {
	defaultLogger.SetOutput(out)
}

// Disable 禁用默认日志记录器
func Disable() {
	defaultLogger.Disable()
}

// Enable 启用默认日志记录器
func Enable() {
	defaultLogger.Enable()
}

// Debug 使用默认日志记录器记录调试级别日志
func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

// Info 使用默认日志记录器记录信息级别日志
func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

// Warning 使用默认日志记录器记录警告级别日志
func Warning(format string, v ...interface{}) {
	defaultLogger.Warning(format, v...)
}

// Error 使用默认日志记录器记录错误级别日志
func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

// Fatal 使用默认日志记录器记录致命错误级别日志
func Fatal(format string, v ...interface{}) {
	defaultLogger.Fatal(format, v...)
}

// GetLogger 获取默认日志记录器
func GetLogger() *Logger {
	return defaultLogger
}
