package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"

	"go.uber.org/zap"
)

var LogEntryCtxKey = &contextKey{"ZapLogger"}

// GetZapLogger retrieves the *zp.Logger from http.Request Context
func GetZapLogger(r *http.Request) *zap.Logger {
	entry, _ := r.Context().Value(LogEntryCtxKey).(*zap.Logger)
	return entry
}

// ZapLogger logger middleware to add the *zap.Logger
func ZapLogger(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			rl := l.With(
				zap.String("proto", r.Proto),
				zap.String("path", r.URL.Path),
				zap.String("reqId", middleware.GetReqID(r.Context())),
			)
			t1 := time.Now()
			defer func() {
				rl.Info("Served",
					zap.Duration("lat", time.Since(t1)),
					zap.Int("status", ww.Status()),
					zap.Int("size", ww.BytesWritten()),
				)
			}()

			next.ServeHTTP(ww, withZapLogger(r, rl))
		})
	}
}

func withZapLogger(r *http.Request, logger *zap.Logger) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), LogEntryCtxKey, logger))
}
