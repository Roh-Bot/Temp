package logger

import (
	"context"

	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zapcore.Field)
	Error(ctx context.Context, msg string, fields ...zapcore.Field)
	Warn(ctx context.Context, msg string, fields ...zapcore.Field)
	Debug(ctx context.Context, msg string, fields ...zapcore.Field)
	Flush() error
}
