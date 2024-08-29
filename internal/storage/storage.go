// Пакет содержит основную структуру Comment для работы с комментариями.
package storage

import (
	"errors"
	"time"
)

// Ошибки при работе с БД.
var (
	ErrEmptyDB = errors.New("database is empty")
)

// Comment - структура комментария к посту.
type Comment struct {
	ID       string    `json:"id" bson:"_id"`
	ParentID string    `json:"parentId" bson:"parentId"`
	PostID   string    `json:"postId" bson:"postId"`
	Content  string    `json:"content" bson:"content"`
	PubTime  time.Time `json:"pubTime" bson:"pubTime"`
	Allowed  bool
}

// Interface - интерфейс хранилища комментариев к постам.
//
//go:generate go run github.com/vektra/mockery/v2@v2.44.1 --name=DB
type DB interface {
}
