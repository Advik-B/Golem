package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func init() {
	// Default to a development logger. Can be replaced by production config.
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Pretty colors
	var err error
	Logger, err = config.Build()
	if err != nil {
		panic("failed to initialize zap logger: " + err.Error())
	}
}

// ReplaceGlobals replaces the global zap logger with our configured one.
func ReplaceGlobals() {
	zap.ReplaceGlobals(Logger)
}
