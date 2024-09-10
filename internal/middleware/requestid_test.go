package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func writeReqID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestId := r.Context().Value(RequestIDKey).(string)
		b := []byte(requestId)
		w.Write(b)
	}
}

func TestRequestID(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("GET /", RequestID(writeReqID()))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("RequestID error, status code = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	body = strings.ReplaceAll(body, "\n", "")
	runes := []rune(body)
	if len(runes) < 10 {
		t.Errorf("RequestID error, len = %d, want %d", len(runes), 10)
	}
}

func TestGetReqID(t *testing.T) {
	var id string
	var fn http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		id = GetReqID(r.Context())
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", RequestID(fn))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if id == "" {
		t.Error("GetReqID error, returns empty string")
	}
}
