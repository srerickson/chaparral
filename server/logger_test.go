package server_test

import (
	"context"
	"io"
	"testing"

	"log/slog"

	"github.com/srerickson/chaparral/server"
)

func TestLoggerFromCtx(t *testing.T) {
	t.Run("with empty context", func(t *testing.T) {
		ctx := context.Background()
		logger := server.LoggerFromCtx(ctx)
		if logger == nil {
			t.Error("LoggerFromCtx() returned nil logger")
		}
	})
}

func TestCtxWithLogger(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := server.CtxWithLogger(context.Background(), logger)
	got := server.LoggerFromCtx(ctx)
	if got != logger {
		t.Error("LoggerFromCtx() didn't return the logger added with CtxWithLogger()")
	}
}
