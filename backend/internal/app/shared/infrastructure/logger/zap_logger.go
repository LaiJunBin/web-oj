package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"judge-backend/internal/config"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type ZapLogger struct {
	logger *zap.SugaredLogger
}

func NewZapLogger(cfg *config.Config) sharedtypes.Logger {
	level := zapcore.InfoLevel
	if cfg.Log != nil {
		switch strings.ToLower(cfg.Log.Level) {
		case "debug":
			level = zapcore.DebugLevel
		case "warn":
			level = zapcore.WarnLevel
		case "error":
			level = zapcore.ErrorLevel
		}
	}

	// Always use JSON output for Loki/Promtail compatibility.
	// Write to stdout so Docker's log driver captures the stream.
	// Debug level adds caller info to aid local development.
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(level)
	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stdout"}
	if level == zapcore.DebugLevel {
		zapCfg.Development = true
		zapCfg.DisableCaller = false
	}
	base, err := zapCfg.Build()
	if err != nil {
		panic(err)
	}

	return &ZapLogger{logger: base.Sugar()}
}

func (l *ZapLogger) Debugf(msg string, args ...interface{}) {
	l.logger.Debugf(msg, args...)
}

func (l *ZapLogger) Infof(msg string, args ...interface{}) {
	l.logger.Infof(msg, args...)
}

func (l *ZapLogger) Warnf(msg string, args ...interface{}) {
	l.logger.Warnf(msg, args...)
}

func (l *ZapLogger) Errorf(msg string, args ...interface{}) {
	l.logger.Errorf(msg, args...)
}

func (l *ZapLogger) Fatalf(msg string, args ...interface{}) {
	l.logger.Fatalf(msg, args...)
}

func (l *ZapLogger) Panicf(msg string, args ...interface{}) {
	l.logger.Panicf(msg, args...)
}
