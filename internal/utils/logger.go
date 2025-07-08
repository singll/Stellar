package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogLevel 定义日志级别
type LogLevel string

const (
	// DebugLevel 定义调试级别
	DebugLevel LogLevel = "debug"
	// InfoLevel 定义信息级别
	InfoLevel LogLevel = "info"
	// WarnLevel 定义警告级别
	WarnLevel LogLevel = "warn"
	// ErrorLevel 定义错误级别
	ErrorLevel LogLevel = "error"
	// FatalLevel 定义致命错误级别
	FatalLevel LogLevel = "fatal"
	// PanicLevel 定义恐慌级别
	PanicLevel LogLevel = "panic"
)

// LogConfig 定义日志配置
type LogConfig struct {
	// Level 日志级别
	Level LogLevel
	// EnableConsole 是否输出到控制台
	EnableConsole bool
	// EnableFile 是否输出到文件
	EnableFile bool
	// FilePath 日志文件路径
	FilePath string
	// MaxSize 单个日志文件最大大小（MB）
	MaxSize int
	// MaxBackups 最大备份文件数
	MaxBackups int
	// MaxAge 最大保留天数
	MaxAge int
	// Compress 是否压缩
	Compress bool
	// TimeFormat 时间格式
	TimeFormat string
}

// DefaultLogConfig 返回默认日志配置
func DefaultLogConfig() LogConfig {
	return LogConfig{
		Level:         InfoLevel,
		EnableConsole: true,
		EnableFile:    true,
		FilePath:      "logs/stellar.log",
		MaxSize:       100,
		MaxBackups:    10,
		MaxAge:        30,
		Compress:      true,
		TimeFormat:    time.RFC3339,
	}
}

// Logger 定义日志记录器
type Logger struct {
	logger zerolog.Logger
	config LogConfig
}

// NewLogger 创建新的日志记录器
func NewLogger(config LogConfig) (*Logger, error) {
	// 设置日志级别
	var level zerolog.Level
	switch config.Level {
	case DebugLevel:
		level = zerolog.DebugLevel
	case InfoLevel:
		level = zerolog.InfoLevel
	case WarnLevel:
		level = zerolog.WarnLevel
	case ErrorLevel:
		level = zerolog.ErrorLevel
	case FatalLevel:
		level = zerolog.FatalLevel
	case PanicLevel:
		level = zerolog.PanicLevel
	default:
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = config.TimeFormat

	// 创建输出写入器
	var writers []io.Writer

	// 控制台输出
	if config.EnableConsole {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: config.TimeFormat,
		}
		writers = append(writers, consoleWriter)
	}

	// 文件输出
	if config.EnableFile {
		// 确保日志目录存在
		logDir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("无法创建日志目录: %v", err)
		}

		// 创建或打开日志文件
		logFile, err := os.OpenFile(config.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("无法打开日志文件: %v", err)
		}

		writers = append(writers, logFile)
	}

	// 创建多输出写入器
	var writer io.Writer
	if len(writers) == 1 {
		writer = writers[0]
	} else {
		writer = zerolog.MultiLevelWriter(writers...)
	}

	// 创建日志记录器
	logger := zerolog.New(writer).With().Timestamp().Logger()

	return &Logger{
		logger: logger,
		config: config,
	}, nil
}

// withCaller 添加调用者信息
func (l *Logger) withCaller() zerolog.Logger {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		file = filepath.Base(file)
		return l.logger.With().Str("file", fmt.Sprintf("%s:%d", file, line)).Logger()
	}
	return l.logger
}

