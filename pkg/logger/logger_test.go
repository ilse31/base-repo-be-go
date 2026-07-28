package logger

import (
	"testing"

	"go.uber.org/zap"
)

func TestInit(t *testing.T) {
	err := Init("development")
	if err != nil {
		t.Fatalf("Init(development) failed: %v", err)
	}
	if Log == nil {
		t.Fatal("expected Log to be initialized")
	}

	Info("test info log", zap.String("key", "val"))
	Debug("test debug log")
	Error("test error log")
	Sync()

	err = Init("production")
	if err != nil {
		t.Fatalf("Init(production) failed: %v", err)
	}
	if Log == nil {
		t.Fatal("expected Log to be initialized in production")
	}

	Info("prod info log")
	Sync()
}
