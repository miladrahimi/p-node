package logger

import (
	"syscall"

	"github.com/cockroachdb/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap.Logger.
type Logger struct {
	e        *zap.Logger
	shutdown chan struct{}
	level    string
	format   string
}

// New creates a new instance of Logger.
func New(level, format string, closer chan struct{}) (logger *Logger, err error) {
	l := &Logger{e: nil, shutdown: closer, level: level, format: format}

	logLevel := zap.NewAtomicLevel()
	if err = logLevel.UnmarshalText([]byte(l.level)); err != nil {
		return nil, errors.Wrapf(err, "invalid log level '%s'", l.level)
	}

	l.e, err = zap.Config{
		Level:             logLevel,
		Development:       false,
		Encoding:          "json",
		DisableStacktrace: true,
		DisableCaller:     true,
		OutputPaths:       []string{"./storage/logs/app-std.log"},
		ErrorOutputPaths:  []string{"./storage/logs/app-err.log"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			EncodeTime:     zapcore.TimeEncoderOfLayout(l.format),
			EncodeDuration: zapcore.StringDurationEncoder,
			LevelKey:       "level",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			NameKey:        "key",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			LineEnding:     zapcore.DefaultLineEnding,
		},
	}.Build()
	return l, errors.Wrap(err, "cannot build logger")
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.e.Debug(msg, fields...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.e.Info(msg, fields...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.e.Error(msg, fields...)
}

// Fatal logs a fatal message and shuts down the application.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.e.Error(msg, fields...)
	l.shutdown <- struct{}{}
}

// Close closes the logger.
func (l *Logger) Close() {
	if err := l.e.Sync(); err != nil && !errors.Is(err, syscall.ENOTTY) {
		l.e.Error("cannot close logger", zap.Error(errors.WithStack(err)))
	} else {
		l.e.Info("logger: closed successfully")
	}
}
