package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultModule     = "app"
	defaultMaxSize    = 10
	defaultMaxAge     = 7
	defaultMaxBackups = 3
)

var (
	// ErrInvalidLevel indicates that Config.Level is not one of the supported levels.
	ErrInvalidLevel = errors.New("invalid log level")

	// ErrInvalidFormat indicates that Config.Format is not one of the supported formats.
	ErrInvalidFormat = errors.New("invalid log format")

	// ErrFilePathRequired indicates that file output is enabled without a log directory.
	ErrFilePathRequired = errors.New("log file path is required")
)

// Level is the minimum severity accepted by the logger.
type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
)

// Format controls how log entries are encoded.
type Format string

const (
	ConsoleFormat Format = "console"
	JSONFormat    Format = "json"
)

// Config describes a logger instance. The zero value is valid and uses
// library-friendly defaults: console output, info level, and no file output.
type Config struct {
	Module string       `mapstructure:"module" yaml:"module" json:"module"`
	Level  Level        `mapstructure:"level" yaml:"level" json:"level"`
	Format Format       `mapstructure:"format" yaml:"format" json:"format"`
	Output OutputConfig `mapstructure:"output" yaml:"output" json:"output"`
}

// OutputConfig groups all logger output targets.
type OutputConfig struct {
	Console ConsoleOutputConfig `mapstructure:"console" yaml:"console" json:"console"`
	File    FileOutputConfig    `mapstructure:"file" yaml:"file" json:"file"`
}

// ConsoleOutputConfig controls stdout logging.
type ConsoleOutputConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
}

// FileOutputConfig controls lumberjack-backed file logging.
type FileOutputConfig struct {
	Enabled    bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Path       string `mapstructure:"path" yaml:"path" json:"path"`
	MaxSize    int    `mapstructure:"max_size" yaml:"max_size" json:"max_size"`
	MaxAge     int    `mapstructure:"max_age" yaml:"max_age" json:"max_age"`
	MaxBackups int    `mapstructure:"max_backups" yaml:"max_backups" json:"max_backups"`
	Compress   bool   `mapstructure:"compress" yaml:"compress" json:"compress"`
	LocalTime  bool   `mapstructure:"local_time" yaml:"local_time" json:"local_time"`
}

// DefaultConfig returns a safe default logger configuration for reusable packages.
func DefaultConfig() Config {
	return Config{
		Module: defaultModule,
		Level:  InfoLevel,
		Format: ConsoleFormat,
		Output: OutputConfig{
			Console: ConsoleOutputConfig{
				Enabled: true,
			},
			File: FileOutputConfig{
				MaxSize:    defaultMaxSize,
				MaxAge:     defaultMaxAge,
				MaxBackups: defaultMaxBackups,
				Compress:   true,
				LocalTime:  true,
			},
		},
	}
}

// NewDefault creates a logger with DefaultConfig.
func NewDefault() (*zap.Logger, error) {
	return New(DefaultConfig())
}

// New creates a zap logger from config.
func New(config Config) (*zap.Logger, error) {
	config = normalizeConfig(config)

	level, err := parseLevel(config.Level)
	if err != nil {
		return nil, err
	}
	if err := validateFormat(config.Format); err != nil {
		return nil, err
	}

	encoderConfig := newEncoderConfig()
	var cores []zapcore.Core

	if config.Output.File.Enabled {
		if strings.TrimSpace(config.Output.File.Path) == "" {
			return nil, ErrFilePathRequired
		}

		fileCore, err := newFileCore(config, encoderConfig, level)
		if err != nil {
			return nil, err
		}
		cores = append(cores, fileCore)
	}

	if config.Output.Console.Enabled {
		cores = append(cores, zapcore.NewCore(
			newEncoder(config.Format, config.Module, encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		))
	}

	options := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}
	if config.Format == JSONFormat {
		options = append(options, zap.Fields(zap.String("module", config.Module)))
	}

	return zap.New(zapcore.NewTee(cores...), options...), nil
}

func normalizeConfig(config Config) Config {
	config.Module = strings.TrimSpace(config.Module)
	if config.Module == "" {
		config.Module = defaultModule
	}
	if config.Level == "" {
		config.Level = InfoLevel
	}
	config.Level = Level(strings.ToLower(strings.TrimSpace(string(config.Level))))
	if config.Format == "" {
		config.Format = ConsoleFormat
	}
	config.Format = Format(strings.ToLower(strings.TrimSpace(string(config.Format))))
	if !config.Output.Console.Enabled && !config.Output.File.Enabled {
		config.Output.Console.Enabled = true
	}
	if config.Output.File.MaxSize <= 0 {
		config.Output.File.MaxSize = defaultMaxSize
	}
	if config.Output.File.MaxAge <= 0 {
		config.Output.File.MaxAge = defaultMaxAge
	}
	if config.Output.File.MaxBackups <= 0 {
		config.Output.File.MaxBackups = defaultMaxBackups
	}

	return config
}

func parseLevel(level Level) (zapcore.Level, error) {
	switch Level(strings.ToLower(string(level))) {
	case DebugLevel:
		return zapcore.DebugLevel, nil
	case InfoLevel:
		return zapcore.InfoLevel, nil
	case WarnLevel:
		return zapcore.WarnLevel, nil
	case ErrorLevel:
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("%w: %s", ErrInvalidLevel, level)
	}
}

func validateFormat(format Format) error {
	switch Format(strings.ToLower(string(format))) {
	case ConsoleFormat, JSONFormat:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidFormat, format)
	}
}

func newFileCore(config Config, encoderConfig zapcore.EncoderConfig, level zapcore.Level) (zapcore.Core, error) {
	logDir := filepath.Join(config.Output.File.Path, config.Module)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(logDir, config.Module+".log"),
		MaxSize:    config.Output.File.MaxSize,
		MaxAge:     config.Output.File.MaxAge,
		MaxBackups: config.Output.File.MaxBackups,
		Compress:   config.Output.File.Compress,
		LocalTime:  config.Output.File.LocalTime,
	})

	return zapcore.NewCore(
		newEncoder(config.Format, config.Module, encoderConfig),
		writer,
		level,
	), nil
}

func newEncoder(format Format, module string, config zapcore.EncoderConfig) zapcore.Encoder {
	if format == JSONFormat {
		return zapcore.NewJSONEncoder(config)
	}

	config.EncodeLevel = func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("[%s] %s", module, level.CapitalString()))
	}

	return zapcore.NewConsoleEncoder(config)
}

func newEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     encodeTime,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func encodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.DateTime))
}
