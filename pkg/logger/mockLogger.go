package logger

import (
	"context"

	"go.uber.org/zap/zapcore"
)

type MockLogger struct {
}

func (m *MockLogger) Info(ctx context.Context, msg string, fields ...zapcore.Field) {

}
func (m *MockLogger) Error(ctx context.Context, msg string, fields ...zapcore.Field) {

}
func (m *MockLogger) Warn(ctx context.Context, msg string, fields ...zapcore.Field) {

}
func (m *MockLogger) Debug(ctx context.Context, msg string, fields ...zapcore.Field) {

}
func (m *MockLogger) Flush() error {
	return nil
}
