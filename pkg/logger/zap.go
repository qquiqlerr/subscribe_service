package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger initializes a new logger with the specified log level and format.
func NewLogger(logLevel, format string) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}
	config := zap.Config{
		Encoding:         format,
		Level:            zap.NewAtomicLevelAt(level),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			TimeKey:      "time",
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}
	return config.Build()
}
