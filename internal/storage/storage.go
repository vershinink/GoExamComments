// Пакет содержит основную структуру Comment для работы с комментариями.
package storage

import (
	"context"
	"errors"
	"time"
)

// Ошибки при работе с БД.
var (
	ErrNoComments        = errors.New("no comments on provided post id")
	ErrNotAdded          = errors.New("comment was not added")
	ErrIncorrectParentID = errors.New("incorrect parent id")
	ErrIncorrectPostID   = errors.New("incorrect post id")
	ErrEmptyContent      = errors.New("empty comment content field")
)

// Comment - структура комментария к посту.
type Comment struct {
	ID       string    `json:"id" bson:"_id"`
	ParentID string    `json:"parentId" bson:"parentId"`
	PostID   string    `json:"postId" bson:"postId"`
	PubTime  time.Time `json:"pubTime" bson:"pubTime"`
	Allowed  bool      `json:"allowed" bson:"allowed"`
	Content  string    `json:"content" bson:"content"`
	Childs   []Comment `json:"childs" bson:"childs"`
}

// Interface - интерфейс хранилища комментариев к постам.
//
//go:generate go run github.com/vektra/mockery/v2@v2.44.1 --name=DB
type DB interface {
	AddComment(ctx context.Context, com Comment) error
	Comments(ctx context.Context, post string) ([]Comment, error)
}
