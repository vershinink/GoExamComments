package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/sqids/sqids-go"
)

// ctxKey - тип ключа для ID запроса внутри контекста.
type ctxKey int

// RequestIDKey - ключ ID запроса внутри контекста.
const RequestIDKey ctxKey = 0

// RequestIDHeader - HTTP заголовок ID запроса.
const RequestIDHeader = "X-Request-Id"

// RequestIDHeader проверяет наличие уникального ID запроса в заголовках
// и записывает значение в контекст. Если не находит, то генерирует новый ID.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		tm := time.Now()
		sec := uint64(tm.Unix())
		nano := uint64(tm.Nanosecond())

		if requestID == "" {
			s, _ := sqids.New(sqids.Options{MinLength: 10})
			id, err := s.Encode([]uint64{sec, nano})
			if err != nil {
				id = "unknown RequestID"
			}
			requestID = id

			r.Header.Set(RequestIDHeader, requestID)
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetReqID возвращает ID запроса из контекста в виде строки.
func GetReqID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
