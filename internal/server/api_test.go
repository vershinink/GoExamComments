// Пакет для работы с сервером и обработчиками API.
package server

import (
	"GoExamComments/internal/logger"
	"GoExamComments/internal/mocks"
	"GoExamComments/internal/storage"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
)

var comment = storage.Comment{ParentID: "", PostID: "news1", Content: "The content of the test comment."}

func TestAddComment(t *testing.T) {
	logger.Discard()

	comm, err := json.Marshal(comment)
	if err != nil {
		t.Fatalf("cannot encode comment, error = %s", err.Error())
	}

	tests := []struct {
		name    string
		header  string
		len     int
		comment []byte
		respErr string
		mockErr error
	}{
		{
			name:    "Comment_OK",
			header:  "Application/json",
			len:     1000,
			comment: comm,
			respErr: "",
			mockErr: nil,
		},
		{
			name:    "Comment_content_type",
			header:  "Other_content_type",
			len:     1000,
			comment: comm,
			respErr: "Content-Type header is not application/json",
			mockErr: nil,
		},
		{
			name:    "Comment_length",
			header:  "Application/json",
			len:     10,
			comment: comm,
			respErr: "the length of the comment must not exceed 1000 characters",
			mockErr: nil,
		},
		{
			name:    "Comment_empty",
			header:  "Application/json",
			len:     1000,
			comment: []byte{},
			respErr: "cannot decode request",
			mockErr: nil,
		},
		{
			name:    "DB_error",
			header:  "Application/json",
			len:     1000,
			comment: comm,
			respErr: "cannot add the comment",
			mockErr: errors.New("DB error"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stMock := mocks.NewDB(t)

			// Исходя из тест-кейса устанавливаем поведение для мока только
			// если планируем дойти до него в тестируемой функции.
			if tt.respErr == "" || tt.mockErr != nil {
				stMock.
					On("AddComment", mock.Anything, mock.AnythingOfType("storage.Comment")).
					Return("comment_id", tt.mockErr).
					Once()
			}

			mux := http.NewServeMux()
			mux.HandleFunc("POST /comments/new", AddComment(tt.len, stMock, nil))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			req := httptest.NewRequest(http.MethodPost, "/comments/new", bytes.NewReader(tt.comment))
			req.Header.Set("Content-Type", tt.header)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			body := rr.Body.String()

			if rr.Code != http.StatusCreated {
				// Проверяем тело ответа и проваливаем тест, если содержимое
				// не совпадает с нашей ожидаемой ошибкой.
				body = strings.ReplaceAll(body, "\n", "")
				if body == tt.respErr {
					t.SkipNow()
				}
				t.Fatalf("AddComment() error = %s, want %s", body, tt.respErr)
			}

			if body != "" {
				t.Errorf("AddComment() error = not empty body")
			}
		})
	}
}

func TestComments(t *testing.T) {
	logger.Discard()

	comm := []storage.Comment{comment}

	tests := []struct {
		name    string
		id      string
		respErr string
		mockErr error
	}{
		{
			name:    "Comments_OK",
			id:      "news1",
			respErr: "",
			mockErr: nil,
		},
		{
			name:    "DB_error",
			id:      "news1",
			respErr: "cannot receive comments",
			mockErr: errors.New("DB error"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stMock := mocks.NewDB(t)

			if tt.respErr == "" || tt.mockErr != nil {
				stMock.
					On("Comments", mock.Anything, mock.AnythingOfType("string")).
					Return(comm, tt.mockErr).
					Once()
			}

			mux := http.NewServeMux()
			mux.HandleFunc("GET /comments/{id}", Comments(stMock))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/comments/%s", tt.id), nil)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			body := rr.Body.String()

			if rr.Code != http.StatusOK {
				// Проверяем тело ответа и проваливаем тест, если содержимое
				// не совпадает с нашей ожидаемой ошибкой.
				body = strings.ReplaceAll(body, "\n", "")
				if body == tt.respErr {
					t.SkipNow()
				}
				t.Fatalf("Comments() error = %s, want %s", body, tt.respErr)
			}

			resp := []storage.Comment{}
			err := json.Unmarshal([]byte(body), &resp)
			if err != nil {
				t.Fatal("Comments() error = cannot unmarshal response")
			}

			// Проверим совпадение контента комментария.
			if comm[0].Content != resp[0].Content {
				t.Errorf("Comments() content = %v, want %v", resp[0].Content, comm[0].Content)
			}

		})
	}
}