// Debug 记录调试级别日志
func (l *Logger) Debug(msg string, fields ...interface{}) {
	logger := l.withCaller()
	event := logger.Debug()
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Info 记录信息级别日志
func (l *Logger) Info(msg string, fields ...interface{}) {
	logger := l.withCaller()
	event := logger.Info()
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Warn 记录警告级别日志
func (l *Logger) Warn(msg string, fields ...interface{}) {
	logger := l.withCaller()
	event := logger.Warn()
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Error 记录错误级别日志
func (l *Logger) Error(msg string, err error, fields ...interface{}) {
	logger := l.withCaller()
	event := logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Fatal 记录致命错误级别日志
func (l *Logger) Fatal(msg string, err error, fields ...interface{}) {
	logger := l.withCaller()
	event := logger.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Panic 记录恐慌级别日志
func (l *Logger) Panic(msg string, err error, fields ...interface{}) {
	logger := l.withCaller()
	event := logger.Panic()
	if err != nil {
		event = event.Err(err)
	}
	l.addFields(event, fields...)
	event.Msg(msg)
}

// addFields 添加字段到日志事件
func (l *Logger) addFields(event *zerolog.Event, fields ...interface{}) {
	if len(fields)%2 != 0 {
		event.Str("INVALID_FIELDS", "日志字段必须是键值对")
		return
	}

	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			event.Str("INVALID_KEY", fmt.Sprintf("%v", fields[i]))
			continue
		}

		switch value := fields[i+1].(type) {
		case string:
			event.Str(key, value)
		case int:
			event.Int(key, value)
		case int64:
			event.Int64(key, value)
		case float64:
			event.Float64(key, value)
		case bool:
			event.Bool(key, value)
		case []string:
			event.Strs(key, value)
		case []int:
			event.Ints(key, value)
		case time.Time:
			event.Time(key, value)
		case time.Duration:
			event.Dur(key, value)
		case error:
			event.Err(value)
		default:
			event.Interface(key, value)
		}
	}
}

// WithComponent 返回带有组件标识的日志记录器
func (l *Logger) WithComponent(component string) *Logger {
	newLogger := l.logger.With().Str("component", component).Logger()
	return &Logger{
		logger: newLogger,
		config: l.config,
	}
}

// WithRequestID 返回带有请求ID的日志记录器
func (l *Logger) WithRequestID(requestID string) *Logger {
	newLogger := l.logger.With().Str("request_id", requestID).Logger()
	return &Logger{
		logger: newLogger,
		config: l.config,
	}
}

// WithUser 返回带有用户信息的日志记录器
func (l *Logger) WithUser(userID string) *Logger {
	newLogger := l.logger.With().Str("user_id", userID).Logger()
	return &Logger{
		logger: newLogger,
		config: l.config,
	}
}

// FormatStackTrace 格式化堆栈跟踪
func FormatStackTrace(skip int) string {
	var stackTrace strings.Builder
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		name := fn.Name()
		if strings.Contains(name, "runtime.") {
			continue
		}
		stackTrace.WriteString(fmt.Sprintf("%s:%d %s\n", filepath.Base(file), line, name))
	}
	return stackTrace.String()
}

// GlobalLogger 全局日志记录器
var GlobalLogger *Logger

// InitGlobalLogger 初始化全局日志记录器
func InitGlobalLogger(config LogConfig) error {
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}
	GlobalLogger = logger
	return nil
}

// Debug 全局调试级别日志
func Debug(msg string, fields ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Debug(msg, fields...)
	} else {
		log.Debug().Msg(msg)
	}
}

// Info 全局信息级别日志
func Info(msg string, fields ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Info(msg, fields...)
	} else {
		log.Info().Msg(msg)
	}
}

// Warn 全局警告级别日志
func Warn(msg string, fields ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Warn(msg, fields...)
	} else {
		log.Warn().Msg(msg)
	}
}

// Error 全局错误级别日志
func Error(msg string, err error, fields ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Error(msg, err, fields...)
	} else {
		log.Error().Err(err).Msg(msg)
	}
}

// Fatal 全局致命错误级别日志
func Fatal(msg string, err error, fields ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Fatal(msg, err, fields...)
	} else {
		log.Fatal().Err(err).Msg(msg)
	}
}

// Panic 全局恐慌级别日志
func Panic(msg string, err error, fields ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Panic(msg, err, fields...)
	} else {
		log.Panic().Err(err).Msg(msg)
	}
}
