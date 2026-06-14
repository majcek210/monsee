package telemetry

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a Zap logger: JSON in production, human-readable console in dev.
func NewLogger(isProd bool) (*zap.Logger, error) {
	if isProd {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		return cfg.Build()
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return cfg.Build()
}

// Must panics if logger creation fails (safe at startup).
func Must(l *zap.Logger, err error) *zap.Logger {
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	return l
}
