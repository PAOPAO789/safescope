package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(environment, level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	if environment == "development" {
		cfg = zap.NewDevelopmentConfig()
	}
	var parsed zapcore.Level
	if err := parsed.UnmarshalText([]byte(strings.ToLower(level))); err != nil {
		return nil, err
	}
	cfg.Level = zap.NewAtomicLevelAt(parsed)
	return cfg.Build()
}
