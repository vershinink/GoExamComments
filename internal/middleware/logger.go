package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
)

// loggingResponseWriter - обертка http.ResponseWriter для сохранения
// кода ответа.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewLoggingResponseWriter - конструктор loggingResponseWriter.
func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

// WriteHeader - метод записи кода в ответ.
func (l *loggingResponseWriter) WriteHeader(code int) {
	l.statusCode = code
	l.ResponseWriter.WriteHeader(code)
}

// Logger записывает логи запроса и ответа в логгер slog.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Request log:",
			slog.String("host", r.Host),
			slog.String("uri", r.RequestURI),
			slog.String("protocol", r.Proto),
			slog.String("method", r.Method),
			slog.String("remote_address", r.RemoteAddr),
			slog.String("real_IP", r.Header.Get("X-Real-IP")),
			slog.String("request_id", GetReqID(r.Context())),
		)

		lw := NewLoggingResponseWriter(w)

		next.ServeHTTP(lw, r)

		code := fmt.Sprintf("%d %s", lw.statusCode, http.StatusText(lw.statusCode))

		slog.Info("Response log:",
			slog.String("status", code),
			slog.String("content-type", w.Header().Get("Content-Type")),
		)
	})
}
