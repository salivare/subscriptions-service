package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/salivare-io/slogx"
)

// LogFieldRequestID constant to add request_id field.
const LogFieldRequestID = "request_id"

// LoggerContext contextualizes the log and request_id to attach to all logs.
func LoggerContext(base *slogx.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				reqID := GetRequestID(r.Context())

				logger := base.With(
					slog.String(LogFieldRequestID, reqID),
				)

				ctx := slogx.ToContext(r.Context(), logger)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

// Logger adds a logic to all queries and fills with fields.
func Logger(log *slogx.Logger) func(next http.Handler) http.Handler {
	log = log.With(
		slog.String("component", "http"),
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				entry := log.With(
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr),
					slog.String("user_agent", r.UserAgent()),
					slog.String(LogFieldRequestID, GetRequestID(r.Context())),
				)

				ww := NewResponseWriter(w)

				t1 := time.Now()

				defer func() {
					entry.Info(
						"request completed",
						slog.Int("status", ww.StatusCode()),
						slog.Int("bytes", ww.BytesWritten()),
						slog.Duration("duration", time.Since(t1)),
					)
				}()

				next.ServeHTTP(ww, r)
			},
		)
	}
}
