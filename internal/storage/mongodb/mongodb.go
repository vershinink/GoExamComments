// Пакет для работы с базой данных MongoDB в сервисе комментариев.
package mongodb

import (
	"GoExamComments/internal/config"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Название базы и коллекции в БД. Используются переменные,
// а не константы, так как в тестах им присваиваются другие
// значения.
var (
	dbName  string = "goNews"
	colName string = "comments"
)

// tmConn - таймаут на создание пула подключений.
const tmConn time.Duration = time.Second * 20

// Storage - пул подключений к БД.
type Storage struct {
	db *mongo.Client
}

// New - обертка для конструктора пула подключений new.
func New(cfg *config.Config) *Storage {
	//opts := setOpts(cfg.StoragePath, cfg.StorageUser, cfg.StoragePasswd)
	opts := setTestOpts(cfg.StoragePath)
	storage, err := new(opts)
	if err != nil {
		log.Fatalf("failed to init storage: %s", err.Error())
	}
	return storage
}

// setOpts настраивает опции нового подключения к БД.
// Функция вынесена отдельно для подмены ее в пакете
// с тестами.
func setOpts(path, user, password string) *options.ClientOptions {
	credential := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    "admin",
		Username:      user,
		Password:      password,
	}
	opts := options.Client().ApplyURI(path).SetAuth(credential)
	return opts
}

// setTestOpts возвращает опции нового подключения без авторизации.
func setTestOpts(path string) *options.ClientOptions {
	return options.Client().ApplyURI(path)
}

// new - конструктор пула подключений к БД.
func new(opts *options.ClientOptions) (*Storage, error) {
	const operation = "storage.mongodb.new"

	tm, cancel := context.WithTimeout(context.Background(), tmConn)
	defer cancel()

	db, err := mongo.Connect(tm, opts)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	err = db.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	// Создаем индекс по полю postId, чтобы ускорить выдачу всех комментариев
	// по переданному ID поста.
	collection := db.Database(dbName).Collection(colName)
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "postId", Value: -1}},
	}
	_, err = collection.Indexes().CreateOne(tm, indexModel)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &Storage{db: db}, nil
}

// Close - обертка для закрытия пула подключений.
func (s *Storage) Close() error {
	return s.db.Disconnect(context.Background())
}
