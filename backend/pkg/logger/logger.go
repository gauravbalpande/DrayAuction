package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(environment string) (*zap.Logger, error) {
	var cfg zap.Config
	if environment == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	return cfg.Build()
}
