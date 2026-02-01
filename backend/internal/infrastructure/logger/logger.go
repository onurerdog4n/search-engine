package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap.Logger
type Logger struct {
	*zap.Logger
}

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	Encoding   string // json or console
	OutputPath string // stdout, stderr, or file path
}

// NewLogger creates a new logger instance
func NewLogger(cfg Config) (*Logger, error) {
	// Parse log level
	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	// Configure encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Choose encoder
	var encoder zapcore.Encoder
	if cfg.Encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Configure output
	var output zapcore.WriteSyncer
	if cfg.OutputPath == "" || cfg.OutputPath == "stdout" {
		output = zapcore.AddSync(os.Stdout)
	} else if cfg.OutputPath == "stderr" {
		output = zapcore.AddSync(os.Stderr)
	} else {
		file, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		output = zapcore.AddSync(file)
	}

	// Create core
	core := zapcore.NewCore(encoder, output, level)

	// Create logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{Logger: zapLogger}, nil
}

// NewDevelopmentLogger creates a logger for development
func NewDevelopmentLogger() (*Logger, error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return &Logger{Logger: zapLogger}, nil
}

// NewProductionLogger creates a logger for production
func NewProductionLogger() (*Logger, error) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &Logger{Logger: zapLogger}, nil
}

// WithRequestID adds request ID to logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.With(zap.String("request_id", requestID)),
	}
}

// WithUserID adds user ID to logger
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{
		Logger: l.With(zap.String("user_id", userID)),
	}
}

// WithComponent adds component name to logger
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.With(zap.String("component", component)),
	}
}

// WithError adds error to logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.With(zap.Error(err)),
	}
}

// Global logger instance
var globalLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(cfg Config) error {
	logger, err := NewLogger(cfg)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetLogger returns the global logger
func GetLogger() *Logger {
	if globalLogger == nil {
		// Fallback to development logger
		logger, _ := NewDevelopmentLogger()
		globalLogger = logger
	}
	return globalLogger
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}
