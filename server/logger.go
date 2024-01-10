package server

import (
	"log/slog"
	"net/http"
	"os"

	"golang.org/x/net/context"
)

type loggerCtxKey struct{}

func CtxWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, logger)
}

func LoggerFromCtx(ctx context.Context) *slog.Logger {
	logger, _ := ctx.Value(loggerCtxKey{}).(*slog.Logger)
	if logger == nil {
		return defaultLogger()
	}
	return logger
}

func defaultLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))
}

// LoggerMiddlware returns middleware that adds the logger to request context.
// The logger can be accessed LoggerFromCtx().
func LoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if logger == nil {
			logger = defaultLogger()
		}
		fn := func(w http.ResponseWriter, r *http.Request) {
			newCtx := CtxWithLogger(r.Context(), logger)
			next.ServeHTTP(w, r.WithContext(newCtx))
		}
		return http.HandlerFunc(fn)
	}
}
