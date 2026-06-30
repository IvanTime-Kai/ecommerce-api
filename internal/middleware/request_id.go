package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const RequestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")

		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	id, _ := ctx.Value(RequestIDKey).(string)
	return id
}

func LoggerWithRequestID(ctx context.Context) *slog.Logger {
	return slog.Default().With("request_id", GetRequestID(ctx))
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log := LoggerWithRequestID(r.Context())

		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"ip", r.RemoteAddr,
		)
	})
}
