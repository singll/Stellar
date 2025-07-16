package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	Logger zerolog.Logger
)

// Config 日志配置
type Config struct {
	Level      string `yaml:"level"`      // 日志级别
	Format     string `yaml:"format"`     // 输出格式: json, console
	Output     string `yaml:"output"`     // 输出目标: stdout, file
	Filename   string `yaml:"filename"`   // 文件路径
	MaxSize    int    `yaml:"max_size"`   // 最大文件大小(MB)
	MaxBackups int    `yaml:"max_backups"` // 最大备份数
	MaxAge     int    `yaml:"max_age"`    // 最大保留天数
	Compress   bool   `yaml:"compress"`   // 是否压缩
}

// Init 初始化日志系统
func Init(config Config) error {
	// 设置日志级别
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// 配置时间格式
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"

	// 配置输出格式
	var output io.Writer = os.Stdout
	if config.Format == "console" {
		output = zerolog.ConsoleWriter{Out: os.Stdout}
	}

	// 创建logger
	Logger = zerolog.New(output).With().Timestamp().Caller().Logger()

	// 设置全局logger
	log.Logger = Logger

	return nil
}

// Debug 调试日志
func Debug(msg string, fields map[string]interface{}) {
	event := Logger.Debug()
	if fields != nil {
		for k, v := range fields {
			event = event.Interface(k, v)
		}
	}
	event.Msg(msg)
}

// Info 信息日志
func Info(msg string, fields map[string]interface{}) {
	event := Logger.Info()
	if fields != nil {
		for k, v := range fields {
			event = event.Interface(k, v)
		}
	}
	event.Msg(msg)
}

// Warn 警告日志
func Warn(msg string, fields map[string]interface{}) {
	event := Logger.Warn()
	if fields != nil {
		for k, v := range fields {
			event = event.Interface(k, v)
		}
	}
	event.Msg(msg)
}

// Error 错误日志
func Error(msg string, fields map[string]interface{}) {
	event := Logger.Error()
	if fields != nil {
		for k, v := range fields {
			event = event.Interface(k, v)
		}
	}
	event.Msg(msg)
}

// Fatal 致命错误日志
func Fatal(msg string, fields map[string]interface{}) {
	event := Logger.Fatal()
	if fields != nil {
		for k, v := range fields {
			event = event.Interface(k, v)
		}
	}
	event.Msg(msg)
}

// WithField 添加字段
func WithField(key string, value interface{}) zerolog.Logger {
	return Logger.With().Interface(key, value).Logger()
}

// WithFields 添加多个字段
func WithFields(fields map[string]interface{}) zerolog.Logger {
	ctx := Logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return ctx.Logger()
}