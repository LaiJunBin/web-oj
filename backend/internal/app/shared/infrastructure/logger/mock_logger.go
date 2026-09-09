package logger

import (
	sharedtypes "judge-backend/internal/app/shared/types"
)

type MockLogger struct {
}

func NewMockLogger() sharedtypes.Logger {
	return &MockLogger{}
}

func (l *MockLogger) Debugf(msg string, args ...interface{}) {
}

func (l *MockLogger) Infof(msg string, args ...interface{}) {
}

func (l *MockLogger) Warnf(msg string, args ...interface{}) {
}

func (l *MockLogger) Errorf(msg string, args ...interface{}) {
}

func (l *MockLogger) Fatalf(msg string, args ...interface{}) {
}

func (l *MockLogger) Panicf(msg string, args ...interface{}) {
}
